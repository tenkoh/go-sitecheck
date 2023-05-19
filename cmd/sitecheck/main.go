// Package main is the entry point of sitecheck command.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

var (
	appConfig  *config
	configPath string
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
			Usage:   "use proxy server registered in config",
		},
	},
	Action: func(c *cli.Context) error {
		return doCheck()
	},
}

func doCheck() error {
	return nil
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
