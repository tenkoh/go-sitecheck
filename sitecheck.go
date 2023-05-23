// Package sitecheck provides core features for sitecheck CLI
package sitecheck

import (
	"context"
	"time"
)

// SiteUpdates is a map of site URLs to their past modified dates.
type SiteUpdates map[string][]time.Time

// RecentUpdate is a map of site URLs to their most recent modified date.
type RecentUpdate map[string]time.Time

// Repository is an interface for data storage
// to read and write records of sites' modified dates.
type Repository interface {
	Query(ctx context.Context, urls ...string) (SiteUpdates, error)
	Upcert(ctx context.Context, updates SiteUpdates) error
}

// Crawler is an interface for crawling sites.
type Crawler interface {
	Crawl(ctx context.Context, urls ...string) (RecentUpdate, error)
}

// isModified returns true if the modified date is newer than any of the existing modified dates.
// When there is no existing modified date, it returns true.
func isModified(exists []time.Time, modified time.Time) bool {
	for _, t := range exists {
		if t.Compare(modified) >= 0 {
			return false
		}
	}
	return true
}

// GetUpdated returns a new modification record
// contains differences between the existing record and the new modified date.
func GetUpdated(exist SiteUpdates, recent RecentUpdate) SiteUpdates {
	diff := SiteUpdates{}
	for url, m := range recent {
		ms, ok := exist[url]
		if !ok {
			diff[url] = []time.Time{m}
			continue
		}
		if isModified(ms, m) {
			diff[url] = append(ms, m)
		}
	}
	return diff
}
