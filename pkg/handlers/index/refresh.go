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

package index

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

// @Id Refresh
// @Summary Resfresh index
// @security BasicAuth
// @Tags    Index
// @Produce json
// @Param   index  path  string  true  "Index"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /api/index/{index}/refresh [post]
func Refresh(c *gin.Context) {
	if refreshTargets(c) {
		c.JSON(http.StatusOK, meta.HTTPResponse{Message: "ok"})
	}
}

// FlushES refreshes the target indexes and responds in the ES _flush format
func FlushES(c *gin.Context) {
	if refreshTargets(c) {
		zutils.GinRenderJSON(c, http.StatusOK, gin.H{"_shards": gin.H{"total": 1, "successful": 1, "failed": 0}})
	}
}

// refreshTargets makes pending writes searchable on every index the target
// resolves to. On failure it writes the error response and returns false.
func refreshTargets(c *gin.Context) bool {
	indexNames, err := resolveTargetIndexes(c.Param("target"))
	if err != nil {
		c.JSON(http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return false
	}

	for _, name := range indexNames {
		idx, exists := core.GetIndex(name)
		if !exists {
			continue
		}
		if err := idx.Refresh(); err != nil {
			c.JSON(http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
			return false
		}
	}
	return true
}

// resolveTargetIndexes expands a comma-separated list of index names, aliases
// and wildcard patterns into the distinct index names it refers to.
// A name without a wildcard that matches nothing is an error.
func resolveTargetIndexes(target string) ([]string, error) {
	var indexNames []string
	seen := make(map[string]struct{})
	add := func(name string) {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			indexNames = append(indexNames, name)
		}
	}

	for _, name := range strings.Split(target, ",") {
		if strings.Contains(name, "*") {
			for _, index := range core.ZINC_INDEX_LIST.List() {
				if indexNameMatches(name, index.GetName()) {
					add(index.GetName())
				}
			}
			continue
		}
		if _, exists := core.GetIndex(name); exists {
			add(name)
			continue
		}
		if aliased, ok := core.ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias(name); ok && len(aliased) > 0 {
			for _, n := range aliased {
				add(n)
			}
			continue
		}
		return nil, errors.New("index " + name + " does not exists")
	}

	return indexNames, nil
}
