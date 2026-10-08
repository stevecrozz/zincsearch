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
	"fmt"
	"strconv"
	"strings"

	"github.com/blugelabs/bluge"

	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/meta"
)

func TermsQuery(query map[string]interface{}, mappings *meta.Mappings) (bluge.Query, error) {
	if len(query) > 2 {
		return nil, errors.New(errors.ErrorTypeParsingException, "[terms] query doesn't support multiple fields")
	}

	field := ""
	values := []string{}
	valueFloat := []float64{}
	valueInts := []int{}
	valueBools := []bool{}
	boost := -1.0
	appendTerm := func(v interface{}) error {
		switch v := v.(type) {
		case string:
			values = append(values, v)
		case float64:
			valueFloat = append(valueFloat, v)
		case int:
			valueInts = append(valueInts, v)
		case bool:
			valueBools = append(valueBools, v)
		default:
			return errors.New(errors.ErrorTypeXContentParseException, fmt.Sprintf("[term] doesn't support values of type: %T", v))
		}
		return nil
	}
	for k, v := range query {
		if strings.ToLower(k) == "boost" {
			boost = v.(float64)
			continue
		}

		field = k
		switch v := v.(type) {
		case []string:
			values = v
		case []float64:
			valueFloat = v
		case []int:
			valueInts = v
		case []bool:
			valueBools = v
		case []interface{}:
			for _, vv := range v {
				if err := appendTerm(vv); err != nil {
					return nil, err
				}
			}
		case map[string]interface{}:
			terms, err := termsLookup(v)
			if err != nil {
				return nil, err
			}
			if len(terms) == 0 {
				return bluge.NewMatchNoneQuery(), nil
			}
			for _, vv := range terms {
				if err := appendTerm(vv); err != nil {
					return nil, err
				}
			}
		default:
			return nil, errors.New(errors.ErrorTypeXContentParseException, fmt.Sprintf("[terms] doesn't support values of type: %T", v))
		}
	}

	// _id is always a string field; coerce numeric values
	if field == "_id" {
		for _, term := range valueFloat {
			if term == float64(int64(term)) {
				values = append(values, fmt.Sprintf("%d", int64(term)))
			} else {
				values = append(values, fmt.Sprintf("%v", term))
			}
		}
		for _, term := range valueInts {
			values = append(values, fmt.Sprintf("%d", term))
		}
		valueFloat = nil
		valueInts = nil
	}

	subq := bluge.NewBooleanQuery()
	for _, term := range values {
		subqq, err := TermQueryText(field, &meta.TermQuery{Value: term})
		if err != nil {
			return nil, err
		}
		subq.AddShould(subqq)
	}
	for _, term := range valueFloat {
		subqq, err := TermQueryNumeric(field, &meta.TermQuery{Value: term})
		if err != nil {
			return nil, err
		}
		subq.AddShould(subqq)
	}
	for _, term := range valueInts {
		subqq, err := TermQueryNumeric(field, &meta.TermQuery{Value: term})
		if err != nil {
			return nil, err
		}
		subq.AddShould(subqq)
	}
	for _, term := range valueBools {
		subqq, err := TermQueryBool(field, &meta.TermQuery{Value: term})
		if err != nil {
			return nil, err
		}
		subq.AddShould(subqq)
	}
	if boost >= 0 {
		subq.SetBoost(boost)
	}

	return subq, nil
}

// TermsLookupFunc returns the _source of document id in index (an index or
// alias), or nil if neither exists. core sets it, since this package can't
// import core.
var TermsLookupFunc func(index, id string) (map[string]interface{}, error)

// termsLookup resolves a terms lookup object ({"index", "id", "path"}) to the
// values found at path in the looked-up document. A missing document or
// field yields no values.
func termsLookup(lookup map[string]interface{}) ([]interface{}, error) {
	index, _ := lookup["index"].(string)
	path, _ := lookup["path"].(string)
	var id string
	switch v := lookup["id"].(type) {
	case string:
		id = v
	case float64:
		id = strconv.FormatFloat(v, 'f', -1, 64)
	}
	if index == "" || id == "" || path == "" {
		return nil, errors.New(errors.ErrorTypeParsingException, "[terms] lookup requires [index], [id] and [path]")
	}
	if TermsLookupFunc == nil {
		return nil, errors.New(errors.ErrorTypeParsingException, "[terms] lookup is not available")
	}

	source, err := TermsLookupFunc(index, id)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, nil
	}
	return extractPath(source, path), nil
}

// extractPath returns the scalar values at a dotted path, flattening arrays
// along the way. Keys that themselves contain dots are matched too.
func extractPath(v interface{}, path string) []interface{} {
	switch v := v.(type) {
	case []interface{}:
		var out []interface{}
		for _, vv := range v {
			out = append(out, extractPath(vv, path)...)
		}
		return out
	case map[string]interface{}:
		if path == "" {
			return nil
		}
		var out []interface{}
		if vv, ok := v[path]; ok {
			out = append(out, extractPath(vv, "")...)
		}
		for i := strings.IndexByte(path, '.'); i >= 0; i = nextDot(path, i) {
			if vv, ok := v[path[:i]]; ok {
				out = append(out, extractPath(vv, path[i+1:])...)
			}
		}
		return out
	case nil:
		return nil
	default:
		if path != "" {
			return nil
		}
		return []interface{}{v}
	}
}

func nextDot(path string, i int) int {
	j := strings.IndexByte(path[i+1:], '.')
	if j < 0 {
		return -1
	}
	return i + 1 + j
}
