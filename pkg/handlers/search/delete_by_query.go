package search

import (
	stderrors "errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

// DeleteByQuery searches the index and deletes all matches
//
// @Id DeleteByQuery
// @Summary Searches the index and deletes all matched documents
// @security BasicAuth
// @Tags    Search
// @Accept  json
// @Produce json
// @Param   index  path  string  true  "Index"
// @Param   query  body  meta.ZincQueryForSDK true  "Query"
// @Success 200 {object} meta.HTTPResponseDeleteByQuery
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/{index}/_delete_by_query [post]
func DeleteByQuery(c *gin.Context) {
	start := time.Now()
	query := &meta.ZincQuery{}
	if err := zutils.GinBindJSON(c, query); err != nil {
		log.Printf("handlers.search.searchDSL: %s", err.Error())
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	ignoreUnavailable, _ := strconv.ParseBool(c.Query("ignore_unavailable"))
	indexNames, err := core.ResolveTargetIndexes(c.Param("target"), ignoreUnavailable)
	if err != nil {
		var notFound *core.IndexNotFoundError
		if stderrors.As(err, &notFound) {
			renderIndexNotFound(c, notFound.Index)
			return
		}
		errors.HandleError(c, err)
		return
	}

	// Deletes go through the WAL, so each batch is refreshed before searching
	// again. That also gives every request the effect of refresh=true.
	// conflicts is accepted but ignored: there are no version conflicts here.
	attempted := make(map[string]struct{})
	failures := []string{}
	deleted := 0
	batches := 0
	for len(indexNames) > 0 {
		query.From = 0
		query.Size = deleteByQueryBatchSize
		resp, err := core.MultiSearch(indexNames, query)
		if err != nil {
			errors.HandleError(c, err)
			return
		}

		// A document whose delete failed keeps matching; stop once a batch
		// turns up nothing that hasn't been tried already.
		fresh := 0
		for _, hit := range resp.Hits.Hits {
			key := hit.Index + "/" + hit.ID
			if _, ok := attempted[key]; ok {
				continue
			}
			attempted[key] = struct{}{}
			fresh++

			index, _ := core.GetIndex(hit.Index)
			if err := index.DeleteDocument(hit.ID); err != nil {
				failures = append(failures, hit.ID)
				continue
			}
			deleted++
		}
		if fresh == 0 {
			break
		}
		batches++

		for _, name := range indexNames {
			if index, ok := core.GetIndex(name); ok {
				if err := index.Refresh(); err != nil {
					errors.HandleError(c, err)
					return
				}
			}
		}
	}

	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponseDeleteByQuery{
		Took:             time.Since(start).Milliseconds(),
		TimedOut:         false,
		Total:            len(attempted),
		Deleted:          deleted,
		Batches:          batches,
		VersionConflicts: 0,
		Noops:            0,
		Failures:         failures,
		Retries: meta.HttpRetriesResponse{
			Bulk:   0,
			Search: 0,
		},
		ThrottledMillis:      0,
		RequestsPerSecond:    -1,
		ThrottledUntilMillis: 0,
	})
}

const deleteByQueryBatchSize = 1000

// renderIndexNotFound writes the ES 404 for a target that names no index or alias
func renderIndexNotFound(c *gin.Context, index string) {
	cause := gin.H{
		"type":   "index_not_found_exception",
		"reason": "no such index [" + index + "]",
		"index":  index,
	}
	zutils.GinRenderJSON(c, http.StatusNotFound, gin.H{
		"error": gin.H{
			"root_cause": []gin.H{cause},
			"type":       cause["type"],
			"reason":     cause["reason"],
			"index":      index,
		},
		"status": http.StatusNotFound,
	})
}
