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
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestESUpdate(t *testing.T) {
	indexName := "es-update"
	aliasName := "es-update-alias"
	index, err := core.NewIndex(indexName, "disk", 1)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))
	require.NoError(t, core.ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias(aliasName, []string{indexName}))
	t.Cleanup(func() {
		_ = core.ZINC_INDEX_ALIAS_LIST.RemoveIndexesFromAlias(aliasName, []string{indexName})
		_ = core.DeleteIndex(indexName)
	})

	update := func(target, id, body string) (int, string) {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, body)
		utils.SetGinRequestParams(c, map[string]string{"target": target, "id": id})
		ESUpdate(c)
		return w.Code, w.Body.String()
	}
	source := func(id string) map[string]interface{} {
		require.NoError(t, index.Refresh())
		hit, err := index.GetDocument(id)
		require.NoError(t, err)
		return hit.Source.(map[string]interface{})
	}

	t.Run("doc_as_upsert creates, then a partial doc merges", func(t *testing.T) {
		code, body := update(indexName, "1", `{"doc":{"a":1,"o":{"x":1}},"doc_as_upsert":true}`)
		assert.Equal(t, http.StatusCreated, code)
		assert.Contains(t, body, `"result":"created"`)

		code, body = update(aliasName, "1", `{"doc":{"b":2,"o":{"y":2}}}`)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"result":"updated"`)
		assert.Contains(t, body, `"_index":"`+indexName+`"`)

		src := source("1")
		assert.Equal(t, float64(1), src["a"])
		assert.Equal(t, float64(2), src["b"])
		assert.Equal(t, map[string]interface{}{"x": float64(1), "y": float64(2)}, src["o"])
	})

	t.Run("unchanged doc is a noop", func(t *testing.T) {
		code, body := update(indexName, "1", `{"doc":{"a":1}}`)
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"result":"noop"`)
	})

	t.Run("missing doc without upsert is document_missing_exception", func(t *testing.T) {
		code, body := update(indexName, "nope", `{"doc":{"a":1}}`)
		assert.Equal(t, http.StatusNotFound, code)
		assert.Contains(t, body, `"type":"document_missing_exception"`)
	})

	t.Run("body without doc is a validation error", func(t *testing.T) {
		code, body := update(indexName, "1", `{"a":1}`)
		assert.Equal(t, http.StatusBadRequest, code)
		assert.Contains(t, body, `action_request_validation_exception`)
	})

	t.Run("scripts are rejected", func(t *testing.T) {
		code, _ := update(indexName, "1", `{"script":{"source":"ctx._source.a++"}}`)
		assert.Equal(t, http.StatusBadRequest, code)
	})
}
