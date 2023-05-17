// Package sitecheck provides core features for sitecheck CLI
package sitecheck

import "time"

// ModificationRecord is a map of site URLs to their modified dates.
type ModificationRecord map[string][]time.Time

// Repository is an interface for data storage
// to read and write records of sites' modified dates.
type Repository interface{}

// Crawler is an interface for crawling sites.
type Crawler interface{}

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

// UpdateModificationRecord returns a new modification record updated
// with the given URL and modified date.
func UpdateModificationRecord(
	exists ModificationRecord,
	url string,
	modified time.Time,
) ModificationRecord {
	m := ModificationRecord{}
	for k, v := range exists {
		m[k] = v
	}
	if ms, ok := m[url]; ok {
		if isModified(ms, modified) {
			m[url] = append(ms, modified)
		}
	} else {
		m[url] = []time.Time{modified}
	}
	return m
}
