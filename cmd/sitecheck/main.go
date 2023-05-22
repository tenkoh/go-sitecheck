// Package main is the entry point of sitecheck command.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/tenkoh/go-sitecheck"
	"github.com/tenkoh/go-sitecheck/crawler"
	"github.com/tenkoh/go-sitecheck/repository"
	"github.com/urfave/cli/v2"
)

var (
	appConfig  *config
	configPath string
)

const (
	repositoryPath = "repository.json"
)

func init() {
	var dir string
	if runtime.GOOS == "windows" {
		dir = os.Getenv("APPDATA")
	} else {
		dir = os.Getenv("HOME")
		dir = filepath.Join(dir, ".config")
	}
	dir = filepath.Join(dir, "sitecheck")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}
	configPath = filepath.Join(dir, "config.json")
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}
	appConfig = cfg
}

var configCmd = &cli.Command{
	Name:  "config",
	Usage: "manage config values",
	Subcommands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "show config list",
			Action: func(c *cli.Context) error {
				fmt.Printf("proxy: %s\n", appConfig.Proxy)
				fmt.Printf("interval: %d sec\n", appConfig.Interval)
				return nil
			},
		},
		{
			Name:  "set",
			Usage: "set a config value",
			UsageText: `command: sitecheck config set <key> <value>

Example:
sitecheck config set proxy http://example-proxy.com:8000
sitecheck config set interval 5

if you want to delete a config value, set empty string to <value>.`,
			Action: func(c *cli.Context) error {
				if c.NArg() != 2 {
					return fmt.Errorf("invalid argument. two arguments are required. got %d", c.NArg())
				}
				key := c.Args().Get(0)
				value := c.Args().Get(1)
				key = strings.TrimSpace(key)
				switch key {
				case "proxy":
					appConfig.setProxy(value)
				case "interval":
					// convert value to int64
					i, err := strconv.Atoi(value)
					if err != nil {
						return fmt.Errorf("invalid interval value: %s", value)
					}
					if err := appConfig.setInterval(i); err != nil {
						return err
					}
				default:
					return fmt.Errorf("invalid key: %s", key)
				}
				return appConfig.save(configPath)
			},
		},
		{
			Name:  "edit",
			Usage: "edit show the config file's path on your terminal",
			Action: func(c *cli.Context) error {
				fmt.Printf("%s\n", configPath)
				return nil
			},
		},
	},
}

var siteCmd = &cli.Command{
	Name:  "site",
	Usage: "manage Web site urls to check updates",
	Subcommands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "show site list",
			Action: func(c *cli.Context) error {
				if len(appConfig.URLs) == 0 {
					fmt.Println("no site url is registered")
					return nil
				}
				for _, url := range appConfig.URLs {
					fmt.Println(url)
				}
				return nil
			},
		},
		{
			Name:      "add",
			Usage:     "add a site url",
			UsageText: "command: sitecheck site add <url>",
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					return fmt.Errorf("invalid argument. only one argument is allowed. got %d", c.NArg())
				}
				u := c.Args().First()
				u = strings.TrimSpace(u)
				appConfig.addURL(u)
				return appConfig.save(configPath)
			},
		},
		{
			Name:      "delete",
			Usage:     "delete a site url",
			UsageText: "command: sitecheck site delete <url>",
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					return fmt.Errorf("invalid argument. only one argument is allowed. got %d", c.NArg())
				}
				u := c.Args().First()
				u = strings.TrimSpace(u)
				appConfig.removeURL(u)
				return appConfig.save(configPath)
			},
		},
	},
}

var checkCmd = &cli.Command{
	Name:  "check",
	Usage: "check updates of the registered sites",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "output",
			Aliases: []string{"o"},
			Usage:   "set output file path",
		},
		&cli.BoolFlag{
			Name:    "proxy",
			Aliases: []string{"p"},
			Usage:   "use proxy server registered in config.",
		},
	},
	Action: func(c *cli.Context) error {
		client, err := customeClient(c)
		if err != nil {
			return err
		}
		cr := crawler.NewIntervalCrawler(client, appConfig.Interval)

		repo, err := newRepository(repositoryPath)
		if err != nil {
			return fmt.Errorf("fail to create repository: %w", err)
		}

		ctx := context.Background()
		exists, err := repo.Query(ctx, appConfig.URLs...)
		if err != nil {
			return fmt.Errorf("fail to query: %w", err)
		}

		founds, err := cr.Crawl(ctx, appConfig.URLs...)
		if err != nil {
			return fmt.Errorf("fail to crawl: %w", err)
		}

		// write log
		var w io.Writer
		w = os.Stdout
		if c.String("output") != "" {
			f, err := os.Create(c.String("output"))
			if err != nil {
				return fmt.Errorf("fail to create %s: %w", c.String("output"), err)
			}
			defer f.Close()
			w = io.MultiWriter(os.Stdout, f)
		}

		updates := sitecheck.GetUpdated(exists, founds)
		if len(founds) == 0 || len(updates) == 0 {
			fmt.Fprint(w, "no updates\n")
			return nil
		}
		for u, mod := range updates {
			fmt.Fprintf(w, "%s: %s\n", u, mod)
		}
		// finish writing log

		// save repository
		if err := repo.Upcert(ctx, updates); err != nil {
			return fmt.Errorf("fail to upcert: %w", err)
		}

		f, err := os.Create(repositoryPath)
		if err != nil {
			return fmt.Errorf("fail to save %s: %w", repositoryPath, err)
		}
		defer f.Close()
		if err := repo.Save(f); err != nil {
			return fmt.Errorf("fail to save %s: %w", repositoryPath, err)
		}
		return nil
	},
}

func initJSON(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("fail to create %s: %w", path, err)
	}
	defer f.Close()

	// initialize repository JSON file.
	_, err = f.WriteString("{}")
	if err != nil {
		return fmt.Errorf("fail to write %s: %w", path, err)
	}
	return nil
}

func customeClient(c *cli.Context) (*http.Client, error) {
	client := http.DefaultClient
	client.Timeout = time.Duration(appConfig.Interval*5) * time.Second
	// use the registered proxy in config if proxy flag is set.
	if c.Bool("proxy") {
		if appConfig.Proxy == "" {
			return nil, errors.New("proxy is not registered in config")
		}
		u, err := url.Parse(appConfig.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy url %s: %w", appConfig.Proxy, err)
		}
		transport := &http.Transport{
			Proxy: http.ProxyURL(u),
		}
		client.Transport = transport
	}
	return client, nil
}

func newRepository(path string) (*repository.JSONRepository, error) {
	// open repository JSON file. if not exists, create and initialize it.
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("fail to open %s: %w", repositoryPath, err)
		}
		if err := initJSON(path); err != nil {
			return nil, fmt.Errorf("fail to initialize %s: %w", repositoryPath, err)
		}
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("fail to create %s: %w", repositoryPath, err)
	}
	defer f.Close()

	repo, err := repository.NewJSONRepository(f)
	if err != nil {
		return nil, fmt.Errorf("fail to open json repository: %w", err)
	}
	return repo, nil
}

func main() {
	app := &cli.App{
		Name:      "sitecheck",
		Version:   "0.0.1",
		Usage:     "check updates of the specified sites",
		UsageText: "sitecheck command [command options] subcommand [arguments...]",
		Commands:  []*cli.Command{configCmd, siteCmd, checkCmd},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
