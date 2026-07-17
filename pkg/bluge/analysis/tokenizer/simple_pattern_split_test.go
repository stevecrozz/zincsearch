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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimplePatternSplitTokenizer(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		want    []string
	}{
		{
			name:    "split on whitespace",
			pattern: `\s+`,
			input:   "hello world foo",
			want:    []string{"hello", "world", "foo"},
		},
		{
			name:    "split on non-word characters",
			pattern: `\W+`,
			input:   "hello-world_foo.bar",
			want:    []string{"hello", "world_foo", "bar"},
		},
		{
			name:    "split on comma",
			pattern: `,`,
			input:   "one,two,three",
			want:    []string{"one", "two", "three"},
		},
		{
			name:    "empty input",
			pattern: `\s+`,
			input:   "",
			want:    []string{},
		},
		{
			name:    "no matches produces single token",
			pattern: `\d+`,
			input:   "hello",
			want:    []string{"hello"},
		},
		{
			name:    "delimiter at start",
			pattern: `\s+`,
			input:   " hello world",
			want:    []string{"hello", "world"},
		},
		{
			name:    "delimiter at end",
			pattern: `\s+`,
			input:   "hello world ",
			want:    []string{"hello", "world"},
		},
		{
			name:    "consecutive delimiters",
			pattern: `-`,
			input:   "a--b",
			want:    []string{"a", "b"},
		},
		{
			name:    "split email on special chars",
			pattern: `[@.]`,
			input:   "user@example.com",
			want:    []string{"user", "example", "com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := regexp.MustCompile(tt.pattern)
			tok := NewSimplePatternSplitTokenizer(r)
			tokens := tok.Tokenize([]byte(tt.input))

			got := make([]string, 0, len(tokens))
			for _, token := range tokens {
				got = append(got, string(token.Term))
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSimplePatternSplitTokenizerPositions(t *testing.T) {
	r := regexp.MustCompile(`\s+`)
	tok := NewSimplePatternSplitTokenizer(r)
	tokens := tok.Tokenize([]byte("hello world foo"))

	assert.Len(t, tokens, 3)

	assert.Equal(t, 0, tokens[0].Start)
	assert.Equal(t, 5, tokens[0].End)
	assert.Equal(t, 1, tokens[0].PositionIncr)

	assert.Equal(t, 6, tokens[1].Start)
	assert.Equal(t, 11, tokens[1].End)
	assert.Equal(t, 1, tokens[1].PositionIncr)

	assert.Equal(t, 12, tokens[2].Start)
	assert.Equal(t, 15, tokens[2].End)
	assert.Equal(t, 1, tokens[2].PositionIncr)
}
