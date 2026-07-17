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
	"strings"
	"testing"

	"github.com/blugelabs/bluge/analysis"
	"github.com/stretchr/testify/assert"
)

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

func collectTerms(tokens analysis.TokenStream) []string {
	terms := make([]string, 0, len(tokens))
	for _, t := range tokens {
		terms = append(terms, string(t.Term))
	}
	return terms
}

func TestCommonGramsFilter(t *testing.T) {
	tests := []struct {
		name        string
		commonWords []string
		input       []string
		want        []string
	}{
		{
			name:        "bigram when first token is common",
			commonWords: []string{"the"},
			input:       []string{"the", "cat", "sat"},
			want:        []string{"the", "the_cat", "cat", "sat"},
		},
		{
			name:        "bigram when second token is common",
			commonWords: []string{"is"},
			input:       []string{"sky", "is", "blue"},
			want:        []string{"sky", "sky_is", "is", "is_blue", "blue"},
		},
		{
			name:        "bigram when both tokens are common",
			commonWords: []string{"the", "is"},
			input:       []string{"the", "is"},
			want:        []string{"the", "the_is", "is"},
		},
		{
			name:        "no bigrams when no common words adjacent",
			commonWords: []string{"the"},
			input:       []string{"cat", "sat", "down"},
			want:        []string{"cat", "sat", "down"},
		},
		{
			name:        "single token unchanged",
			commonWords: []string{"the"},
			input:       []string{"hello"},
			want:        []string{"hello"},
		},
		{
			name:        "empty input",
			commonWords: []string{"the"},
			input:       []string{},
			want:        []string{},
		},
		{
			name:        "multiple common words",
			commonWords: []string{"the", "a", "is"},
			input:       []string{"the", "cat", "is", "a", "pet"},
			want:        []string{"the", "the_cat", "cat", "cat_is", "is", "is_a", "a", "a_pet", "pet"},
		},
		{
			name:        "empty common words list",
			commonWords: []string{},
			input:       []string{"hello", "world"},
			want:        []string{"hello", "world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewCommonGramsFilter(tt.commonWords)
			input := makeTokenStream(tt.input)
			got := f.Filter(input)
			assert.Equal(t, tt.want, collectTerms(got))
		})
	}
}

func TestCommonGramsFilterPositionIncr(t *testing.T) {
	f := NewCommonGramsFilter([]string{"the"})
	input := makeTokenStream([]string{"the", "cat"})
	got := f.Filter(input)

	assert.Len(t, got, 3)
	// "the" - original token
	assert.Equal(t, "the", string(got[0].Term))
	assert.Equal(t, 1, got[0].PositionIncr)
	// "the_cat" - bigram at same position
	assert.Equal(t, "the_cat", string(got[1].Term))
	assert.Equal(t, 0, got[1].PositionIncr)
	// "cat" - next position
	assert.Equal(t, "cat", string(got[2].Term))
	assert.Equal(t, 1, got[2].PositionIncr)
}

func TestCommonGramsFilterByteOffsets(t *testing.T) {
	f := NewCommonGramsFilter([]string{"the"})
	// "the cat" -> tokens at [0,3] and [4,7]
	input := analysis.TokenStream{
		{Term: []byte("the"), Start: 0, End: 3, PositionIncr: 1, Type: analysis.AlphaNumeric},
		{Term: []byte("cat"), Start: 4, End: 7, PositionIncr: 1, Type: analysis.AlphaNumeric},
	}
	got := f.Filter(input)

	assert.Len(t, got, 3)
	// bigram spans from start of first to end of second
	assert.Equal(t, 0, got[1].Start)
	assert.Equal(t, 7, got[1].End)
}

func TestCommonGramsFilterSeparator(t *testing.T) {
	f := NewCommonGramsFilter([]string{"the"})
	input := makeTokenStream([]string{"the", "quick"})
	got := f.Filter(input)

	// Verify the separator is underscore
	bigram := string(got[1].Term)
	assert.True(t, strings.Contains(bigram, "_"))
	assert.Equal(t, "the_quick", bigram)
}
