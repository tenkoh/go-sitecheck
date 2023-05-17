package sitecheck_test

import (
	"testing"
	"time"

	"github.com/tenkoh/go-sitecheck"
)

/*
TODO:
ok 1. 既存の更新日時一覧と、新たに取得した更新日時を比較する。
2. 更新日時が既存の更新日時のどれよりも新しい場合、更新日時を追加して保存する。
3. 更新日時が既存の更新日時のどれよりも古い場合、何もしない。
*/

func TestIsModified(t *testing.T) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		exists   []time.Time
		modified time.Time
		want     bool
	}{
		{"modified is after existing dates", []time.Time{t1}, t1.Add(1 * time.Second), true},
		{"modified is equal to one of existing dates", []time.Time{t1}, t1, false},
		{"modified is before existing dates", []time.Time{t1}, t1.Add(-1 * time.Second), false},
		{"empty input", []time.Time{}, t1, true},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sitecheck.IsModified(tt.exists, tt.modified)
			if got != tt.want {
				t.Errorf("isModified() = %t, want %t", got, tt.want)
			}
		})
	}
}
