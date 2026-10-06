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

package zutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCgroupMemoryLimit(t *testing.T) {
	tests := []struct {
		name   string
		files  map[string]string
		limit  int64
		exists bool
	}{
		{
			name:   "cgroup v2 limit",
			files:  map[string]string{"memory.max": "134217728\n"},
			limit:  134217728,
			exists: true,
		},
		{
			name:  "cgroup v2 unlimited",
			files: map[string]string{"memory.max": "max\n"},
		},
		{
			name:   "cgroup v1 limit",
			files:  map[string]string{"memory/memory.limit_in_bytes": "67108864\n"},
			limit:  67108864,
			exists: true,
		},
		{
			name:  "cgroup v1 unlimited",
			files: map[string]string{"memory/memory.limit_in_bytes": "9223372036854771712\n"},
		},
		{
			name:  "no cgroup files",
			files: map[string]string{},
		},
		{
			name:  "unparseable",
			files: map[string]string{"memory.max": "lots\n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(root, name)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
				require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
			}

			limit, exists := CgroupMemoryLimit(root)
			assert.Equal(t, tt.exists, exists)
			assert.Equal(t, tt.limit, limit)
		})
	}
}
