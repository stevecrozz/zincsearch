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
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// IndexNotFoundError reports a target name that is neither an index nor an alias
type IndexNotFoundError struct {
	Index string
}

func (e *IndexNotFoundError) Error() string {
	return "index " + e.Index + " does not exists"
}

// ResolveTargetIndexes expands a comma-separated list of index names, aliases
// and wildcard patterns into the distinct index names it refers to.
// A name without a wildcard that matches nothing is an *IndexNotFoundError,
// unless ignoreUnavailable is set, in which case it is skipped.
func ResolveTargetIndexes(target string, ignoreUnavailable bool) ([]string, error) {
	var indexNames []string
	seen := make(map[string]struct{})
	add := func(name string) {
		if _, ok := seen[name]; !ok {
			seen[name] = struct{}{}
			indexNames = append(indexNames, name)
		}
	}

	for _, name := range strings.Split(target, ",") {
		if strings.Contains(name, "*") {
			for _, index := range ZINC_INDEX_LIST.List() {
				if IndexNameMatches(name, index.GetName()) {
					add(index.GetName())
				}
			}
			continue
		}
		if _, exists := GetIndex(name); exists {
			add(name)
			continue
		}
		if aliased, ok := ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias(name); ok && len(aliased) > 0 {
			for _, n := range aliased {
				add(n)
			}
			continue
		}
		if ignoreUnavailable {
			continue
		}
		return nil, &IndexNotFoundError{Index: name}
	}

	return indexNames, nil
}

// IndexNameMatches reports whether indexName is name or matches it as a wildcard pattern
func IndexNameMatches(name, indexName string) bool {
	if name == indexName {
		return true
	}

	if strings.Contains(name, "*") {
		p, err := getRegex(name)
		if err != nil {
			log.Err(err).Msg("failed to compile regex")
			return false
		}

		return p.MatchString(indexName)
	}

	return false
}

func getRegex(s string) (*regexp.Regexp, error) {
	parts := strings.Split(s, "*")
	pattern := ""
	for i, part := range parts {
		pattern += part
		if i < len(parts)-1 {
			pattern += "[a-zA-Z0-9_.-]+"
		}
	}

	return regexp.Compile(pattern)
}
