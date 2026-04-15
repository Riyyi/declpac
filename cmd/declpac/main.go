package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/Riyyi/declpac/pkg/input"
	"github.com/Riyyi/declpac/pkg/merge"
	"github.com/Riyyi/declpac/pkg/output"
	"github.com/Riyyi/declpac/pkg/pacman"
	"github.com/Riyyi/declpac/pkg/state"
	"github.com/Riyyi/declpac/pkg/validation"
)

type Config struct {
	StateFiles []string
	NoConfirm  bool
	DryRun     bool
}

func main() {
	cfg := &Config{}

	cmd := &cli.Command{
		Name:  "declpac",
		Usage: "Declarative pacman package manager",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:        "state",
				Aliases:     []string{"s"},
				Usage:       "State file(s) to read package list from",
				Destination: &cfg.StateFiles,
			},
			&cli.BoolFlag{
				Name:        "dry-run",
				Usage:       "Simulate the sync without making changes",
				Destination: &cfg.DryRun,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return run(cfg)
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		os.Exit(1)
	}
}

func run(cfg *Config) error {
	if err := state.OpenLog(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	defer state.Close()

	start := time.Now()
	fmt.Fprintf(os.Stderr, "[debug] run: starting...\n")

	packages, err := input.ReadPackages(cfg.StateFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}
	fmt.Fprintf(os.Stderr, "[debug] run: packages read (%.2fs)\n", time.Since(start).Seconds())

	merged := merge.Merge(packages)

	if !cfg.DryRun {
		if err := validation.CheckDBFreshness(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return err
		}
	}

	if cfg.DryRun {
		result, err := pacman.DryRun(merged)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			return err
		}
		fmt.Println(output.Format(result))
		fmt.Fprintf(os.Stderr, "[debug] run: dry-run done (%.2fs)\n", time.Since(start).Seconds())
		return nil
	}

	result, err := pacman.Sync(merged)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return err
	}

	fmt.Println(output.Format(result))
	fmt.Fprintf(os.Stderr, "[debug] run: sync done (%.2fs)\n", time.Since(start).Seconds())
	return nil
}
