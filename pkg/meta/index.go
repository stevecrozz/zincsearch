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
	stdjson "encoding/json"

	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

type Index struct {
	ShardNum    int64                  `json:"shard_num"`
	Name        string                 `json:"name"`
	StorageType string                 `json:"storage_type"`
	Settings    *IndexSettings         `json:"settings,omitempty"`
	Mappings    *Mappings              `json:"mappings,omitempty"`
	Shards      map[string]*IndexShard `json:"shards"`
	Stats       IndexStat              `json:"stats"`
	Version     string                 `json:"version"`
}

type IndexShard struct {
	ShardNum int64               `json:"shard_num"`
	ID       string              `json:"id"`
	NodeID   string              `json:"node_id"` // remote instance ID
	Shards   []*IndexSecondShard `json:"shards"`
	Stats    IndexStat           `json:"stats"`
}

type IndexSecondShard struct {
	ID    int64     `json:"id"`
	Stats IndexStat `json:"stats"`
}

type IndexStat struct {
	DocTimeMin  int64  `json:"doc_time_min"`
	DocTimeMax  int64  `json:"doc_time_max"`
	DocNum      uint64 `json:"doc_num"`
	StorageSize uint64 `json:"storage_size"`
	WALSize     uint64 `json:"wal_size"`
}

type IndexSimple struct {
	Name        string                 `json:"name"`
	StorageType string                 `json:"storage_type"`
	ShardNum    int64                  `json:"shard_num"`
	Settings    *IndexSettings         `json:"settings,omitempty"`
	Mappings    map[string]interface{} `json:"mappings,omitempty"`
}

type IndexSettings struct {
	NumberOfShards   int64          `json:"number_of_shards,omitempty"`
	NumberOfReplicas int64          `json:"number_of_replicas,omitempty"`
	Analysis         *IndexAnalysis `json:"analysis,omitempty"`
}

// UnmarshalJSON handles ES-compatible formats:
// - Flat: {"number_of_shards": 1, "analysis": {...}}
// - Nested: {"index": {"number_of_shards": 1, "analysis": {...}}}
// - String numbers: {"number_of_shards": "1"} (ES accepts both string and int)
func (s *IndexSettings) UnmarshalJSON(data []byte) error {
	var raw struct {
		NumberOfShards   interface{}    `json:"number_of_shards,omitempty"`
		NumberOfReplicas interface{}    `json:"number_of_replicas,omitempty"`
		Analysis         *IndexAnalysis `json:"analysis,omitempty"`
		Index            *stdjson.RawMessage `json:"index,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.Index != nil {
		var nested struct {
			NumberOfShards   interface{}    `json:"number_of_shards,omitempty"`
			NumberOfReplicas interface{}    `json:"number_of_replicas,omitempty"`
			Analysis         *IndexAnalysis `json:"analysis,omitempty"`
		}
		if err := json.Unmarshal(*raw.Index, &nested); err == nil {
			raw.NumberOfShards = nested.NumberOfShards
			raw.NumberOfReplicas = nested.NumberOfReplicas
			raw.Analysis = nested.Analysis
		}
	}

	s.NumberOfShards = toInt64(raw.NumberOfShards)
	s.NumberOfReplicas = toInt64(raw.NumberOfReplicas)
	s.Analysis = raw.Analysis
	return nil
}

func toInt64(v interface{}) int64 {
	switch v := v.(type) {
	case float64:
		return int64(v)
	case string:
		var n int64
		for _, c := range v {
			if c >= '0' && c <= '9' {
				n = n*10 + int64(c-'0')
			}
		}
		return n
	default:
		return 0
	}
}

type IndexAnalysis struct {
	Analyzer    map[string]*Analyzer   `json:"analyzer,omitempty"`
	CharFilter  map[string]interface{} `json:"char_filter,omitempty"`
	Tokenizer   map[string]interface{} `json:"tokenizer,omitempty"`
	TokenFilter map[string]interface{} `json:"token_filter,omitempty"`
	Filter      map[string]interface{} `json:"filter,omitempty"` // compatibility with es, alias for TokenFilter
}
