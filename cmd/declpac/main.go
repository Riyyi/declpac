package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/Riyyi/declpac/pkg/auth"
	"github.com/Riyyi/declpac/pkg/backend"
	"github.com/Riyyi/declpac/pkg/input"
	"github.com/Riyyi/declpac/pkg/log"
	"github.com/Riyyi/declpac/pkg/output"

	_ "github.com/Riyyi/declpac/pkg/dnf"
	_ "github.com/Riyyi/declpac/pkg/pacman"
)

type Config struct {
	StateFiles []string
	NoCheck    bool
	DryRun     bool
	Prune      bool
	Verbose    bool
}

func main() {
	cfg := &Config{}

	cmd := &cli.Command{
		Name:  "declpac",
		Usage: "Declarative package manager",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "state",
				Aliases:     []string{"s"},
				Usage:       "State file(s) to read package list from",
				Destination: &cfg.StateFiles,
			},
			&cli.BoolFlag{
				Name:        "nocheck",
				Usage:       "Skip safety check",
				Destination: &cfg.NoCheck,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Usage:       "Simulate the sync without making changes",
				Destination: &cfg.DryRun,
			},
			&cli.BoolFlag{
				Name:        "prune",
				Usage:       "Mark packages and cleanup orphans",
				Destination: &cfg.Prune,
			},
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Usage:       "Enable verbose output",
				Destination: &cfg.Verbose,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			log.Verbose = cfg.Verbose
			return run(cfg)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}

func run(cfg *Config) error {
	start := time.Now()
	log.Debug("run: starting...")

	b, err := backend.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	packages, err := input.ReadPackages(cfg.StateFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	log.Debug("run: packages read (%.2fs)", time.Since(start).Seconds())

	merged, err := input.Merge(packages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	if cfg.DryRun {
		result, err := backend.DryRun(b, merged)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return err
		}
		fmt.Println(output.Format(result))
		log.Debug("run: dry-run done (%.2fs)", time.Since(start).Seconds())
		return nil
	}

	auth.Start()

	if err := log.OpenLog(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	defer log.Close()

	result, err := backend.Sync(b, merged, cfg.NoCheck, cfg.Prune)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Println(output.Format(result))
	log.Debug("run: sync done (%.2fs)", time.Since(start).Seconds())
	return nil
}
