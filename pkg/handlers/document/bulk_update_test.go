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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestBulkUpdateMergesPartialDoc(t *testing.T) {
	indexName := "bulk-update-merge"
	index, err := core.NewIndex(indexName, "disk", 1)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))
	t.Cleanup(func() { _ = core.DeleteIndex(indexName) })

	bulk := func(data string) string {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, data)
		utils.SetGinRequestParams(c, map[string]string{"target": indexName})
		ESBulk(c)
		require.Equal(t, http.StatusOK, w.Code)
		return w.Body.String()
	}
	source := func(id string) map[string]interface{} {
		require.NoError(t, index.Refresh())
		hit, err := index.GetDocument(id)
		require.NoError(t, err)
		return hit.Source.(map[string]interface{})
	}

	t.Run("back-to-back partial updates keep each other's fields", func(t *testing.T) {
		// separate requests with no wait: the first is still in the WAL
		body := bulk(`{"update":{"_id":2}}
{"doc":{"text":["yay"],"meta":{"a":1}},"doc_as_upsert":true}
`)
		assert.Contains(t, body, `"result":"created"`)
		body = bulk(`{"update":{"_id":2}}
{"doc":{"name":"notyay","meta":{"b":2}},"doc_as_upsert":true}
`)
		assert.Contains(t, body, `"result":"updated"`)

		src := source("2")
		assert.Equal(t, []interface{}{"yay"}, src["text"])
		assert.Equal(t, "notyay", src["name"])
		assert.Equal(t, map[string]interface{}{"a": float64(1), "b": float64(2)}, src["meta"])
	})

	t.Run("upsert doc is used when the document is missing", func(t *testing.T) {
		bulk(`{"update":{"_id":"3"}}
{"doc":{"name":"x"},"upsert":{"name":"fresh"}}
`)
		assert.Equal(t, "fresh", source("3")["name"])
	})

	t.Run("partial update of a missing doc without upsert is a per-item 404", func(t *testing.T) {
		body := bulk(`{"update":{"_id":"404"}}
{"doc":{"name":"x"}}
`)
		assert.Contains(t, body, `"errors":true`)
		assert.Contains(t, body, `"status":404`)
	})
}

func TestBulkItemKeys(t *testing.T) {
	indexName := "bulk-item-keys"
	index, err := core.NewIndex(indexName, "disk", 1)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))
	t.Cleanup(func() { _ = core.DeleteIndex(indexName) })

	c, w := utils.NewGinContext()
	utils.SetGinRequestData(c, `{"index":{"_id":"1"}}
{"a":1}
{"create":{"_id":"2"}}
{"a":2}
{"update":{"_id":"1"}}
{"doc":{"a":1}}
{"delete":{"_id":"2"}}
`)
	utils.SetGinRequestParams(c, map[string]string{"target": indexName})
	ESBulk(c)
	require.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Items []map[string]BulkResponseItem `json:"items"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Items, 4)
	keys := []string{}
	for _, item := range resp.Items {
		for k := range item {
			keys = append(keys, k)
		}
	}
	assert.Equal(t, []string{"index", "create", "update", "delete"}, keys)
	assert.Equal(t, "noop", resp.Items[2]["update"].Result)
}
