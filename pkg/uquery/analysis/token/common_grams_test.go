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

package token

import (
	"testing"

	"github.com/blugelabs/bluge/analysis"
	"github.com/stretchr/testify/assert"
)

func TestNewCommonGramsTokenFilter(t *testing.T) {
	tests := []struct {
		name    string
		options interface{}
		input   []string
		want    []string
	}{
		{
			name: "common_words option",
			options: map[string]interface{}{
				"type":         "common_grams",
				"common_words": []interface{}{"the", "is"},
			},
			input: []string{"the", "cat", "is", "here"},
			want:  []string{"the", "the_cat", "cat", "cat_is", "is", "is_here", "here"},
		},
		{
			name: "stopwords option as fallback",
			options: map[string]interface{}{
				"type":      "common_grams",
				"stopwords": []interface{}{"a"},
			},
			input: []string{"what", "a", "day"},
			want:  []string{"what", "what_a", "a", "a_day", "day"},
		},
		{
			name:    "nil options produces no bigrams",
			options: nil,
			input:   []string{"hello", "world"},
			want:    []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewCommonGramsTokenFilter(tt.options)
			assert.NoError(t, err)

			input := makeTokenStream(tt.input)
			got := f.Filter(input)

			terms := make([]string, 0, len(got))
			for _, token := range got {
				terms = append(terms, string(token.Term))
			}
			assert.Equal(t, tt.want, terms)
		})
	}
}

func makeTokenStream(terms []string) analysis.TokenStream {
	tokens := make(analysis.TokenStream, 0, len(terms))
	pos := 0
	for _, term := range terms {
		tokens = append(tokens, &analysis.Token{
			Term:         []byte(term),
			Start:        pos,
			End:          pos + len(term),
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		})
		pos += len(term) + 1
	}
	return tokens
}
