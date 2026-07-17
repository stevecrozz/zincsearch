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
	"fmt"
	"net/http"
	"path"
	"sort"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

func CatIndices(c *gin.Context) {
	pattern := c.Param("target")

	items := core.ZINC_INDEX_LIST.List()
	sort.Slice(items, func(i, j int) bool {
		return items[i].GetName() < items[j].GetName()
	})

	type catEntry struct {
		Health       string `json:"health"`
		Status       string `json:"status"`
		Index        string `json:"index"`
		UUID         string `json:"uuid"`
		Pri          string `json:"pri"`
		Rep          string `json:"rep"`
		DocsCount    string `json:"docs.count"`
		DocsDeleted  string `json:"docs.deleted"`
		StoreSize    string `json:"store.size"`
		PriStoreSize string `json:"pri.store.size"`
	}

	result := make([]catEntry, 0)
	for _, idx := range items {
		name := idx.GetName()
		if pattern != "" {
			matched, _ := path.Match(pattern, name)
			if !matched {
				continue
			}
		}
		stats := idx.GetIndex().Stats
		result = append(result, catEntry{
			Health:       "green",
			Status:       "open",
			Index:        name,
			UUID:         name,
			Pri:          "1",
			Rep:          "0",
			DocsCount:    fmt.Sprintf("%d", stats.DocNum),
			DocsDeleted:  "0",
			StoreSize:    fmt.Sprintf("%db", stats.StorageSize),
			PriStoreSize: fmt.Sprintf("%db", stats.StorageSize),
		})
	}

	zutils.GinRenderJSON(c, http.StatusOK, result)
}
