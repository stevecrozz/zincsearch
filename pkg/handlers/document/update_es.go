/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package document

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/core"
	zincerrors "github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

// ESUpdate applies an ES partial update: the body's doc is merged into the
// document, with doc_as_upsert or upsert used when it doesn't exist.
// Scripted updates aren't supported.
//
// @Id ESUpdateDocument
// @Summary ES partial update of a document
// @security BasicAuth
// @Tags    Document
// @Accept  json
// @Produce json
// @Param   index  path  string  true  "Index"
// @Param   id     path  string  true  "ID"
// @Param   body   body  map[string]interface{}  true  "{doc, upsert, doc_as_upsert}"
// @Success 200 {object} meta.HTTPResponseESID
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /es/{index}/_update/{id} [post]
func ESUpdate(c *gin.Context) {
	indexName := c.Param("target")
	docID := c.Param("id")

	var body map[string]interface{}
	if err := zutils.GinBindJSON(c, &body); err != nil {
		renderESError(c, http.StatusBadRequest, zincerrors.ErrorTypeXContentParseException, err.Error(), "")
		return
	}
	if _, ok := body["script"]; ok {
		renderESError(c, http.StatusBadRequest, zincerrors.ErrorTypeIllegalArgumentException, "scripted updates are not supported", "")
		return
	}
	partial, ok := body["doc"].(map[string]interface{})
	if !ok {
		renderESError(c, http.StatusBadRequest, "action_request_validation_exception", "Validation Failed: 1: script or doc is missing;", "")
		return
	}
	upsert, _ := body["upsert"].(map[string]interface{})
	if asUpsert, _ := body["doc_as_upsert"].(bool); asUpsert {
		upsert = partial
	}

	if indexes, ok := core.ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias(indexName); ok && len(indexes) > 0 {
		indexName = indexes[0]
	}
	index, _, err := core.GetOrCreateIndex(indexName, "", 0)
	if err != nil {
		renderESError(c, http.StatusInternalServerError, zincerrors.ErrorTypeRuntimeException, err.Error(), indexName)
		return
	}

	result, err := index.MergeDocument(docID, partial, upsert)
	if err == zincerrors.ErrorIDNotFound {
		renderESError(c, http.StatusNotFound, "document_missing_exception", "["+docID+"]: document missing", indexName)
		return
	}
	if err != nil {
		renderESError(c, http.StatusInternalServerError, zincerrors.ErrorTypeRuntimeException, err.Error(), indexName)
		return
	}

	status := http.StatusOK
	if result == core.MergeResultCreated {
		status = http.StatusCreated
	}
	zutils.GinRenderJSON(c, status, meta.HTTPResponseESID{
		Message: "ok",
		ID:      docID,
		ESID:    docID,
		Index:   indexName,
		Version: 1,
		Result:  result,
	})
}

// renderESError writes an error in ES's response shape.
func renderESError(c *gin.Context, status int, errType, reason, index string) {
	cause := gin.H{"type": errType, "reason": reason}
	if index != "" {
		cause["index"] = index
	}
	errBody := gin.H{"root_cause": []gin.H{cause}}
	for k, v := range cause {
		errBody[k] = v
	}
	zutils.GinRenderJSON(c, status, gin.H{"error": errBody, "status": status})
}
