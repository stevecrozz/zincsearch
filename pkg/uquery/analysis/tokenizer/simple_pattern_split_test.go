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

package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSimplePatternSplitTokenizer(t *testing.T) {
	tests := []struct {
		name    string
		options interface{}
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:    "default pattern splits on non-word",
			options: nil,
			input:   "hello-world foo",
			want:    []string{"hello", "world", "foo"},
		},
		{
			name: "custom pattern",
			options: map[string]interface{}{
				"type":    "simple_pattern_split",
				"pattern": ",",
			},
			input: "one,two,three",
			want:  []string{"one", "two", "three"},
		},
		{
			name: "empty pattern uses default",
			options: map[string]interface{}{
				"type":    "simple_pattern_split",
				"pattern": "",
			},
			input: "hello world",
			want:  []string{"hello", "world"},
		},
		{
			name: "invalid regex returns error",
			options: map[string]interface{}{
				"type":    "simple_pattern_split",
				"pattern": "[invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := NewSimplePatternSplitTokenizer(tt.options)
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
