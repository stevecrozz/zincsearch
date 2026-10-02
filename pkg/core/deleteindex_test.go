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
)

func TestDeleteIndex(t *testing.T) {
	indexName := "TestDeleteIndex.index_1"
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "exist",
			args: args{
				name: indexName,
			},
			wantErr: false,
		},
		{
			name: "not exist",
			args: args{
				name: "my-index-not-exist",
			},
			wantErr: true,
		},
	}

	t.Run("prepare", func(t *testing.T) {
		index, err := NewIndex(indexName, "disk", 2)
		assert.NoError(t, err)
		assert.NotNil(t, index)
		err = StoreIndex(index)
		assert.NoError(t, err)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteIndex(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("DeleteIndex() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteIndexRemovesAliases(t *testing.T) {
	indexName := "TestDeleteIndexRemovesAliases.index_1"
	otherName := "TestDeleteIndexRemovesAliases.index_2"
	for _, name := range []string{indexName, otherName} {
		index, err := NewIndex(name, "disk", 2)
		require.NoError(t, err)
		require.NoError(t, StoreIndex(index))
	}
	defer func() { _ = DeleteIndex(otherName) }()

	require.NoError(t, ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias("TestDeleteIndexRemovesAliases.shared", []string{indexName, otherName}))
	require.NoError(t, ZINC_INDEX_ALIAS_LIST.AddIndexesToAlias("TestDeleteIndexRemovesAliases.only", []string{indexName}))
	defer func() { _ = ZINC_INDEX_ALIAS_LIST.RemoveIndexFromAllAliases(otherName) }()

	require.NoError(t, DeleteIndex(indexName))

	assert.Empty(t, ZINC_INDEX_ALIAS_LIST.GetAliasesForIndex(indexName))
	indexes, ok := ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias("TestDeleteIndexRemovesAliases.shared")
	assert.True(t, ok)
	assert.Equal(t, []string{otherName}, indexes)
	_, ok = ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias("TestDeleteIndexRemovesAliases.only")
	assert.False(t, ok)
}
