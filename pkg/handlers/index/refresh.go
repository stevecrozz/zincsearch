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
	"net/http"

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
	indexNames, err := core.ResolveTargetIndexes(c.Param("target"), false)
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
