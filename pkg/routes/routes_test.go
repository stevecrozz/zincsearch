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

package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetESRoutesAtRoot(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/", func(c *gin.Context) { c.Status(http.StatusFound) })
	setESRoutes(r, "")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("HEAD", "/", nil))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Elasticsearch", w.Header().Get("X-elastic-product"))

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	assert.Equal(t, http.StatusFound, w.Code)
}
