// Package main is the entry point of sitecheck command.
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	defaultInterval = 1
)

type config struct {
	Proxy    string   `json:"proxy"`
	Interval int      `json:"interval"`
	URLs     []string `json:"urls"`
}

func (c *config) setProxy(proxy string) error {
	c.Proxy = proxy
	return nil
}

func (c *config) setInterval(interval int) error {
	if interval < 1 {
		return fmt.Errorf("interval must be greater than 0, got %d", interval)
	}
	c.Interval = interval
	return nil
}

func (c *config) addURL(url string) {
	c.URLs = append(c.URLs, url)
}

func (c *config) removeURL(url string) {
	for i, u := range c.URLs {
		if u == url {
			c.URLs = append(c.URLs[:i], c.URLs[i+1:]...)
			break
		}
	}
}

func loadConfig(path string) (*config, error) {
	var cfg config
	_, err := os.Stat(path)
	if err == nil {
		f, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open config file %s: %w", path, err)
		}
		defer f.Close()
		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return nil, fmt.Errorf("failed to decode config file %s: %w", path, err)
		}
		return &cfg, nil
	}

	// when config file does not exist, create it with default values.
	cfg.Interval = defaultInterval
	cfg.URLs = []string{}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return nil, fmt.Errorf("failed to create config file %s: %w", path, err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize config file %s: %w", path, err)
	}

	return &cfg, nil
}

// save saves the config to the specified path.
// if the file exists, it will be overwritten.
func (c *config) save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to open config file %s: %w", path, err)
	}
	defer f.Close()

	b, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	if _, err := f.Write(b); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
