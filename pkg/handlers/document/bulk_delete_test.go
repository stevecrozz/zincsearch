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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestBulkDeleteNumericID(t *testing.T) {
	indexName := "bulk-delete-numeric-id"
	index, err := core.NewIndex(indexName, "disk", 1)
	require.NoError(t, err)
	require.NoError(t, core.StoreIndex(index))
	t.Cleanup(func() { _ = core.DeleteIndex(indexName) })

	bulk := func(data string) (int, string) {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, data)
		ESBulk(c)
		return w.Code, w.Body.String()
	}

	code, body := bulk(`{"index":{"_index":"` + indexName + `","_id":10}}
{"name":"a"}
`)
	require.Equal(t, http.StatusOK, code)
	require.Contains(t, body, `"_id":"10"`)
	time.Sleep(2 * time.Second) // WAL consumption

	t.Run("numeric id of an existing doc is deleted", func(t *testing.T) {
		code, body := bulk(`{"delete":{"_index":"` + indexName + `","_id":10}}` + "\n")
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"errors":false`)
		assert.Contains(t, body, `"_id":"10"`)
		assert.Contains(t, body, `"result":"deleted"`)
	})

	t.Run("numeric id of a missing doc is a per-item 404", func(t *testing.T) {
		code, body := bulk(`{"delete":{"_index":"` + indexName + `","_id":999}}` + "\n")
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"errors":false`)
		assert.Contains(t, body, `"result":"not_found"`)
		assert.Contains(t, body, `"status":404`)
	})

	t.Run("non-scalar id is a per-item 400", func(t *testing.T) {
		code, body := bulk(`{"delete":{"_index":"` + indexName + `","_id":{"x":1}}}` + "\n")
		assert.Equal(t, http.StatusOK, code)
		assert.Contains(t, body, `"errors":true`)
		assert.True(t, strings.Contains(body, `"status":400`), body)
	})

	t.Run("delete without _index uses the path target", func(t *testing.T) {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, `{"delete":{"_id":5}}`+"\n")
		utils.SetGinRequestParams(c, map[string]string{"target": indexName})
		ESBulk(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"_index":"`+indexName+`"`)
	})
}
