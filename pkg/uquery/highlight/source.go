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
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/blugelabs/bluge/analysis"
	"github.com/blugelabs/bluge/analysis/analyzer"

	"github.com/zincsearch/zincsearch/pkg/meta"
	zincanalysis "github.com/zincsearch/zincsearch/pkg/uquery/analysis"
	"github.com/zincsearch/zincsearch/pkg/zutils"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

const (
	defaultFragmentSize = 100
	defaultPreTag       = "<em>"
	defaultPostTag      = "</em>"
)

// Source highlights a field from the document's _source, the way ES's
// default highlighter does for fields without stored term positions: the
// query's terms for the field and the field's value are analyzed with the
// field's analyzer, and fragments are cut around the matching tokens.
// query must come from QueryMap. Returns nil when nothing matches.
func Source(query interface{}, field string, source map[string]interface{}, global, options *meta.Highlight,
	mappings *meta.Mappings, analyzers map[string]*analysis.Analyzer) []string {
	zer := fieldAnalyzer(field, mappings, analyzers)

	terms := make(map[string]struct{})
	texts, exacts := queryTerms(query, field)
	for _, text := range texts {
		for _, token := range zer.Analyze([]byte(text)) {
			terms[string(token.Term)] = struct{}{}
		}
	}
	for _, term := range exacts {
		terms[term] = struct{}{}
	}
	if len(terms) == 0 {
		return nil
	}

	pre, post := tags(global, options)
	size, num := options.FragmentSize, options.NumberOfFragments
	var out []string
	for _, value := range sourceValues(source, field) {
		var spans []span
		for _, token := range zer.Analyze([]byte(value)) {
			if _, ok := terms[string(token.Term)]; !ok {
				continue
			}
			if token.Start < 0 || token.End > len(value) || token.Start >= token.End {
				continue // offsets shifted by a char filter
			}
			spans = append(spans, span{token.Start, token.End})
		}
		if len(spans) == 0 {
			continue
		}
		limit := 0
		if num > 0 {
			limit = num - len(out)
		}
		out = append(out, fragments(value, spans, size, limit, pre, post)...)
		if num > 0 && len(out) >= num {
			break
		}
	}
	return out
}

func fieldAnalyzer(field string, mappings *meta.Mappings, analyzers map[string]*analysis.Analyzer) *analysis.Analyzer {
	indexZer, searchZer := zincanalysis.QueryAnalyzerForField(analyzers, mappings, field)
	if indexZer != nil {
		return indexZer
	}
	if searchZer != nil {
		return searchZer
	}
	if mappings != nil {
		if prop, ok := mappings.GetProperty(field); ok && prop.Type != "text" {
			return analyzer.NewKeywordAnalyzer() // match whole values
		}
	}
	return analyzer.NewStandardAnalyzer() // what bluge indexes text fields with by default
}

func tags(global, options *meta.Highlight) (string, string) {
	pre, post := defaultPreTag, defaultPostTag
	if global != nil && len(global.PreTags) > 0 && len(global.PostTags) > 0 {
		pre, post = global.PreTags[0], global.PostTags[0]
	}
	if len(options.PreTags) > 0 && len(options.PostTags) > 0 {
		pre, post = options.PreTags[0], options.PostTags[0]
	}
	return pre, post
}

// QueryMap returns query as decoded JSON, which Source walks. Requests
// from HTTP already are; typed queries are round-tripped through JSON.
func QueryMap(query interface{}) interface{} {
	switch query.(type) {
	case map[string]interface{}, []interface{}, nil:
		return query
	}
	data, err := json.Marshal(query)
	if err != nil {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil
	}
	return v
}

// queryTerms walks a query DSL tree and returns, for field, the query
// texts that need analyzing (match family) and the exact terms (term,
// terms). must_not clauses are skipped, as in ES.
func queryTerms(query interface{}, field string) (texts, exacts []string) {
	var walk func(q interface{})
	walk = func(q interface{}) {
		switch q := q.(type) {
		case []interface{}:
			for _, v := range q {
				walk(v)
			}
		case map[string]interface{}:
			for k, v := range q {
				switch strings.ToLower(k) {
				case "bool":
					if b, ok := v.(map[string]interface{}); ok {
						walk(b["must"])
						walk(b["should"])
						walk(b["filter"])
					}
				case "constant_score":
					if b, ok := v.(map[string]interface{}); ok {
						walk(b["filter"])
					}
				case "boosting":
					if b, ok := v.(map[string]interface{}); ok {
						walk(b["positive"])
					}
				case "dis_max":
					if b, ok := v.(map[string]interface{}); ok {
						walk(b["queries"])
					}
				case "match", "match_phrase", "match_phrase_prefix", "match_bool_prefix":
					if text, ok := fieldQuery(v, field, "query"); ok {
						texts = append(texts, text)
					}
				case "multi_match":
					if b, ok := v.(map[string]interface{}); ok && multiMatchHasField(b["fields"], field) {
						if text, err := zutils.ToString(b["query"]); err == nil {
							texts = append(texts, text)
						}
					}
				case "term":
					if term, ok := fieldQuery(v, field, "value"); ok {
						exacts = append(exacts, term)
					}
				case "terms":
					if b, ok := v.(map[string]interface{}); ok {
						if values, ok := b[field].([]interface{}); ok {
							for _, value := range values {
								if term, err := zutils.ToString(value); err == nil {
									exacts = append(exacts, term)
								}
							}
						}
					}
				}
			}
		}
	}
	walk(query)
	return texts, exacts
}

// fieldQuery reads {field: "text"} or {field: {key: "text"}}.
func fieldQuery(v interface{}, field, key string) (string, bool) {
	b, ok := v.(map[string]interface{})
	if !ok {
		return "", false
	}
	switch fv := b[field].(type) {
	case nil:
		return "", false
	case map[string]interface{}:
		text, err := zutils.ToString(fv[key])
		return text, err == nil && text != ""
	default:
		text, err := zutils.ToString(fv)
		return text, err == nil && text != ""
	}
}

func multiMatchHasField(fields interface{}, field string) bool {
	list, ok := fields.([]interface{})
	if !ok {
		return false
	}
	for _, f := range list {
		name, _ := f.(string)
		if i := strings.IndexByte(name, '^'); i >= 0 {
			name = name[:i]
		}
		if name == field {
			return true
		}
	}
	return false
}

// sourceValues returns the string values at a dotted field path, one per
// array element.
func sourceValues(source map[string]interface{}, field string) []string {
	var out []string
	var walk func(v interface{}, path string)
	walk = func(v interface{}, path string) {
		switch v := v.(type) {
		case []interface{}:
			for _, vv := range v {
				walk(vv, path)
			}
		case map[string]interface{}:
			if path == "" {
				return
			}
			if vv, ok := v[path]; ok {
				walk(vv, "")
			}
			for i := 0; i < len(path); i++ {
				if path[i] == '.' {
					if vv, ok := v[path[:i]]; ok {
						walk(vv, path[i+1:])
					}
				}
			}
		case nil:
		default:
			if path == "" {
				if s, err := zutils.ToString(v); err == nil {
					out = append(out, s)
				}
			}
		}
	}
	walk(source, field)
	return out
}

type span struct{ start, end int }

// fragments cuts text into up to limit fragments of about size bytes
// around the matching spans, wrapping each match in pre and post. A limit
// of 0 returns the whole text as one fragment, as ES does for
// number_of_fragments 0.
func fragments(text string, spans []span, size, limit int, pre, post string) []string {
	sort.Slice(spans, func(i, j int) bool { return spans[i].start < spans[j].start })
	merged := spans[:1]
	for _, s := range spans[1:] {
		last := &merged[len(merged)-1]
		if s.start <= last.end {
			if s.end > last.end {
				last.end = s.end
			}
			continue
		}
		merged = append(merged, s)
	}

	if limit == 0 {
		return []string{tag(text, 0, len(text), merged, pre, post)}
	}
	if size <= 0 {
		size = defaultFragmentSize
	}

	var out []string
	fragEnd := -1
	for _, s := range merged {
		if s.start < fragEnd {
			continue // already in the previous fragment
		}
		start := s.start - (size-(s.end-s.start))/2
		if start < 0 {
			start = 0
		}
		end := start + size
		if end > len(text) {
			end = len(text)
			if start = end - size; start < 0 {
				start = 0
			}
		}
		if end < s.end {
			end = s.end
		}
		start, end = wordStart(text, start, s.start), wordEnd(text, end, s.end)
		out = append(out, strings.TrimSpace(tag(text, start, end, merged, pre, post)))
		fragEnd = end
		if len(out) == limit {
			break
		}
	}
	return out
}

// tag returns text[start:end] with every span inside it wrapped.
func tag(text string, start, end int, spans []span, pre, post string) string {
	var b strings.Builder
	pos := start
	for _, s := range spans {
		if s.start < start || s.end > end {
			continue
		}
		b.WriteString(text[pos:s.start])
		b.WriteString(pre)
		b.WriteString(text[s.start:s.end])
		b.WriteString(post)
		pos = s.end
	}
	b.WriteString(text[pos:end])
	return b.String()
}

// wordStart moves i forward to the start of a word, but not past limit.
func wordStart(text string, i, limit int) int {
	if i == 0 || isSpaceBefore(text, i) {
		return i
	}
	for j := i; j < limit; {
		r, n := utf8.DecodeRuneInString(text[j:])
		j += n
		if unicode.IsSpace(r) {
			return j
		}
	}
	for i > 0 && !utf8.RuneStart(text[i]) {
		i--
	}
	return i
}

// wordEnd moves i back to the end of a word, but not before floor; if that
// isn't possible it moves forward to the end of the current word instead.
func wordEnd(text string, i, floor int) int {
	if i >= len(text) {
		return len(text)
	}
	if r, _ := utf8.DecodeRuneInString(text[i:]); unicode.IsSpace(r) {
		return i
	}
	for j := i; j > floor; {
		r, n := utf8.DecodeLastRuneInString(text[:j])
		if unicode.IsSpace(r) {
			return j - n
		}
		j -= n
	}
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		if unicode.IsSpace(r) {
			break
		}
		i += n
	}
	return i
}

func isSpaceBefore(text string, i int) bool {
	r, _ := utf8.DecodeLastRuneInString(text[:i])
	return unicode.IsSpace(r)
}
