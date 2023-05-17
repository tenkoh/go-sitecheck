package sitecheck_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/tenkoh/go-sitecheck"
)

/*
TODO:
1. 既存の更新日時一覧と、新たに取得した更新日時を比較する。
-> ok
-> 2で網羅可能なためリファクタリングで削除しプライベート関数にする

2. 比較結果に応じて既存の更新日時一覧を更新する。
  - 新たに取得した更新日時が既存の更新日時のどれよりも新しい場合、既存の更新日時一覧に追加する。
  - 新たに取得した更新日時が既存の更新日時のいずれかよりも古い場合、何もしない。
-> ok
*/

func TestUpdateModificationRecord(t *testing.T) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name     string
		record   sitecheck.ModificationRecord
		url      string
		modified time.Time
		want     sitecheck.ModificationRecord
	}{
		{
			"add new modified date",
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
			"https://example.com",
			t1.Add(1 * time.Second),
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1, t1.Add(1 * time.Second)}},
		},
		{
			"no update when modified date is older",
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
			"https://example.com",
			t1.Add(-1 * time.Second),
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
		},
		{
			"no update when modified date is equal to one of existing dates",
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
			"https://example.com",
			t1,
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
		},
		{
			"add new modified date when there is no existing date for the URL",
			sitecheck.ModificationRecord{"https://example1.com": []time.Time{t1}},
			"https://example.com",
			t1,
			sitecheck.ModificationRecord{
				"https://example1.com": []time.Time{t1},
				"https://example.com":  []time.Time{t1},
			},
		},
		{
			"add new modified date when exists is empty",
			sitecheck.ModificationRecord{},
			"https://example.com",
			t1,
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
		},
		{
			"add new modified date when the key exists but the value is empty",
			sitecheck.ModificationRecord{"https://example.com": []time.Time{}},
			"https://example.com",
			t1,
			sitecheck.ModificationRecord{"https://example.com": []time.Time{t1}},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sitecheck.UpdateModificationRecord(tt.record, tt.url, tt.modified)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateModificationRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}
