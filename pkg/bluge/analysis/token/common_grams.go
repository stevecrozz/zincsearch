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
	"github.com/blugelabs/bluge/analysis"
)

type CommonGramsFilter struct {
	commonWords map[string]struct{}
}

func NewCommonGramsFilter(commonWords []string) *CommonGramsFilter {
	m := make(map[string]struct{}, len(commonWords))
	for _, w := range commonWords {
		m[w] = struct{}{}
	}
	return &CommonGramsFilter{commonWords: m}
}

func (f *CommonGramsFilter) isCommon(term string) bool {
	_, ok := f.commonWords[term]
	return ok
}

func (f *CommonGramsFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	if len(input) < 2 {
		return input
	}

	rv := make(analysis.TokenStream, 0, len(input)*2)

	for i, token := range input {
		rv = append(rv, token)

		if i < len(input)-1 {
			next := input[i+1]
			if f.isCommon(string(token.Term)) || f.isCommon(string(next.Term)) {
				bigram := make([]byte, 0, len(token.Term)+1+len(next.Term))
				bigram = append(bigram, token.Term...)
				bigram = append(bigram, '_')
				bigram = append(bigram, next.Term...)
				rv = append(rv, &analysis.Token{
					Term:         bigram,
					Start:        token.Start,
					End:          next.End,
					PositionIncr: 0,
					Type:         analysis.Shingle,
				})
			}
		}
	}

	return rv
}
