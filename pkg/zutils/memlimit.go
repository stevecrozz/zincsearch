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
	"strconv"
	"strings"
)

// cgroup v1 reports "no limit" as a page-aligned value near MaxInt64
const cgroupV1Unlimited = int64(1) << 62

// CgroupMemoryLimit returns the memory limit of the cgroup mounted at root
// (normally /sys/fs/cgroup), trying cgroup v2 and then v1. It returns false
// when there is no limit or it can't be read.
func CgroupMemoryLimit(root string) (int64, bool) {
	for _, name := range []string{"memory.max", filepath.Join("memory", "memory.limit_in_bytes")} {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		limit, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if err != nil || limit <= 0 || limit >= cgroupV1Unlimited {
			return 0, false
		}
		return limit, true
	}
	return 0, false
}
