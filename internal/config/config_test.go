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

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "access"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when API key is missing")
		}
		if err != nil && !strings.Contains(err.Error(), "API key is required") {
			t.Errorf("Expected API key error, got: %v", err)
		}
	})

	t.Run("missing command should return error", func(t *testing.T) {
		// Set API key but no command
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options but no command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when no command is provided")
		}
		if err != nil && !strings.Contains(err.Error(), "command is required") {
			t.Errorf("Expected command required error, got: %v", err)
		}
	})

	t.Run("multiple commands should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with multiple commands
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "access", "licenses"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when multiple commands are provided")
		}
		if err != nil && !strings.Contains(err.Error(), "only one command is allowed") {
			t.Errorf("Expected multiple commands error, got: %v", err)
		}
	})

	t.Run("invalid command should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with invalid command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "invalid-command"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when invalid command is provided")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid command") {
			t.Errorf("Expected invalid command error, got: %v", err)
		}
	})

	t.Run("valid access command config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")
		os.Setenv("MERAKI_ORG", "test-org")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "access"}

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
		if cfg.Command != "access" {
			t.Errorf("Expected Command 'access', got '%s'", cfg.Command)
		}
	})

	t.Run("valid licenses command config with options", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "licenses"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if cfg.Command != "licenses" {
			t.Errorf("Expected Command 'licenses', got '%s'", cfg.Command)
		}
		if cfg.Organization != "test-org" {
			t.Errorf("Expected Organization 'test-org', got '%s'", cfg.Organization)
		}
	})

	t.Run("valid route-tables command config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "-format", "json", "route-tables"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if cfg.Command != "route-tables" {
			t.Errorf("Expected Command 'route-tables', got '%s'", cfg.Command)
		}
		if cfg.OutputType != "json" {
			t.Errorf("Expected OutputType 'json', got '%s'", cfg.OutputType)
		}
	})

	t.Run("valid down command config", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "down"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if cfg.Command != "down" {
			t.Errorf("Expected Command 'down', got '%s'", cfg.Command)
		}
	})

	t.Run("all flag with network should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with conflicting options, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "-network", "test-net", "-all", "licenses"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using -all with -network")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot specify -network when using -all") {
			t.Errorf("Expected -all/-network conflict error, got: %v", err)
		}
	})

	t.Run("all flag with stdout output should succeed", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "-all", "-output", "-", "licenses"}

		_, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error when using -all with stdout output, got: %v", err)
		}
	})

	t.Run("all flag with empty output should succeed", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options first, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-org", "test-org", "-all", "licenses"}

		_, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error when using -all with empty output (stdout), got: %v", err)
		}
	})

	t.Run("access with all should return error", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with conflicting options, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-all", "-output", "test.txt", "access"}

		_, err := parseConfigWithValidation()
		if err == nil {
			t.Error("Expected error when using access command with -all")
		}
		if err != nil && !strings.Contains(err.Error(), "cannot use -all with access command") {
			t.Errorf("Expected access/-all conflict error, got: %v", err)
		}
	})

	t.Run("options followed by command format", func(t *testing.T) {
		os.Setenv("MERAKI_APIKEY", "test-key")

		// Reset flags and set test args with options in various orders, then command
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		os.Args = []string{"meraki-info", "-all", "-output", "test.txt", "-org", "test-org", "-format", "json", "licenses"}

		cfg, err := parseConfigWithValidation()
		if err != nil {
			t.Errorf("Expected no error when using [OPTIONS] COMMAND format, got: %v", err)
		}

		if cfg.Command != "licenses" {
			t.Errorf("Expected Command 'licenses', got '%s'", cfg.Command)
		}
		if cfg.OutputFile != "test.txt" {
			t.Errorf("Expected OutputFile 'test.txt', got '%s'", cfg.OutputFile)
		}
		if cfg.OutputType != "json" {
			t.Errorf("Expected OutputType 'json', got '%s'", cfg.OutputType)
		}
		if !cfg.InfoAll {
			t.Error("Expected InfoAll to be true")
		}
	})
}
