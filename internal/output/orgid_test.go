package output

import (
	"bytes"
	"strings"
	"testing"

	"meraki-info/internal/meraki"
)

func TestDeviceWithNetworkOrganizationID(t *testing.T) {
	device := meraki.DeviceWithNetwork{
		Device: meraki.Device{
			Serial:         "Q2XX-XXXX-XXXX",
			Name:           "Test Device",
			Model:          "MX64",
			MAC:            "00:11:22:33:44:55",
			Status:         "alerting",
			LastReportedAt: "2025-07-17T10:30:00Z",
			ProductType:    "appliance",
			Tags:           []string{"test", "alerting"},
		},
		NetworkName:    "Test Network",
		NetworkID:      "N_123456789",
		Organization:   "Test Organization",
		OrganizationID: "123456",
	}

	devices := []meraki.DeviceWithNetwork{device}

	// Test text output includes organization ID
	writer := NewWriter("text")
	var buf bytes.Buffer
	if err := writer.WriteTo(devices, &buf); err != nil {
		t.Fatalf("Failed to write devices: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Organization: Test Organization") {
		t.Error("Expected organization name in output")
	}
	if !strings.Contains(output, "Organization ID: 123456") {
		t.Error("Expected organization ID in output")
	}
	if !strings.Contains(output, "Network Name: Test Network") {
		t.Error("Expected network name in output")
	}
	if !strings.Contains(output, "Network ID: N_123456789") {
		t.Error("Expected network ID in output")
	}
}

func TestLicenseWithNetworkOrganizationID(t *testing.T) {
	license := meraki.LicenseWithNetwork{
		License: meraki.License{
			ID:             "L_123456789",
			DeviceSerial:   "Q2XX-XXXX-XXXX",
			NetworkID:      "N_123456789",
			State:          "active",
			Edition:        "enterprise",
			Mode:           "addons",
			LicenseType:    "per-device",
			LicenseKey:     "XXXX-XXXX-XXXX-XXXX",
			OrderNumber:    "123456789",
			DurationInDays: 365,
			ExpirationDate: "2026-07-17T10:30:00Z",
		},
		Organization:   "Test Organization",
		OrganizationID: "123456",
	}

	licenses := []meraki.LicenseWithNetwork{license}

	// Test text output includes organization ID
	writer := NewWriter("text")
	var buf bytes.Buffer
	if err := writer.WriteTo(licenses, &buf); err != nil {
		t.Fatalf("Failed to write licenses: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Organization: Test Organization") {
		t.Error("Expected organization name in output")
	}
	if !strings.Contains(output, "Organization ID: 123456") {
		t.Error("Expected organization ID in output")
	}
}
