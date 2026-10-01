package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"
	"github.com/woodleighschool/jamf-user-sync/internal/config"
	"github.com/woodleighschool/jamf-user-sync/internal/syncer"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	cancel()
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := pflag.NewFlagSet("jamf-user-sync", pflag.ContinueOnError)
	flags.SetOutput(stderr)
	dryRun := flags.Bool("dry-run", false, "Inspect differences without updating Jamf")
	showVersion := flags.Bool("version", false, "Print build version")
	help := flags.BoolP("help", "h", false, "Print usage")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *help {
		if _, err := fmt.Fprintln(stdout, "Usage: jamf-user-sync [--dry-run] [--version]\n\nSynchronize AD user and location fields to managed Jamf Macs and iOS devices once."); err != nil {
			return 1
		}
		flags.SetOutput(stdout)
		flags.PrintDefaults()
		return 0
	}
	if *showVersion {
		if _, err := fmt.Fprintf(stdout, "jamf-user-sync %s (%s, %s)\n", version, commit, date); err != nil {
			return 1
		}
		return 0
	}
	if flags.NArg() != 0 {
		_, _ = fmt.Fprintln(stderr, "jamf-user-sync: unexpected positional arguments")
		return 2
	}
	logger := slog.New(slog.NewJSONHandler(stderr, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.ErrorContext(ctx, "configuration failed", "error", err)
		return 1
	}
	if flags.Changed("dry-run") {
		cfg.DryRun = *dryRun
	}
	logger = slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: cfg.LogLevel}))
	ctx, cancel := context.WithTimeout(ctx, cfg.RunTimeout)
	defer cancel()
	jamf, err := syncer.NewJamf(cfg)
	if err != nil {
		logger.ErrorContext(ctx, "Jamf initialization failed", "error", err)
		return 1
	}
	defer func() {
		if err := jamf.Close(); err != nil {
			logger.WarnContext(ctx, "Jamf client closure failed")
		}
	}()
	directory, err := syncer.NewLDAP(ctx, cfg)
	if err != nil {
		logger.ErrorContext(ctx, "directory initialization failed", "error", err)
		return 1
	}
	defer directory.Close()
	summary, syncErr := syncer.Run(ctx, jamf, directory, cfg.DryRun, logger)
	if err := json.NewEncoder(stdout).Encode(summary); err != nil {
		logger.ErrorContext(ctx, "summary output failed")
		return 1
	}
	if syncErr != nil {
		logger.ErrorContext(ctx, "sync failed", "error", syncErr)
		return 1
	}
	logger.InfoContext(ctx, "sync complete", "devices", summary.Devices, "updated", summary.Updated, "would_update", summary.WouldUpdate, "dry_run", summary.DryRun)
	return 0
}
