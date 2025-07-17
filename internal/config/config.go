// Package config handles command line arguments and environment variables
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration options for the application
type Config struct {
	Organization string
	Network      string
	APIKey       string
	OutputFile   string
	OutputType   string
	LogLevel     string
	Command      string // The command argument (access, route-tables, licenses, down)
	InfoAll      bool
}

// ParseConfig parses command line arguments and environment variables
func ParseConfig() *Config {
	cfg, err := parseConfigWithValidation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		printUsage()
		os.Exit(1)
	}
	return cfg
}

// printUsage prints custom usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "COMMANDS:\n")
	fmt.Fprintf(os.Stderr, "  access        Show available organizations and networks for the API key\n")
	fmt.Fprintf(os.Stderr, "  alerting      Output all devices that are alerting\n")
	fmt.Fprintf(os.Stderr, "  down          Output all devices that are down/offline\n")
	fmt.Fprintf(os.Stderr, "  licenses      Output license information\n")
	fmt.Fprintf(os.Stderr, "  route-tables  Output route tables\n")
	fmt.Fprintf(os.Stderr, "\nOPTIONS:\n")

	// Manually print each flag, with special handling for apikey
	fmt.Fprintf(os.Stderr, "  -all\n    \tGet info for all networks. If -org specified, get info for all networks in that organization. If -org not specified, get info for all networks in all organizations.\n")

	// Special handling for apikey
	apikeyDescription := "Meraki API key"
	if os.Getenv("MERAKI_APIKEY") != "" {
		apikeyDescription += " (env MERAKI_APIKEY is set)"
	}
	fmt.Fprintf(os.Stderr, "  -apikey string\n    \t%s\n", apikeyDescription)

	fmt.Fprintf(os.Stderr, "  -format string\n    \tOutput format: text, xml, json, csv (default \"text\")\n")
	fmt.Fprintf(os.Stderr, "  -loglevel string\n    \tLog level: debug, info, error (default \"error\")\n")
	fmt.Fprintf(os.Stderr, "  -network string\n    \tMeraki network ID or name\n")
	fmt.Fprintf(os.Stderr, "  -org string\n    \tMeraki organization ID or name\n")
	fmt.Fprintf(os.Stderr, "  -output string\n    \tOutput file path. Use '-' or omit for stdout\n")
}

// parseConfigWithValidation parses config and returns validation errors (for testing)
func parseConfigWithValidation() (*Config, error) {
	cfg := &Config{}

	// Define command line flags (options only, not commands)
	flag.StringVar(&cfg.Organization, "org", os.Getenv("MERAKI_ORG"), "Meraki organization ID or name")
	flag.StringVar(&cfg.Network, "network", os.Getenv("MERAKI_NET"), "Meraki network ID or name")

	// Special handling for apikey to not show default in usage
	apikeyDefault := os.Getenv("MERAKI_APIKEY")
	flag.StringVar(&cfg.APIKey, "apikey", apikeyDefault, "Meraki API key")

	flag.StringVar(&cfg.OutputFile, "output", "", "Output file path. Use '-' or omit for stdout")
	flag.StringVar(&cfg.OutputType, "format", "text", "Output format: text, xml, json, csv")
	flag.StringVar(&cfg.LogLevel, "loglevel", "error", "Log level: debug, info, error")
	flag.BoolVar(&cfg.InfoAll, "all", false, "Get info for all networks. If -org specified, get info for all networks in that organization. If -org not specified, get info for all networks in all organizations.")

	// Custom usage function
	flag.Usage = printUsage

	flag.Parse()

	// Get the command from positional arguments
	args := flag.Args()
	if len(args) == 0 {
		return nil, fmt.Errorf("one of the arguments access, route-tables, licenses, down, or alerting is required")
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("only one command argument is allowed, got: %s", strings.Join(args, ", "))
	}

	command := strings.ToLower(args[0])
	switch command {
	case "access", "route-tables", "licenses", "down", "alerting":
		cfg.Command = command
	default:
		return nil, fmt.Errorf("invalid command '%s'. Must be one of: access, route-tables, licenses, down, alerting", args[0])
	}

	// Validate required fields
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required. Use -apikey flag or MERAKI_APIKEY environment variable")
	}

	// If showing access or using --all, organization is not required
	// For other commands without --all, organization is required
	if cfg.Command != "access" && !cfg.InfoAll && cfg.Organization == "" {
		return nil, fmt.Errorf("organization is required when not using --all or access command. Use --org flag or MERAKI_ORG environment variable")
	}

	// If using -all, network should not be specified
	if cfg.InfoAll && cfg.Network != "" {
		return nil, fmt.Errorf("cannot specify -network when using -all. The -all flag processes all networks in the organization")
	}

	// Note: -all with stdout is now supported for consolidated output with network information
	// The validation requiring -output default for -all has been removed to support this use case

	// Access mode doesn't support -all
	if cfg.Command == "access" && cfg.InfoAll {
		return nil, fmt.Errorf("cannot use -all with access command. Use access command alone to show organizations/networks")
	}

	return cfg, nil
}
