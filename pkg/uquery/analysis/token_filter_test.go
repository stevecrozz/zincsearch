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

package analysis

import (
	"testing"

	"github.com/blugelabs/bluge/analysis"
	"github.com/stretchr/testify/assert"
)

func TestRequestTokenFilterSingle_ASCIIFolding(t *testing.T) {
	f, err := RequestTokenFilterSingle("asciifolding", nil)
	assert.NoError(t, err)

	input := analysis.TokenStream{
		{Term: []byte("café"), Start: 0, End: 5, PositionIncr: 1, Type: analysis.AlphaNumeric},
		{Term: []byte("naïve"), Start: 6, End: 12, PositionIncr: 1, Type: analysis.AlphaNumeric},
	}

	got := f.Filter(input)
	terms := make([]string, 0, len(got))
	for _, token := range got {
		terms = append(terms, string(token.Term))
	}
	assert.Equal(t, []string{"cafe", "naive"}, terms)
}

func TestRequestTokenFilterSingle_CommonGrams(t *testing.T) {
	options := map[string]interface{}{
		"type":         "common_grams",
		"common_words": []interface{}{"the"},
	}

	f, err := RequestTokenFilterSingle("common_grams", options)
	assert.NoError(t, err)

	input := analysis.TokenStream{
		{Term: []byte("the"), Start: 0, End: 3, PositionIncr: 1, Type: analysis.AlphaNumeric},
		{Term: []byte("cat"), Start: 4, End: 7, PositionIncr: 1, Type: analysis.AlphaNumeric},
		{Term: []byte("sat"), Start: 8, End: 11, PositionIncr: 1, Type: analysis.AlphaNumeric},
	}

	got := f.Filter(input)
	terms := make([]string, 0, len(got))
	for _, token := range got {
		terms = append(terms, string(token.Term))
	}
	assert.Equal(t, []string{"the", "the_cat", "cat", "sat"}, terms)
}
