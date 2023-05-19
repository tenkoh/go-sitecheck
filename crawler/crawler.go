// Package crawler provides some implementations of sitecheck.Crawler interface.
package crawler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/tenkoh/go-sitecheck"
)

// IntervalCrawler implements sitecheck.Crawler interface.
type IntervalCrawler struct {
	client   *http.Client
	interval time.Duration
}

// NewIntervalCrawler returns a new IntervalCrawler.
func NewIntervalCrawler(client *http.Client, interval time.Duration) *IntervalCrawler {
	return &IntervalCrawler{
		client:   client,
		interval: interval,
	}
}

// Crawl returns the last modified time of the given urls.
func (c *IntervalCrawler) Crawl(ctx context.Context, urls ...string) (sitecheck.RecentUpdate, error) {
	u := sitecheck.RecentUpdate{}
	var merr error
	for i, url := range urls {
		// do not crawl with too narrow interval
		if i > 0 {
			time.Sleep(c.interval)
		}
		// check context
		select {
		case <-ctx.Done():
			return u, ctx.Err()
		default:
		}

		resp, err := c.client.Head(url)
		if err != nil {
			merr = errors.Join(merr, err)
			continue
		}
		defer resp.Body.Close()
		//Last-Modified: Wed, 21 Oct 2015 07:28:00 GMT

		lu, err := time.Parse("Mon, 2 Jan 2006 15:04:05 GMT", resp.Header.Get("Last-Modified"))
		if err != nil {
			merr = errors.Join(merr, err)
			continue
		}
		u[url] = lu
	}
	if merr != nil {
		return u, merr
	}
	return u, nil
}
