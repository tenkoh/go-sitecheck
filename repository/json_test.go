package repository_test

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	sc "github.com/tenkoh/go-sitecheck"
	repo "github.com/tenkoh/go-sitecheck/repository"
)

func TestJSONRepository_Query(t *testing.T) {
	s := `{"a":["2020-01-01T00:00:00Z"],"b":["2020-01-01T00:00:00Z"]}`
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	r, _ := repo.NewJSONRepository(strings.NewReader(s))
	t.Parallel()
	tests := []struct {
		name    string
		r       *repo.JSONRepository
		urls    []string
		want    sc.SiteUpdates
		wantErr bool
	}{
		{
			"common key exists",
			r,
			[]string{"a"},
			sc.SiteUpdates{"a": []time.Time{t1}},
			false,
		}, {
			"common key does not exist. return map with empty value",
			r,
			[]string{"c"},
			sc.SiteUpdates{"c": []time.Time{}},
			false,
		}, {
			"multiple keys",
			r,
			[]string{"a", "b"},
			sc.SiteUpdates{"a": []time.Time{t1}, "b": []time.Time{t1}},
			false,
		}, {
			"empty keys",
			r,
			[]string{},
			sc.SiteUpdates{},
			false,
		},
	}
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.r.Query(ctx, tt.urls...)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("unintended error: JSONRepository.Query() error = %v", err)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JSONRepository.Query() = %v, want %v", got, tt.want)
			}
		})
	}
}
