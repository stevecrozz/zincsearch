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

	"github.com/stretchr/testify/assert"
)

func TestASCIIFoldingTokenFilter(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "folds accented characters",
			input: []string{"café", "résumé", "naïve"},
			want:  []string{"cafe", "resume", "naive"},
		},
		{
			name:  "preserves plain ASCII",
			input: []string{"hello", "world"},
			want:  []string{"hello", "world"},
		},
		{
			name:  "folds various unicode",
			input: []string{"über", "straße", "año"},
			want:  []string{"uber", "strasse", "ano"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewASCIIFoldingTokenFilter()
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
