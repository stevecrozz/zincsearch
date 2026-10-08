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

	"github.com/blugelabs/bluge"
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

func TestTermsQuery_Lookup(t *testing.T) {
	docs := map[string]map[string]interface{}{
		"links/2": {
			"ids":    []interface{}{float64(10), float64(11)},
			"nested": map[string]interface{}{"tags": []interface{}{map[string]interface{}{"name": "a"}, map[string]interface{}{"name": "b"}}},
			"single": "x",
		},
	}
	orig := TermsLookupFunc
	TermsLookupFunc = func(index, id string) (map[string]interface{}, error) {
		return docs[index+"/"+id], nil
	}
	t.Cleanup(func() { TermsLookupFunc = orig })

	lookup := func(id, path string) map[string]interface{} {
		return map[string]interface{}{"_id": map[string]interface{}{"index": "links", "id": id, "path": path}}
	}

	t.Run("array field builds a terms query", func(t *testing.T) {
		q, err := TermsQuery(lookup("2", "ids"), nil)
		assert.NoError(t, err)
		assert.IsType(t, bluge.NewBooleanQuery(), q)
	})

	t.Run("missing document matches none", func(t *testing.T) {
		q, err := TermsQuery(lookup("nope", "ids"), nil)
		assert.NoError(t, err)
		assert.IsType(t, bluge.NewMatchNoneQuery(), q)
	})

	t.Run("missing field matches none", func(t *testing.T) {
		q, err := TermsQuery(lookup("2", "missing"), nil)
		assert.NoError(t, err)
		assert.IsType(t, bluge.NewMatchNoneQuery(), q)
	})

	t.Run("lookup without path is a parse error", func(t *testing.T) {
		_, err := TermsQuery(map[string]interface{}{"_id": map[string]interface{}{"index": "links", "id": "2"}}, nil)
		assert.Error(t, err)
	})
}

func TestExtractPath(t *testing.T) {
	source := map[string]interface{}{
		"ids":    []interface{}{float64(10), []interface{}{float64(11)}},
		"single": "x",
		"nested": map[string]interface{}{"tags": []interface{}{map[string]interface{}{"name": "a"}, map[string]interface{}{"name": "b"}}},
		"a.b":    "dotted",
		"obj":    map[string]interface{}{"k": "v"},
	}
	assert.Equal(t, []interface{}{float64(10), float64(11)}, extractPath(source, "ids"))
	assert.Equal(t, []interface{}{"x"}, extractPath(source, "single"))
	assert.Equal(t, []interface{}{"a", "b"}, extractPath(source, "nested.tags.name"))
	assert.Equal(t, []interface{}{"dotted"}, extractPath(source, "a.b"))
	assert.Nil(t, extractPath(source, "obj"))
	assert.Nil(t, extractPath(source, "single.x"))
	assert.Nil(t, extractPath(source, "missing"))
}
