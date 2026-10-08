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
	"reflect"
	"strings"
	"time"

	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

// CreateDocument inserts or updates a document in the zinc index
func (index *Index) CreateDocument(docID string, doc map[string]interface{}, update bool) error {
	// metrics
	IncrMetricStatsByIndex(index.GetName(), "wal_request")

	// check WAL
	shard := index.GetShardByDocID(docID)
	if err := shard.OpenWAL(); err != nil {
		return err
	}

	secondShardID := ShardIDNeedLatest
	if update {
		secondShardID = ShardIDNeedUpdate
	}
	data, err := shard.CheckDocument(docID, doc, update, secondShardID)
	if err != nil {
		return err
	}

	return shard.wal.Write(data)
}

// GetDocument get a document in the zinc index
func (index *Index) GetDocument(docID string) (*meta.Hit, error) {
	// check WAL
	shard := index.GetShardByDocID(docID)
	if err := shard.OpenWAL(); err != nil {
		return nil, err
	}

	return shard.FindDocumentByDocID(docID)
}

// Results of MergeDocument, as ES reports them.
const (
	MergeResultCreated = "created"
	MergeResultUpdated = "updated"
	MergeResultNoop    = "noop"
)

// MergeDocument applies an ES-style partial update: partial is merged into
// the document's current source, with objects merged recursively and other
// values replaced. If the document doesn't exist, upsert is inserted
// instead, or errors.ErrorIDNotFound is returned when upsert is nil. As in
// ES, an update that changes nothing is a noop and isn't written.
func (index *Index) MergeDocument(docID string, partial, upsert map[string]interface{}) (string, error) {
	shard := index.GetShardByDocID(docID)
	if err := shard.OpenWAL(); err != nil {
		return "", err
	}

	shard.mergeLock.Lock()
	defer shard.mergeLock.Unlock()

	// earlier writes to this doc may still be in the WAL
	if _, err := shard.DrainWAL(); err != nil {
		return "", err
	}
	hit, err := shard.FindDocumentByDocID(docID)
	if err == errors.ErrorIDNotFound {
		if upsert == nil {
			return "", err
		}
		return MergeResultCreated, index.CreateDocument(docID, upsert, false)
	}
	if err != nil {
		return "", err
	}

	current, _ := hit.Source.(map[string]interface{})
	if current == nil {
		current = make(map[string]interface{})
	}
	if !mergeSource(current, partial) {
		return MergeResultNoop, nil
	}
	return MergeResultUpdated, index.CreateDocument(docID, current, true)
}

// mergeSource merges src into dst the way ES merges a partial update doc,
// and reports whether dst changed.
func mergeSource(dst, src map[string]interface{}) bool {
	changed := false
	for k, v := range src {
		if vm, ok := v.(map[string]interface{}); ok {
			if dm, ok := dst[k].(map[string]interface{}); ok {
				if mergeSource(dm, vm) {
					changed = true
				}
				continue
			}
		}
		if old, ok := dst[k]; !ok || !reflect.DeepEqual(old, v) {
			dst[k] = v
			changed = true
		}
	}
	return changed
}

// UpdateDocument updates a document in the zinc index
func (index *Index) UpdateDocument(docID string, doc map[string]interface{}, insert bool) error {
	// metrics
	IncrMetricStatsByIndex(index.GetName(), "wal_request")

	// check WAL
	shard := index.GetShardByDocID(docID)
	if err := shard.OpenWAL(); err != nil {
		return err
	}

	update := true
	secondShardID, err := shard.FindShardByDocID(docID)
	if err != nil {
		if insert && err == errors.ErrorIDNotFound {
			update = false
		} else {
			return err
		}
	}

	data, err := shard.CheckDocument(docID, doc, update, secondShardID)
	if err != nil {
		return err
	}

	return shard.wal.Write(data)
}

// DeleteDocument deletes a document in the zinc index
func (index *Index) DeleteDocument(docID string) error {
	// metrics
	IncrMetricStatsByIndex(index.GetName(), "wal_request")

	// check WAL
	shard := index.GetShardByDocID(docID)
	if err := shard.OpenWAL(); err != nil {
		return err
	}

	secondShardID, err := shard.FindShardByDocID(docID)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		meta.IDFieldName:     docID,
		meta.ActionFieldName: meta.ActionTypeDelete,
		meta.ShardFieldName:  secondShardID,
	}
	jstr, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return shard.wal.Write(jstr)
}

// isDateProperty returns true if the given value matches the default date format.
func isDateProperty(value string) (string, bool) {
	layout := detectTimeLayout(value)
	if layout == "" {
		return "", false
	}
	_, err := time.Parse(layout, value)
	return layout, err == nil
}

// detectTimeLayout tries to figure out the correct layout of the input date.
func detectTimeLayout(value string) string {
	layout := ""
	switch {
	case len(value) == 19 && strings.Index(value, " ") == 10:
		layout = "2006-01-02 15:04:05"
	case len(value) == 19 && strings.Index(value, "T") == 10:
		layout = "2006-01-02T15:04:05"
	case len(value) == 25 && strings.Index(value, "T") == 10:
		layout = time.RFC3339
	case len(value) == 29 && strings.Index(value, "T") == 10 && strings.Index(value, ".") == 19:
		layout = "2006-01-02T15:04:05.999Z07:00"
	}

	return layout
}
