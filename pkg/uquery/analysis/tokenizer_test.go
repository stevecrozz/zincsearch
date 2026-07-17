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

	"github.com/stretchr/testify/assert"
)

func TestRequestTokenizerSingle_SimplePatternSplit(t *testing.T) {
	tests := []struct {
		name    string
		typ     string
		options interface{}
		input   string
		want    []string
		wantErr bool
	}{
		{
			name: "simple_pattern_split with default pattern",
			typ:  "simple_pattern_split",
			options: map[string]interface{}{
				"type": "simple_pattern_split",
			},
			input: "hello-world foo",
			want:  []string{"hello", "world", "foo"},
		},
		{
			name: "simple_pattern_split with custom pattern",
			typ:  "simple_pattern_split",
			options: map[string]interface{}{
				"type":    "simple_pattern_split",
				"pattern": `\.`,
			},
			input: "www.example.com",
			want:  []string{"www", "example", "com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := RequestTokenizerSingle(tt.typ, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			tokens := tok.Tokenize([]byte(tt.input))
			got := make([]string, 0, len(tokens))
			for _, token := range tokens {
				got = append(got, string(token.Term))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequestTokenizer_SimplePatternSplit(t *testing.T) {
	data := map[string]interface{}{
		"my_splitter": map[string]interface{}{
			"type":    "simple_pattern_split",
			"pattern": `_`,
		},
	}

	tokenizers, err := RequestTokenizer(data)
	assert.NoError(t, err)
	assert.Contains(t, tokenizers, "my_splitter")

	tokens := tokenizers["my_splitter"].Tokenize([]byte("foo_bar_baz"))
	got := make([]string, 0, len(tokens))
	for _, token := range tokens {
		got = append(got, string(token.Term))
	}
	assert.Equal(t, []string{"foo", "bar", "baz"}, got)
}
