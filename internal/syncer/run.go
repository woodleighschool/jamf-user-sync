package syncer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type Device struct {
	ID   string
	Kind string
	User User
}

type Inventory interface {
	Devices(context.Context) ([]Device, error)
	Update(context.Context, Device, User) error
}

type Directory interface {
	Lookup(context.Context, string) (User, error)
}

type Summary struct {
	Devices      int  `json:"devices"`
	Unassigned   int  `json:"unassigned"`
	Unchanged    int  `json:"unchanged"`
	WouldUpdate  int  `json:"would_update"`
	Updated      int  `json:"updated"`
	LookupFailed int  `json:"lookup_failed"`
	UpdateFailed int  `json:"update_failed"`
	DryRun       bool `json:"dry_run"`
}

// Run reconciles the five AD-owned fields; failures on one device do not stop others.
func Run(ctx context.Context, inventory Inventory, directory Directory, dryRun bool, logger *slog.Logger) (Summary, error) {
	summary := Summary{DryRun: dryRun}
	devices, err := inventory.Devices(ctx)
	if err != nil {
		return summary, fmt.Errorf("inventory: %w", err)
	}
	summary.Devices = len(devices)
	for _, device := range devices {
		if err := ctx.Err(); err != nil {
			return summary, fmt.Errorf("sync: %w", err)
		}
		if strings.TrimSpace(device.User.Username) == "" {
			summary.Unassigned++
			continue
		}
		user, err := directory.Lookup(ctx, device.User.Username)
		if err != nil {
			if ctx.Err() != nil {
				return summary, fmt.Errorf("sync: %w", ctx.Err())
			}
			summary.LookupFailed++
			logger.ErrorContext(ctx, "device synchronization failed", "device_kind", device.Kind, "device_id", device.ID, "operation", "directory_lookup", "error", err)
			continue
		}
		if user == device.User {
			summary.Unchanged++
			continue
		}
		summary.WouldUpdate++
		if dryRun {
			continue
		}
		if err := inventory.Update(ctx, device, user); err != nil {
			if ctx.Err() != nil {
				return summary, fmt.Errorf("sync: %w", ctx.Err())
			}
			summary.UpdateFailed++
			logger.ErrorContext(ctx, "device synchronization failed", "device_kind", device.Kind, "device_id", device.ID, "operation", "jamf_update", "error", err)
			continue
		}
		summary.Updated++
	}
	if summary.LookupFailed+summary.UpdateFailed != 0 {
		return summary, errors.New("sync: one or more devices failed")
	}
	return summary, nil
}
