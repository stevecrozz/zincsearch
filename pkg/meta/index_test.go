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

package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

func TestIndexSettings_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantAnalysis bool
		wantShards   int64
	}{
		{
			name:         "flat format",
			input:        `{"number_of_shards": 3, "analysis": {"analyzer": {"my_analyzer": {"type": "custom", "tokenizer": "standard"}}}}`,
			wantAnalysis: true,
			wantShards:   3,
		},
		{
			name:         "nested under index key",
			input:        `{"index": {"number_of_shards": 5, "analysis": {"analyzer": {"my_analyzer": {"type": "custom", "tokenizer": "standard"}}}}}`,
			wantAnalysis: true,
			wantShards:   5,
		},
		{
			name:         "empty settings",
			input:        `{}`,
			wantAnalysis: false,
			wantShards:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s IndexSettings
			err := json.Unmarshal([]byte(tt.input), &s)
			assert.NoError(t, err)
			if tt.wantAnalysis {
				assert.NotNil(t, s.Analysis)
				assert.Contains(t, s.Analysis.Analyzer, "my_analyzer")
			} else {
				assert.Nil(t, s.Analysis)
			}
			assert.Equal(t, tt.wantShards, s.NumberOfShards)
		})
	}
}
