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

package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTermsQuery_IdCoercion(t *testing.T) {
	t.Run("integer values on _id are coerced to strings", func(t *testing.T) {
		query := map[string]interface{}{
			"_id": []interface{}{float64(1), float64(23), float64(456)},
		}
		q, err := TermsQuery(query, nil)
		assert.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("string values on _id work normally", func(t *testing.T) {
		query := map[string]interface{}{
			"_id": []interface{}{"1", "23", "456"},
		}
		q, err := TermsQuery(query, nil)
		assert.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("mixed numeric and string on _id", func(t *testing.T) {
		query := map[string]interface{}{
			"_id": []interface{}{float64(1), "23"},
		}
		q, err := TermsQuery(query, nil)
		assert.NoError(t, err)
		assert.NotNil(t, q)
	})

	t.Run("numeric values on regular field stay numeric", func(t *testing.T) {
		query := map[string]interface{}{
			"age": []interface{}{float64(25), float64(30)},
		}
		q, err := TermsQuery(query, nil)
		assert.NoError(t, err)
		assert.NotNil(t, q)
	})
}
