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

package highlight

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

func parse(t *testing.T, s string) map[string]interface{} {
	var v map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(s), &v))
	return v
}

func TestSource(t *testing.T) {
	// the query elaine's bundle search sends
	query := parse(t, `{"bool":{
		"should":[
			{"multi_match":{"query":"yay","fields":["name"],"type":"most_fields"}},
			{"match":{"text":{"query":"yay","boost":0.4}}},
			{"match_phrase":{"text":{"query":"yay","analyzer":"ds_common_search_analyzer"}}}
		],
		"must_not":[{"match":{"text":"never"}}],
		"filter":[{"terms":{"_id":["1"]}}]}}`)
	global := &meta.Highlight{PreTags: []string{""}, PostTags: []string{""}}
	options := &meta.Highlight{FragmentSize: 150, NumberOfFragments: 1}

	t.Run("matches in an array field", func(t *testing.T) {
		doc := map[string]interface{}{"text": []interface{}{"nothing here", "Yay, it works"}}
		assert.Equal(t, []string{"Yay, it works"}, Source(query, "text", doc, global, options, nil, nil))
	})

	t.Run("tags wrap the match", func(t *testing.T) {
		doc := map[string]interface{}{"name": "say yay twice yay"}
		got := Source(query, "name", doc, nil, &meta.Highlight{NumberOfFragments: 3}, nil, nil)
		assert.Equal(t, []string{"say <em>yay</em> twice <em>yay</em>"}, got)
	})

	t.Run("no match returns nil", func(t *testing.T) {
		doc := map[string]interface{}{"text": "nothing here"}
		assert.Nil(t, Source(query, "text", doc, global, options, nil, nil))
	})

	t.Run("must_not terms are not highlighted", func(t *testing.T) {
		doc := map[string]interface{}{"text": "never"}
		assert.Nil(t, Source(query, "text", doc, global, options, nil, nil))
	})

	t.Run("field not in the query returns nil", func(t *testing.T) {
		doc := map[string]interface{}{"other": "yay"}
		assert.Nil(t, Source(query, "other", doc, global, options, nil, nil))
	})
}

func TestFragments(t *testing.T) {
	text := strings.Repeat("lorem ipsum ", 20) + "needle " + strings.Repeat("dolor sit ", 20)
	start := strings.Index(text, "needle")
	got := fragments(text, []span{{start, start + 6}}, 40, 1, "[", "]")
	assert.Len(t, got, 1)
	assert.Contains(t, got[0], "[needle]")
	assert.LessOrEqual(t, len(got[0]), 40+2+10)
	// fragments start and end on word boundaries
	assert.False(t, strings.HasPrefix(got[0], "orem") || strings.HasPrefix(got[0], "psum"), got[0])
	for _, w := range strings.Fields(got[0]) {
		assert.Contains(t, []string{"lorem", "ipsum", "dolor", "sit", "[needle]"}, w)
	}

	t.Run("limit 0 returns the whole text", func(t *testing.T) {
		assert.Equal(t, []string{"a [b] c"}, fragments("a b c", []span{{2, 3}}, 1, 0, "[", "]"))
	})

	t.Run("separate fragments for distant matches", func(t *testing.T) {
		text := "aa " + strings.Repeat("x ", 50) + "bb"
		got := fragments(text, []span{{0, 2}, {len(text) - 2, len(text)}}, 10, 5, "[", "]")
		assert.Len(t, got, 2)
		assert.True(t, strings.HasPrefix(got[0], "[aa]"), got[0])
		assert.True(t, strings.HasSuffix(got[1], "[bb]"), got[1])
	})

	t.Run("multibyte text is not split mid-rune", func(t *testing.T) {
		text := strings.Repeat("ñandú ", 30) + "yay " + strings.Repeat("über ", 30)
		start := strings.Index(text, "yay")
		got := fragments(text, []span{{start, start + 3}}, 25, 1, "[", "]")
		for _, w := range strings.Fields(got[0]) {
			assert.Contains(t, []string{"ñandú", "über", "[yay]"}, w)
		}
	})
}
