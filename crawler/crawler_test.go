package crawler_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/tenkoh/go-sitecheck"
	"github.com/tenkoh/go-sitecheck/crawler"
)

func TestIntervalCrawler_Crawl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"HEAD",
		"http://example.com",
		httpmock.NewBytesResponder(200, []byte(``)).HeaderSet(
			http.Header{
				"Last-Modified": []string{"Wed, 21 Oct 2015 07:28:00 GMT"},
			},
		),
	)

	httpmock.RegisterResponder(
		"HEAD",
		"http://example1.com",
		httpmock.NewBytesResponder(200, []byte(``)).HeaderSet(
			http.Header{
				"Last-Modified": []string{"Wed, 21 Oct 2015 07:28:00 GMT"},
			},
		),
	)

	want := sitecheck.RecentUpdate{
		"http://example.com":  time.Date(2015, 10, 21, 7, 28, 0, 0, time.UTC),
		"http://example1.com": time.Date(2015, 10, 21, 7, 28, 0, 0, time.UTC),
	}

	cr := crawler.NewIntervalCrawler(&http.Client{}, 1*time.Second)
	got, err := cr.Crawl(context.Background(), "http://example.com", "http://example1.com")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got = %v, want %v", got, want)
	}

}
