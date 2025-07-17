// Package config handles command line arguments and environment variables
package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds all configuration options for the application
type Config struct {
	Organization    string
	Network         string
	APIKey          string
	OutputFile      string
	OutputType      string
	LogLevel        string
	ShowAccess      bool
	BackupAll       bool
	ShowRouteTables bool
	ShowLicenses    bool
	ShowDownDevices bool
}

// ParseConfig parses command line flags and environment variables
func ParseConfig() *Config {
	cfg, err := parseConfigWithValidation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}
	return cfg
}

// parseConfigWithValidation parses config and returns validation errors (for testing)
func parseConfigWithValidation() (*Config, error) {
	cfg := &Config{}

	// Define command line flags
	flag.StringVar(&cfg.Organization, "org", os.Getenv("MERAKI_ORG"), "Meraki organization ID or name")
	flag.StringVar(&cfg.Network, "network", os.Getenv("MERAKI_NET"), "Meraki network ID or name")
	flag.StringVar(&cfg.APIKey, "apikey", os.Getenv("MERAKI_APIKEY"), "Meraki API key")
	flag.StringVar(&cfg.OutputFile, "output", "", "Output file path. Use '-' or omit for stdout, 'default' for default filenames")
	flag.StringVar(&cfg.OutputType, "format", "text", "Output format: text, xml, json, csv")
	flag.StringVar(&cfg.LogLevel, "loglevel", "error", "Log level: debug, info, error")
	flag.BoolVar(&cfg.ShowAccess, "access", false, "Show available organizations and networks for the API key (can be filtered with --org)")
	flag.BoolVar(&cfg.BackupAll, "all", false, "Backup for all networks. If --org specified, backup all networks in that organization. If --org not specified, backup all networks in all organizations.")
	flag.BoolVar(&cfg.ShowRouteTables, "route-tables", false, "Output route tables (default filename: RouteTables-<org>-<network>-<RFC3339 datetime>.txt)")
	flag.BoolVar(&cfg.ShowLicenses, "licenses", false, "Output license information (default filename: Licenses-<org>-<network>-<RFC3339 datetime>.txt)")
	flag.BoolVar(&cfg.ShowDownDevices, "down", false, "Output all devices that are down/offline (default filename: Down-<org>-<network>-<RFC3339 datetime>.txt)")

	flag.Parse()

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required. Use --apikey flag or MERAKI_APIKEY environment variable")
	}

	// Require at least one of the action flags
	actionCount := 0
	if cfg.ShowAccess {
		actionCount++
	}
	if cfg.ShowRouteTables {
		actionCount++
	}
	if cfg.ShowLicenses {
		actionCount++
	}
	if cfg.ShowDownDevices {
		actionCount++
	}
	if actionCount == 0 {
		return nil, fmt.Errorf("one of the parameters --access, --route-tables, --licenses, or --down is required")
	}
	if actionCount > 1 {
		return nil, fmt.Errorf("only one of --access, --route-tables, --licenses, or --down can be specified at a time")
	}

	// If showing access or using --all without --org, organization is not required for some cases
	if !cfg.ShowAccess && !cfg.BackupAll && cfg.Organization == "" {
		return nil, fmt.Errorf("organization is required. Use --org flag or MERAKI_ORG environment variable")
	}

	// If using --all, network should not be specified
	if cfg.BackupAll && cfg.Network != "" {
		return nil, fmt.Errorf("cannot specify --network when using --all. The --all flag processes all networks in the organization")
	}

	// If using --all, stdout output is not supported
	if cfg.BackupAll && (cfg.OutputFile == "-" || cfg.OutputFile == "") {
		return nil, fmt.Errorf("cannot use --output '-' or omit --output (stdout) with --all. The --all flag creates separate files for each network. Use --output 'default' for default filenames")
	}

	// Access mode doesn't support --all
	if cfg.ShowAccess && cfg.BackupAll {
		return nil, fmt.Errorf("cannot use --all with --access. Use --access alone to show organizations/networks")
	}

	return cfg, nil
}
