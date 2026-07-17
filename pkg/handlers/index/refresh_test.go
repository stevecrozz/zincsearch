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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestRefresh(t *testing.T) {
	type args struct {
		code   int
		params map[string]string
		result string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				code:   http.StatusOK,
				params: map[string]string{"target": "TestRefresh.index_1"},
				result: "ok",
			},
			wantErr: false,
		},
		{
			name: "empty",
			args: args{
				code:   http.StatusBadRequest,
				params: map[string]string{"target": ""},
				result: "does not exists",
			},
			wantErr: false,
		},
	}

	t.Run("prepare", func(t *testing.T) {
		index, err := core.NewIndex("TestRefresh.index_1", "disk", 2)
		assert.NoError(t, err)
		assert.NotNil(t, index)

		err = core.StoreIndex(index)
		assert.NoError(t, err)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := utils.NewGinContext()
			utils.SetGinRequestParams(c, tt.args.params)
			Refresh(c)
			assert.Equal(t, tt.args.code, w.Code)
			assert.Contains(t, w.Body.String(), tt.args.result)

			resp := make(map[string]string)
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
		})
	}

	t.Run("refresh via alias with single index", func(t *testing.T) {
		require.NoError(t, core.ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias("TestRefresh.alias_1", []string{"TestRefresh.index_1"}))

		c, w := utils.NewGinContext()
		utils.SetGinRequestParams(c, map[string]string{"target": "TestRefresh.alias_1"})
		Refresh(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")

		_ = core.ZINC_INDEX_ALIAS_LIST.RemoveIndexesFromAlias("TestRefresh.alias_1", []string{"TestRefresh.index_1"})
	})

	t.Run("refresh via alias with multiple indexes", func(t *testing.T) {
		idx2, err := core.NewIndex("TestRefresh.index_2", "disk", 2)
		require.NoError(t, err)
		require.NoError(t, core.StoreIndex(idx2))

		require.NoError(t, core.ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias("TestRefresh.alias_multi", []string{"TestRefresh.index_1", "TestRefresh.index_2"}))

		c, w := utils.NewGinContext()
		utils.SetGinRequestParams(c, map[string]string{"target": "TestRefresh.alias_multi"})
		Refresh(c)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ok")

		_ = core.ZINC_INDEX_ALIAS_LIST.RemoveIndexesFromAlias("TestRefresh.alias_multi", []string{"TestRefresh.index_1", "TestRefresh.index_2"})
		_ = core.DeleteIndex("TestRefresh.index_2")
	})

	t.Run("cleanup", func(t *testing.T) {
		_ = core.DeleteIndex("TestRefresh.index_1")
	})
}
