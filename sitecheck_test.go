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

2. 比較結果に応じて既存の更新日時一覧を更新する。更新差分を返す
  - 新たに取得した更新日時が既存の更新日時のどれよりも新しい場合、既存の更新日時一覧に追加する。
  - 新たに取得した更新日時が既存の更新日時のいずれかよりも古い場合、何もしない。
-> ok
*/

func TestGetUpdated(t *testing.T) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		exist  sitecheck.SiteUpdates
		recent sitecheck.RecentUpdate
		want   sitecheck.SiteUpdates
	}{
		{
			"add new modified date",
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
			sitecheck.RecentUpdate{"https://example.com": t1.Add(1 * time.Second)},
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1, t1.Add(1 * time.Second)}},
		},
		{
			"no update when modified date is older",
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
			sitecheck.RecentUpdate{"https://example.com": t1.Add(-1 * time.Second)},
			sitecheck.SiteUpdates{},
		},
		{
			"no update when modified date is equal to one of existing dates",
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
			sitecheck.RecentUpdate{"https://example.com": t1},
			sitecheck.SiteUpdates{},
		},
		{
			"add new modified date when there is no existing date for the URL",
			sitecheck.SiteUpdates{"https://example1.com": []time.Time{t1}},
			sitecheck.RecentUpdate{"https://example.com": t1},
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
		},
		{
			"add new modified date when exists is empty",
			sitecheck.SiteUpdates{},
			sitecheck.RecentUpdate{"https://example.com": t1},
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
		},
		{
			"add new modified date when the key exists but the value is empty",
			sitecheck.SiteUpdates{"https://example.com": []time.Time{}},
			sitecheck.RecentUpdate{"https://example.com": t1},
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
		},
		{
			"do nothing when the empty update is passed",
			sitecheck.SiteUpdates{"https://example.com": []time.Time{t1}},
			sitecheck.RecentUpdate{},
			sitecheck.SiteUpdates{},
		},
	}

	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sitecheck.GetUpdated(tt.exist, tt.recent)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDiffRecord() = %v, want %v", got, tt.want)
			}
		})
	}
}
