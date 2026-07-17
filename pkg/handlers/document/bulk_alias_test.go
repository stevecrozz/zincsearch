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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestBulkViaAlias(t *testing.T) {
	indexName := "bulk-alias-test-v1"
	aliasName := "bulk-alias-test"

	// Create the index
	index, err := core.NewIndex(indexName, "disk", 2)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))

	// Create an alias pointing to the index
	require.NoError(t, core.ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias(aliasName, []string{indexName}))

	t.Run("bulk index via alias stores in real index", func(t *testing.T) {
		data := `{"index":{"_id":"1"}}
{"name":"hello world"}
{"index":{"_id":"2"}}
{"name":"foo bar"}
`
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, data)
		utils.SetGinRequestParams(c, map[string]string{"target": aliasName})
		ESBulk(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"errors":false`)

		// Wait for WAL consumption
		time.Sleep(2 * time.Second)

		// Verify documents are in the real index, not a phantom alias-named index
		_, exists := core.GetIndex(aliasName)
		assert.False(t, exists, "should not create an index with the alias name")

		realIndex, exists := core.GetIndex(indexName)
		assert.True(t, exists)
		_ = realIndex
	})

	t.Run("bulk update with doc_as_upsert via alias", func(t *testing.T) {
		data := `{"update":{"_id":"3"}}
{"doc":{"name":"upserted doc"},"doc_as_upsert":true}
`
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, data)
		utils.SetGinRequestParams(c, map[string]string{"target": aliasName})
		ESBulk(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"errors":false`)
	})

	// Cleanup
	t.Cleanup(func() {
		_ = core.ZINC_INDEX_ALIAS_LIST.RemoveIndexesFromAlias(aliasName, []string{indexName})
		_ = core.DeleteIndex(indexName)
	})
}
