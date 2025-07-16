package config

import (
	"flag"
	"os"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	// Save original environment
	originalOrg := os.Getenv("MERAKI_ORG")
	originalNet := os.Getenv("MERAKI_NET")
	originalKey := os.Getenv("MERAKI_APIKEY")

	defer func() {
		// Restore original environment
		os.Setenv("MERAKI_ORG", originalOrg)
		os.Setenv("MERAKI_NET", originalNet)
		os.Setenv("MERAKI_APIKEY", originalKey)
		// Reset flag package for next test
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	t.Run("missing API key should return error", func(t *testing.T) {
		// Clear environment
		os.Unsetenv("MERAKI_APIKEY")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--access"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when API key is missing")
		}
		if err != nil && !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected API key error, got: %v", err)
		}
	})

	t.Run("missing required action flag should return error", func(t *testing.T) {
		// Set API key but no action flags
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--org", "test-org"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when no action flag is provided")
		}
		if err != nil && !strings.Contains(err.Error(), "one of the parameters") {
			t.Errorf("Expected action flag error, got: %v", err)
		}
	})

	t.Run("multiple action flags should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with multiple action flags
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--access", "--licenses", "--org", "test-org"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when multiple action flags are provided")
		}
		if err != nil && !strings.Contains(err.Error(), "only one of") {
			t.Errorf("Expected multiple action flags error, got: %v", err)
		}
	})

	t.Run("valid access mode config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")
		os.Setenv("MERAKI_ORG", "test-org")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--access"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if cfg.APIKey != "test-key" {
			t.Errorf("Expected APIKey 'test-key', got '%s'", cfg.APIKey)
		}
		if cfg.Organization != "test-org" {
			t.Errorf("Expected Organization 'test-org', got '%s'", cfg.Organization)
		}
		if !cfg.ShowAccess {
			t.Error("Expected ShowAccess to be true")
		}
		if cfg.ShowLicenses || cfg.ShowRouteTables {
			t.Error("Expected other action flags to be false")
		}
	})

	t.Run("valid licenses mode config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")
		os.Setenv("MERAKI_ORG", "test-org")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--licenses", "--org", "test-org"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !cfg.ShowLicenses {
			t.Error("Expected ShowLicenses to be true")
		}
		if cfg.ShowAccess || cfg.ShowRouteTables {
			t.Error("Expected other action flags to be false")
		}
	})

	t.Run("valid route-tables mode config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")
		os.Setenv("MERAKI_ORG", "test-org")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--route-tables", "--org", "test-org"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if !cfg.ShowRouteTables {
			t.Error("Expected ShowRouteTables to be true")
		}
		if cfg.ShowAccess || cfg.ShowLicenses {
			t.Error("Expected other action flags to be false")
		}
	})

	t.Run("all flag with network should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--licenses", "--org", "test-org", "--network", "test-net", "--all"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using --all with --network")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot specify --network when using --all") {
			t.Errorf("Expected --all/--network conflict error, got: %v", err)
		}
	})

	t.Run("all flag with stdout output should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--licenses", "--org", "test-org", "--all", "--output", "-"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using --all with stdout output")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot use --output '-' or omit --output (stdout) with --all") {
			t.Errorf("Expected --all/stdout conflict error, got: %v", err)
		}
	})

	t.Run("all flag with empty output should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--licenses", "--org", "test-org", "--all"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using --all with empty output (stdout)")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot use --output '-' or omit --output (stdout) with --all") {
			t.Errorf("Expected --all/empty output conflict error, got: %v", err)
		}
	})

	t.Run("access with all should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "--access", "--all", "--output", "default"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using --access with --all")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot use --all with --access") {
			t.Errorf("Expected --access/--all conflict error, got: %v", err)
		}
	})
}
