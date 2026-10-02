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

package elastic

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/zutils/json"
	"github.com/zincsearch/zincsearch/test/utils"
)

func TestPutClusterSettings(t *testing.T) {
	t.Run("echoes settings back acknowledged", func(t *testing.T) {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, `{"transient":{"logger._root":"DEBUG"}}`)
		PutClusterSettings(c)
		require.Equal(t, http.StatusOK, w.Code)

		resp := map[string]interface{}{}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, true, resp["acknowledged"])
		assert.Equal(t, map[string]interface{}{"logger._root": "DEBUG"}, resp["transient"])
		assert.Equal(t, map[string]interface{}{}, resp["persistent"])
	})

	t.Run("invalid body", func(t *testing.T) {
		c, w := utils.NewGinContext()
		utils.SetGinRequestData(c, `not json`)
		PutClusterSettings(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetClusterSettings(t *testing.T) {
	c, w := utils.NewGinContext()
	GetClusterSettings(c)
	require.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"persistent":{},"transient":{}}`, w.Body.String())
}
