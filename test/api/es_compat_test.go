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

package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiESPutAlias(t *testing.T) {
	index := "es-compat-put-alias"
	alias := index + "-alias"
	require.Equal(t, http.StatusOK, request("PUT", "/es/"+index, strings.NewReader(`{}`)).Code)
	defer request("DELETE", "/es/"+index, nil)

	for _, path := range []string{"/es/" + index + "/_alias/" + alias, "/es/" + index + "/_aliases/" + alias} {
		t.Run("PUT "+path, func(t *testing.T) {
			resp := request("PUT", path, nil)
			assert.Equal(t, http.StatusOK, resp.Code)
			assert.JSONEq(t, `{"acknowledged":true}`, resp.Body.String())

			resp = request("GET", "/es/_alias/"+alias, nil)
			assert.Equal(t, http.StatusOK, resp.Code)
			assert.Contains(t, resp.Body.String(), index)

			resp = request("DELETE", path, nil)
			assert.Equal(t, http.StatusOK, resp.Code)
		})
	}
}
