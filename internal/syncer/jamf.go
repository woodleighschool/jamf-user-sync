package syncer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-sdk-jamfpro-v2/jamfpro"
	"github.com/deploymenttheory/go-sdk-jamfpro-v2/jamfpro/constants"
	"github.com/woodleighschool/jamf-user-sync/internal/config"
	"go.uber.org/zap"
	"resty.dev/v3"
)

type Jamf struct {
	client *jamfpro.Client
}

func NewJamf(cfg config.Config) (*Jamf, error) {
	client, err := jamfpro.NewClient(&jamfpro.AuthConfig{
		InstanceDomain: cfg.JamfHost, AuthMethod: constants.AuthMethodOAuth2,
		ClientID: cfg.JamfClientID, ClientSecret: cfg.JamfSecret, HideSensitiveData: true,
	}, jamfpro.WithTimeout(cfg.RequestTimeout), jamfpro.WithLogger(zap.NewNop()))
	if err != nil {
		return nil, errors.New("jamf: could not configure SDK client")
	}
	return &Jamf{client: client}, nil
}

func (j *Jamf) Close() error {
	return j.client.GetTransport().GetHTTPClient().Close()
}

func (j *Jamf) Devices(ctx context.Context) ([]Device, error) {
	computers, response, err := j.client.JamfProAPI.ComputerInventory.ListV4(ctx, map[string]string{
		"section": "GENERAL,USER_AND_LOCATION", "filter": "general.remoteManagement.managed==true", "sort": "id:asc",
	})
	if err != nil {
		return nil, jamfError(ctx, "list computers", response)
	}
	devices := make([]Device, 0, len(computers.Results))
	for _, computer := range computers.Results {
		if !computer.General.RemoteManagement.Managed {
			continue
		}
		if computer.ID == "" {
			return nil, errors.New("jamf: computer inventory is missing a device ID")
		}
		user := computer.UserAndLocation
		devices = append(devices, Device{ID: computer.ID, Kind: "macos", User: User{
			Username: user.Username, RealName: user.Realname, Email: user.Email,
			BuildingID: user.BuildingId, DepartmentID: user.DepartmentId,
		}})
	}
	mobile, response, err := j.client.JamfProAPI.MobileDevices.GetDetailV2(ctx, map[string]string{
		"section": "GENERAL,USER_AND_LOCATION", "filter": "managed==true", "sort": "mobileDeviceId:asc",
	})
	if err != nil {
		return nil, jamfError(ctx, "list mobile devices", response)
	}
	if mobile.TotalCount != len(mobile.Results) {
		return nil, errors.New("jamf: mobile inventory returned incomplete pagination")
	}
	for _, mobile := range mobile.Results {
		if !strings.EqualFold(mobile.DeviceType, "ios") {
			continue
		}
		if mobile.General == nil || mobile.UserAndLocation == nil || mobile.MobileDeviceID == "" {
			return nil, errors.New("jamf: mobile inventory is missing required sections or a device ID")
		}
		if !mobile.General.Managed {
			continue
		}
		user := mobile.UserAndLocation
		devices = append(devices, Device{ID: mobile.MobileDeviceID, Kind: "ios", User: User{
			Username: user.Username, RealName: user.RealName, Email: user.EmailAddress,
			BuildingID: user.BuildingID, DepartmentID: user.DepartmentID,
		}})
	}
	return devices, nil
}

func (j *Jamf) Update(ctx context.Context, device Device, user User) error {
	var path, section, nameField, emailField string
	switch device.Kind {
	case "macos":
		path = constants.EndpointJamfProComputerInventoryV4 + "-detail/" + device.ID
		section, nameField, emailField = "userAndLocation", "realname", "email"
	case "ios":
		path = constants.EndpointJamfProMobileDevicesV2 + "/" + device.ID
		section, nameField, emailField = "location", "realName", "emailAddress"
	default:
		return errors.New("jamf: unsupported device kind")
	}
	if device.ID == "" {
		return errors.New("jamf: device ID is required")
	}
	// The SDK's computer update DTO serializes unrelated zero-value inventory
	// sections, and its mobile DTO omits empty strings. Use the SDK's supported
	// request builder to preserve ownership and explicit name/email clearing.
	body := map[string]any{section: map[string]string{
		"username": user.Username, nameField: user.RealName, emailField: user.Email,
		"buildingId": user.BuildingID, "departmentId": user.DepartmentID,
	}}
	response, err := j.client.GetTransport().NewRequest(ctx).
		SetHeader("Accept", constants.ApplicationJSON).
		SetHeader("Content-Type", constants.ApplicationJSON).
		SetBody(body).Patch(path)
	if err != nil {
		return jamfError(ctx, "update user", response)
	}
	return nil
}

func jamfError(ctx context.Context, operation string, response *resty.Response) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("jamf %s: %w", operation, err)
	}
	// SDK error bodies can contain personal information. Keep diagnostics to
	// the operation and status while the run summary counts affected devices.
	if response != nil {
		return fmt.Errorf("jamf %s: HTTP %d", operation, response.StatusCode())
	}
	return fmt.Errorf("jamf %s: request failed", operation)
}
