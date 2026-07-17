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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "1s",
			args: args{
				s: "1s",
			},
			want:    time.Second,
			wantErr: false,
		},
		{
			name: "180s",
			args: args{
				s: "180s",
			},
			want:    time.Second * 180,
			wantErr: false,
		},
		{
			name: "1m",
			args: args{
				s: "1m",
			},
			want:    time.Minute,
			wantErr: false,
		},
		{
			name: "1h",
			args: args{
				s: "1h",
			},
			want:    time.Hour,
			wantErr: false,
		},
		{
			name: "1d",
			args: args{
				s: "1d",
			},
			want:    time.Hour * 24,
			wantErr: false,
		},
		{
			name: "1y",
			args: args{
				s: "1y",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.args.s)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "10s",
			args: args{
				d: time.Second * 10,
			},
			want: "10s",
		},
		{
			name: "5m",
			args: args{
				d: time.Minute * 5,
			},
			want: "5m",
		},
		{
			name: "1h",
			args: args{
				d: time.Hour * 1,
			},
			want: "1h",
		},
		{
			name: "1d",
			args: args{
				d: time.Hour * 24,
			},
			want: "1d",
		},
		{
			name: "1M",
			args: args{
				d: time.Hour * 24 * 30,
			},
			want: "1M",
		},
		{
			name: "1y",
			args: args{
				d: time.Hour * 24 * 30 * 12,
			},
			want: "1y",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.args.d)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnix(t *testing.T) {
	type args struct {
		n int64
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "Zero",
			args: args{
				n: 0,
			},
			want: time.Unix(0, 0),
		},
		{
			name: "Unix",
			args: args{
				n: 1652176732,
			},
			want: time.Unix(1652176732, 0),
		},
		{
			name: "UnixMilli",
			args: args{
				n: 1652176732575,
			},
			want: time.UnixMilli(1652176732575),
		},
		{
			name: "UnixMicro",
			args: args{
				n: 1652176732575067,
			},
			want: time.UnixMicro(1652176732575067),
		},
		{
			name: "UnixNano",
			args: args{
				n: 1652176732575076000,
			},
			want: time.Unix(0, 1652176732575076000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unix(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEsFormatToGoLayout(t *testing.T) {
	tests := []struct {
		name   string
		format string
		want   string
	}{
		{"date_time_no_millis", "date_time_no_millis", "2006-01-02T15:04:05Z07:00"},
		{"strict_date_time_no_millis", "strict_date_time_no_millis", "2006-01-02T15:04:05Z07:00"},
		{"date_time", "date_time", "2006-01-02T15:04:05.000Z07:00"},
		{"date", "date", "2006-01-02"},
		{"basic_date", "basic_date", "20060102"},
		{"passthrough Go layout", "2006-01-02", "2006-01-02"},
		{"passthrough unknown", "yyyy-MM-dd", "yyyy-MM-dd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := esFormatToGoLayout(tt.format)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseTime_ESFormats(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		format  string
		wantErr bool
	}{
		{"date_time_no_millis with Z", "2026-07-17T10:00:00Z", "date_time_no_millis", false},
		{"date_time_no_millis with +00:00", "2026-07-17T10:00:00+00:00", "date_time_no_millis", false},
		{"date_time_no_millis with -07:00", "2026-07-17T10:00:00-07:00", "date_time_no_millis", false},
		{"strict_date_time_no_millis with Z", "2026-07-17T10:00:00Z", "strict_date_time_no_millis", false},
		{"date format", "2026-07-17", "date", false},
		{"invalid value", "not-a-date", "date_time_no_millis", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTime(tt.value, tt.format, "")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	nowStr := time.Now().Format(time.RFC3339)
	now, _ := time.Parse(time.RFC3339, nowStr)

	type args struct {
		value    interface{}
		format   string
		timeZone string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "ParseTime RFC3339",
			args: args{
				value:    now.Format(time.RFC3339),
				format:   "",
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime RFC3339Nano",
			args: args{
				value:    now.Format(time.RFC3339Nano),
				format:   "",
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime RFC1123Z",
			args: args{
				value:    now.Format(time.RFC1123Z),
				format:   time.RFC1123Z,
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime RFC1123",
			args: args{
				value:    now.Format(time.RFC1123),
				format:   time.RFC1123,
				timeZone: time.Local.String(),
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime epoch_millis",
			args: args{
				value:    now.UnixNano(),
				format:   "",
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime epoch_millis",
			args: args{
				value:    now.UnixNano(),
				format:   "epoch_millis",
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
		{
			name: "ParseTime epoch_millis",
			args: args{
				value:    fmt.Sprintf("%d", now.UnixNano()),
				format:   "epoch_millis",
				timeZone: "",
			},
			want:    now,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTime(tt.args.value, tt.args.format, tt.args.timeZone)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.False(t, got.IsZero())
			assert.Equal(t, tt.want.UnixNano(), got.UnixNano())
		})
	}
}
