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

	"github.com/blugelabs/bluge/analysis"
)

type SimplePatternSplitTokenizer struct {
	pattern *regexp.Regexp
}

func NewSimplePatternSplitTokenizer(pattern *regexp.Regexp) *SimplePatternSplitTokenizer {
	return &SimplePatternSplitTokenizer{pattern: pattern}
}

func (t *SimplePatternSplitTokenizer) Tokenize(input []byte) analysis.TokenStream {
	tokens := make(analysis.TokenStream, 0)
	matches := t.pattern.FindAllIndex(input, -1)

	prevEnd := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		if start > prevEnd {
			tokens = append(tokens, &analysis.Token{
				Start:        prevEnd,
				End:          start,
				Term:         input[prevEnd:start],
				PositionIncr: 1,
				Type:         analysis.AlphaNumeric,
			})
		}
		prevEnd = end
	}

	if prevEnd < len(input) {
		tokens = append(tokens, &analysis.Token{
			Start:        prevEnd,
			End:          len(input),
			Term:         input[prevEnd:],
			PositionIncr: 1,
			Type:         analysis.AlphaNumeric,
		})
	}

	return tokens
}
