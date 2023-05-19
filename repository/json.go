// Package repository implements Repository interface
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/tenkoh/go-sitecheck"
)

var _ sitecheck.Repository = (*JSONRepository)(nil)

// JSONRepository implements Repository interface.
// This read and write data in JSON format.
// Save method must be called to save the data.
type JSONRepository struct {
	updates sitecheck.SiteUpdates
}

// NewJSONRepository returns a new JSONRepository.
func NewJSONRepository(r io.Reader) (*JSONRepository, error) {
	u := sitecheck.SiteUpdates{}
	if err := json.NewDecoder(r).Decode(&u); err != nil {
		return nil, fmt.Errorf("error decoding json: %w", err)
	}
	return &JSONRepository{updates: u}, nil
}

// Query returns the site updates for the given urls.
// If a url is not found in the repository, it contains no updates.
func (r *JSONRepository) Query(_ context.Context, urls ...string) (sitecheck.SiteUpdates, error) {
	u := sitecheck.SiteUpdates{}
	for _, url := range urls {
		if us, ok := r.updates[url]; ok {
			u[url] = us
		} else {
			u[url] = []time.Time{}
		}
	}
	return u, nil
}

// Upcert updates the site updates with the given updates.
func (r *JSONRepository) Upcert(_ context.Context, updates sitecheck.SiteUpdates) error {
	for url, ud := range updates {
		r.updates[url] = ud
	}
	return nil
}

// Save saves the site updates to the given writer.
func (r *JSONRepository) Save(w io.Writer) error {
	return json.NewEncoder(w).Encode(r.updates)
}
