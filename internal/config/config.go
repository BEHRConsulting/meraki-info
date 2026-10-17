// Package config turns the meraki-info command line and environment into a
// validated [Config].
//
// Options are Go flag-package flags and must come before the single
// positional command (access, alerting, down, licenses, route-tables);
// the flag package stops parsing at the first non-flag argument, so an
// option written after the command is treated as a second command and
// rejected. Each of -org, -network and -apikey falls back to an
// environment variable (MERAKI_ORG, MERAKI_NET, MERAKI_APIKEY) so the key
// need not appear on the command line.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds every option the application runs with, after flags and
// environment variables have been merged and validated.
type Config struct {
	// Organization is the Meraki organization ID or name from -org or
	// MERAKI_ORG. main resolves a name to its ID before any command runs.
	Organization string
	// Network is the Meraki network ID or name from -network or MERAKI_NET.
	// Empty means "all networks" (see InfoAll).
	Network string
	// APIKey is the Meraki Dashboard API key from -apikey or MERAKI_APIKEY.
	APIKey string
	// OutputFile is the -output path. Empty or "-" means stdout.
	OutputFile string
	// OutputType is the -format value: text, json, xml or csv. Any other
	// value falls back to text in output.NewWriter.
	OutputType string
	// LogLevel is the -loglevel value: debug, info or error.
	LogLevel string
	// Command is the single positional command, lower-cased. It is one of
	// access, alerting, down, licenses or route-tables.
	Command string // The command argument (access, route-tables, licenses, down)
	// InfoAll is true when every network should be processed. It is set by
	// -all, and also forced on whenever -network is empty and the command
	// is not access.
	InfoAll bool
}

// ParseConfig parses the process's command line and environment into a
// Config.
//
// It is the program's only entry point into this package. On any
// validation error it prints the error and the usage text to stderr and
// exits the process with status 1, so callers never see an invalid Config.
//
// Returns the validated configuration.
func ParseConfig() *Config {
	cfg, err := parseConfigWithValidation()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n\n", err.Error())
		printUsage()
		os.Exit(1)
	}
	return cfg
}

// printUsage writes the custom usage text to stderr.
//
// It replaces the flag package's default usage so the API key's
// environment-variable default is never echoed back to the terminal; it
// only says whether MERAKI_APIKEY is set. It is installed as flag.Usage
// and is also called by ParseConfig after a validation error.
func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] COMMAND\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "OPTIONS:\n")

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

	fmt.Fprintf(os.Stderr, "\nCOMMANDS:\n")
	fmt.Fprintf(os.Stderr, "  access        Show available organizations and networks for the API key\n")
	fmt.Fprintf(os.Stderr, "  alerting      Output all devices that are alerting\n")
	fmt.Fprintf(os.Stderr, "  down          Output all devices that are down/offline\n")
	fmt.Fprintf(os.Stderr, "  licenses      Output license information\n")
	fmt.Fprintf(os.Stderr, "  route-tables  Output route tables\n")
}

// parseConfigWithValidation parses config and returns validation errors (for testing)
//
// It registers the flags on the global flag.CommandLine, parses os.Args,
// reads the single positional command, applies the "no -network means all
// networks" rule and checks the option combinations. Tests replace
// flag.CommandLine and os.Args before calling it.
//
// Returns the Config, or an error describing the first invalid option or
// combination found.
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

	// Get the command from positional arguments (after options)
	args := flag.Args()
	if len(args) == 0 {
		return nil, fmt.Errorf("command is required. Must be one of: access, alerting, down, licenses, route-tables")
	}
	if len(args) > 1 {
		return nil, fmt.Errorf("only one command is allowed, got: %s", strings.Join(args, ", "))
	}

	command := strings.ToLower(args[0])
	switch command {
	case "access", "route-tables", "licenses", "down", "alerting":
		cfg.Command = command
	default:
		return nil, fmt.Errorf("invalid command '%s'. Must be one of: access, alerting, down, licenses, route-tables", args[0])
	}

	// Set InfoAll to true if no network is specified (as per requirements)
	// Exception: access command doesn't use InfoAll
	if cfg.Network == "" && cfg.Command != "access" {
		cfg.InfoAll = true
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
