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

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

type ClusterSettings struct {
	Persistent map[string]interface{} `json:"persistent"`
	Transient  map[string]interface{} `json:"transient"`
}

// GetClusterSettings returns empty cluster settings, Zinc has none to report
func GetClusterSettings(c *gin.Context) {
	zutils.GinRenderJSON(c, http.StatusOK, ClusterSettings{
		Persistent: map[string]interface{}{},
		Transient:  map[string]interface{}{},
	})
}

// PutClusterSettings acknowledges cluster settings without applying them.
// ES clients set things like logger levels here; Zinc has no equivalent.
func PutClusterSettings(c *gin.Context) {
	var settings ClusterSettings
	if err := zutils.GinBindJSON(c, &settings); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}
	if settings.Persistent == nil {
		settings.Persistent = map[string]interface{}{}
	}
	if settings.Transient == nil {
		settings.Transient = map[string]interface{}{}
	}

	zutils.GinRenderJSON(c, http.StatusOK, gin.H{
		"acknowledged": true,
		"persistent":   settings.Persistent,
		"transient":    settings.Transient,
	})
}
