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
	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/uquery/query"
)

func init() {
	query.TermsLookupFunc = lookupTermsSource
}

// lookupTermsSource fetches the _source of a document for a terms lookup
// query, resolving aliases. A missing index or document returns nil.
func lookupTermsSource(indexName, id string) (map[string]interface{}, error) {
	names := []string{indexName}
	if indexes, ok := ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias(indexName); ok && len(indexes) > 0 {
		names = indexes
	}
	for _, name := range names {
		index, ok := GetIndex(name)
		if !ok {
			continue
		}
		hit, err := index.GetDocument(id)
		if err == errors.ErrorIDNotFound {
			continue
		}
		if err != nil {
			return nil, err
		}
		if source, ok := hit.Source.(map[string]interface{}); ok {
			return source, nil
		}
	}
	return nil, nil
}
