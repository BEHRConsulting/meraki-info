package main

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"
	"time"

	"meraki-info/internal/config"
	"meraki-info/internal/logger"
	"meraki-info/internal/meraki"
	"meraki-info/internal/output"
)

func main() {
	// Parse command line flags and environment variables
	cfg := config.ParseConfig()

	// Initialize logger
	logger.InitLogger(cfg.LogLevel)

	slog.Info("Starting Meraki Info", "version", "1.0.0")

	// Create Meraki client
	client, err := meraki.NewClient(cfg.APIKey)
	if err != nil {
		slog.Error("Failed to create Meraki client", "error", err)
		os.Exit(1)
	}

	// Resolve organization name to ID if needed
	if cfg.Organization != "" {
		resolvedOrgID, err := client.ResolveOrganizationID(cfg.Organization)
		if err != nil {
			slog.Error("Failed to resolve organization", "org", cfg.Organization, "error", err)
			os.Exit(1)
		}
		cfg.Organization = resolvedOrgID
	}

	// If --access flag is provided, show available organizations and networks
	if cfg.ShowAccess {
		showAccessInformation(client, cfg.Organization)
		return
	}

	// If --route-tables flag is provided, handle route table output
	if cfg.ShowRouteTables {
		if cfg.BackupAll {
			err := backupAllNetworkRoutes(client, cfg)
			if err != nil {
				slog.Error("Failed to backup all network route tables", "error", err)
				os.Exit(1)
			}
		} else {
			err := backupSingleNetworkRoutes(client, cfg)
			if err != nil {
				slog.Error("Failed to backup route tables", "error", err)
				os.Exit(1)
			}
		}
		return
	}

	// If --licenses flag is provided, handle license output
	if cfg.ShowLicenses {
		if cfg.BackupAll {
			err := backupAllNetworkLicenses(client, cfg)
			if err != nil {
				slog.Error("Failed to backup all network licenses", "error", err)
				os.Exit(1)
			}
		} else {
			err := backupSingleNetworkLicenses(client, cfg)
			if err != nil {
				slog.Error("Failed to backup licenses", "error", err)
				os.Exit(1)
			}
		}
		return
	}
}

// backupSingleNetworkRoutes backs up routes for a single network
func backupSingleNetworkRoutes(client *meraki.Client, cfg *config.Config) error {
	// Fetch routes for single network
	routes, err := client.GetRoutes(cfg.Organization, cfg.Network)
	if err != nil {
		return fmt.Errorf("failed to fetch routes: %w", err)
	}

	slog.Info("Retrieved routes", "count", len(routes))

	// Determine output filename
	outputFile := cfg.OutputFile
	if outputFile == "" || outputFile == "-" {
		// Send to stdout when not provided or explicitly set to "-"
		outputWriter := output.NewWriter(cfg.OutputType)
		if err := outputWriter.WriteTo(routes, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Route tables sent to stdout", "route_count", len(routes))
		return nil
	} else if outputFile == "default" {
		// Generate default filename for route tables
		defaultFile, err := generateDefaultRouteTablesFilename(client, cfg.Organization, cfg.Network, cfg.OutputType)
		if err != nil {
			return fmt.Errorf("failed to generate default filename: %w", err)
		}
		outputFile = defaultFile
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(routes, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Route tables backup completed successfully", "output_file", outputFile)

	return nil
}

// backupSingleNetworkLicenses backs up licenses for a single network/organization
func backupSingleNetworkLicenses(client *meraki.Client, cfg *config.Config) error {
	// Fetch licenses for the organization
	licenses, err := client.GetLicenses(cfg.Organization)
	if err != nil {
		return fmt.Errorf("failed to fetch licenses: %w", err)
	}

	slog.Info("Retrieved licenses", "count", len(licenses))

	// Determine output filename
	outputFile := cfg.OutputFile
	if outputFile == "" || outputFile == "-" {
		// Send to stdout when not provided or explicitly set to "-"
		outputWriter := output.NewWriter(cfg.OutputType)
		if err := outputWriter.WriteTo(licenses, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Licenses sent to stdout", "license_count", len(licenses))
		return nil
	} else if outputFile == "default" {
		// Generate default filename for licenses
		defaultFile, err := generateDefaultLicensesFilename(client, cfg.Organization, cfg.Network, cfg.OutputType)
		if err != nil {
			return fmt.Errorf("failed to generate default filename: %w", err)
		}
		outputFile = defaultFile
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(licenses, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Licenses backup completed successfully", "output_file", outputFile)

	return nil
}

// backupAllNetworkRoutes backs up routes for all networks in the organization(s) to separate files
func backupAllNetworkRoutes(client *meraki.Client, cfg *config.Config) error {
	if cfg.Organization != "" {
		// Backup all networks in a specific organization
		return backupOrganizationNetworkRoutes(cfg, client, cfg.Organization)
	} else {
		// Backup all networks in all organizations
		return backupAllOrganizationRoutes(cfg, client)
	}
}

// backupAllNetworkLicenses backs up licenses for all networks in the organization(s) to separate files
func backupAllNetworkLicenses(client *meraki.Client, cfg *config.Config) error {
	if cfg.Organization != "" {
		// Backup all networks in a specific organization
		return backupOrganizationNetworkLicenses(cfg, client, cfg.Organization)
	} else {
		// Backup all networks in all organizations
		return backupAllOrganizationLicenses(cfg, client)
	}
}

// getFileExtension returns the appropriate file extension for the output type
func getFileExtension(outputType string) string {
	switch strings.ToLower(outputType) {
	case "json":
		return ".json"
	case "xml":
		return ".xml"
	case "csv":
		return ".csv"
	default:
		return ".txt"
	}
}

// sanitizeFilename removes or replaces characters that are not safe for filenames
func sanitizeFilename(input string) string {
	// Replace spaces and special characters with underscores
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	sanitized := reg.ReplaceAllString(input, "_")

	// Remove multiple consecutive underscores
	reg2 := regexp.MustCompile(`_{2,}`)
	sanitized = reg2.ReplaceAllString(sanitized, "_")

	// Trim underscores from beginning and end
	sanitized = strings.Trim(sanitized, "_")

	// Ensure it's not empty
	if sanitized == "" {
		sanitized = "unknown"
	}

	return sanitized
}

// generateDefaultRouteTablesFilename generates a default filename for route tables
func generateDefaultRouteTablesFilename(client *meraki.Client, organizationID, networkIdentifier, outputType string) (string, error) {
	return generateFilenameWithPrefix(client, "RouteTables", organizationID, networkIdentifier, outputType)
}

// generateDefaultLicensesFilename generates a default filename for licenses
func generateDefaultLicensesFilename(client *meraki.Client, organizationID, networkIdentifier, outputType string) (string, error) {
	return generateFilenameWithPrefix(client, "Licenses", organizationID, networkIdentifier, outputType)
}

// generateFilenameWithPrefix generates a filename with a given prefix in the format: <prefix>-<org>-<network>-<RFC3339 datetime>.ext
func generateFilenameWithPrefix(client *meraki.Client, prefix, organizationID, networkIdentifier, outputType string) (string, error) {
	// Get organization name
	orgs, err := client.GetOrganizations()
	if err != nil {
		return "", fmt.Errorf("failed to get organizations: %w", err)
	}

	var orgName string
	for _, org := range orgs {
		if org.ID == organizationID {
			orgName = org.Name
			break
		}
	}
	if orgName == "" {
		orgName = organizationID // fallback to ID if name not found
	}

	// Get network name if networkIdentifier is provided
	networkName := ""
	if networkIdentifier != "" {
		// Resolve network identifier to get network info
		networks, err := client.GetOrganizationNetworks(organizationID)
		if err != nil {
			return "", fmt.Errorf("failed to get networks: %w", err)
		}

		// Find the network by ID or name
		for _, network := range networks {
			if network.ID == networkIdentifier || network.Name == networkIdentifier {
				networkName = network.Name
				break
			}
		}

		if networkName == "" {
			networkName = networkIdentifier // fallback to identifier if not found
		}
	} else {
		networkName = "AllNetworks"
	}

	// Generate RFC3339-style timestamp (filename-safe version)
	timestamp := time.Now().Format("2006-01-02T15-04-05Z07-00")

	// Sanitize names for filename
	sanitizedOrgName := sanitizeFilename(orgName)
	sanitizedNetworkName := sanitizeFilename(networkName)

	// Get file extension based on output type, default to .txt
	ext := getFileExtension(outputType)
	if ext == "" {
		ext = ".txt"
	}

	// Generate filename: <prefix>-<org>-<network>-<RFC3339 datetime>.ext
	filename := fmt.Sprintf("%s-%s-%s-%s%s",
		prefix,
		sanitizedOrgName,
		sanitizedNetworkName,
		timestamp,
		ext)

	return filename, nil
}

// showAccessInformation displays available organizations and networks for the API key
func showAccessInformation(client *meraki.Client, orgFilter string) {
	fmt.Println("╔════════════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                          Meraki API Access Information                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get organizations
	orgs, err := client.GetOrganizations()
	if err != nil {
		fmt.Printf("❌ Error fetching organizations: %v\n", err)
		os.Exit(1)
	}

	if len(orgs) == 0 {
		fmt.Println("⚠️  No organizations found. Please check your API key permissions.")
		return
	}

	// Filter organizations if orgFilter is provided
	var filteredOrgs []meraki.Organization
	if orgFilter != "" {
		for _, org := range orgs {
			if org.ID == orgFilter || org.Name == orgFilter {
				filteredOrgs = append(filteredOrgs, org)
			}
		}
		if len(filteredOrgs) == 0 {
			fmt.Printf("⚠️  No organization found matching '%s'\n", orgFilter)
			fmt.Println("\n📋 Available organizations:")
			for i, org := range orgs {
				fmt.Printf("   %d. %s (ID: %s)\n", i+1, org.Name, org.ID)
			}
			return
		}
		orgs = filteredOrgs
		fmt.Printf("🔍 Filtering by organization: %s\n\n", orgFilter)
	}

	fmt.Printf("✅ Found %d organization(s) accessible with your API key:\n\n", len(orgs))

	for i, org := range orgs {
		fmt.Printf("┌─ Organization %d ─────────────────────────────────────────────────────────\n", i+1)
		fmt.Printf("│ ID:         %s\n", org.ID)
		fmt.Printf("│ Name:       %s\n", org.Name)
		fmt.Printf("│ API:        %s\n", formatBool(org.API.Enabled, "Enabled", "Disabled"))
		fmt.Printf("│ Licensing:  %s\n", org.Licensing.Model)
		fmt.Printf("│ Region:     %s (%s)\n", org.Cloud.Region.Name, org.Cloud.Region.Host.Name)
		if org.URL != "" {
			fmt.Printf("│ Dashboard:  %s\n", org.URL)
		}
		fmt.Printf("└───────────────────────────────────────────────────────────────────────────\n")

		// Get networks for this organization
		networks, err := client.GetOrganizationNetworks(org.ID)
		if err != nil {
			fmt.Printf("  ⚠️  Error fetching networks for %s: %v\n", org.Name, err)
			continue
		}

		if len(networks) == 0 {
			fmt.Println("  📭 No networks found in this organization")
		} else {
			fmt.Printf("  📶 Networks (%d):\n", len(networks))
			for j, network := range networks {
				fmt.Printf("    %d. %s (ID: %s)", j+1, network.Name, network.ID)
				if len(network.ProductTypes) > 0 {
					fmt.Printf(" - Products: %v", network.ProductTypes)
				}
				if network.TimeZone != "" {
					fmt.Printf(" - TZ: %s", network.TimeZone)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}

	fmt.Println("💡 Usage Examples:")
	if len(orgs) > 0 {
		if orgFilter != "" {
			// When filtering by org, show specific examples for that org
			fmt.Printf("   # Backup routes from organization '%s':\n", orgs[0].Name)
			fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)

			networks, err := client.GetOrganizationNetworks(orgs[0].ID)
			if err == nil && len(networks) > 0 {
				fmt.Printf("   # Backup routes from specific network in '%s':\n", orgs[0].Name)
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --network \"%s\"\n", orgs[0].ID, networks[0].ID)
				fmt.Printf("   # Backup all networks in '%s' to separate files:\n", orgs[0].Name)
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --all\n", orgs[0].ID)
				fmt.Printf("   # View access info for this organization only:\n")
				fmt.Printf("   ./meraki-routes-backup --access --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)
			}
		} else {
			// General examples when showing all orgs
			fmt.Printf("   # Backup routes from a specific organization:\n")
			fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)

			networks, err := client.GetOrganizationNetworks(orgs[0].ID)
			if err == nil && len(networks) > 0 {
				fmt.Printf("   # Backup routes from a specific network:\n")
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --network \"%s\"\n", orgs[0].ID, networks[0].ID)
			}
			fmt.Printf("   # Backup all networks to separate files:\n")
			fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --all\n", orgs[0].ID)
			fmt.Printf("   # View access info for specific organization:\n")
			fmt.Printf("   ./meraki-routes-backup --access --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)
		}
	}
	fmt.Println()
}

// formatBool formats a boolean value with custom true/false strings
func formatBool(value bool, trueStr, falseStr string) string {
	if value {
		return trueStr
	}
	return falseStr
}

// backupOrganizationNetworkRoutes backs up routes for all networks in an organization
func backupOrganizationNetworkRoutes(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Generate filename for this network
		if networkCfg.OutputFile == "" || networkCfg.OutputFile == "default" {
			outputFile, err := generateDefaultRouteTablesFilename(client, organizationID, network.ID, cfg.OutputType)
			if err != nil {
				slog.Error("Failed to generate filename for network", "network", network.Name, "error", err)
				continue
			}
			networkCfg.OutputFile = outputFile
		}

		err := backupSingleNetworkRoutes(client, &networkCfg)
		if err != nil {
			slog.Error("Failed to backup routes for network", "network", network.Name, "error", err)
			continue
		}
	}

	return nil
}

// backupOrganizationNetworkLicenses backs up licenses for all networks in an organization
func backupOrganizationNetworkLicenses(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Generate filename for this network
		if networkCfg.OutputFile == "" || networkCfg.OutputFile == "default" {
			outputFile, err := generateDefaultLicensesFilename(client, organizationID, network.ID, cfg.OutputType)
			if err != nil {
				slog.Error("Failed to generate filename for network", "network", network.Name, "error", err)
				continue
			}
			networkCfg.OutputFile = outputFile
		}

		err := backupSingleNetworkLicenses(client, &networkCfg)
		if err != nil {
			slog.Error("Failed to backup licenses for network", "network", network.Name, "error", err)
			continue
		}
	}

	return nil
}

// backupAllOrganizationRoutes backs up routes for all organizations
func backupAllOrganizationRoutes(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for route tables", "org", org.Name, "id", org.ID)
		err := backupOrganizationNetworkRoutes(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to backup routes for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}

// backupAllOrganizationLicenses backs up licenses for all organizations
func backupAllOrganizationLicenses(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for licenses", "org", org.Name, "id", org.ID)
		err := backupOrganizationNetworkLicenses(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to backup licenses for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}
