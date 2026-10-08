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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

func TestIndex_SearchHighlightFromSource(t *testing.T) {
	index, err := NewIndex("TestIndex_SearchHighlightFromSource", "disk", 1)
	require.NoError(t, err)
	require.NoError(t, StoreIndex(index))
	t.Cleanup(func() { _ = DeleteIndex(index.GetName()) })

	require.NoError(t, index.CreateDocument("1", map[string]interface{}{
		"name": "notyay",
		"text": []interface{}{"the word yay is in here"},
	}, false))
	require.NoError(t, index.Refresh())

	// as elaine sends it: highlight a field that _source excludes
	var query meta.ZincQuery
	require.NoError(t, json.Unmarshal([]byte(`{
		"query":{"bool":{"should":[{"match":{"text":{"query":"yay"}}}]}},
		"highlight":{"pre_tags":[""],"post_tags":[""],
			"fields":{"text":{"fragment_size":150,"number_of_fragments":1}}},
		"_source":{"excludes":"text"},
		"size":10}`), &query))

	resp, err := index.Search(&query)
	require.NoError(t, err)
	require.Len(t, resp.Hits.Hits, 1)
	hit := resp.Hits.Hits[0]
	assert.Equal(t, []string{"the word yay is in here"}, hit.Highlight["text"])
}
