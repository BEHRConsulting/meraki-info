package main

import (
	"fmt"
	"log/slog"
	"os"

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

	// Handle commands based on the Command field
	switch cfg.Command {
	case "access":
		showAccessInformation(client, cfg.Organization)
		return

	case "route-tables":
		if cfg.InfoAll {
			err := infoAllNetworkRoutes(client, cfg)
			if err != nil {
				slog.Error("Failed to get info for all network route tables", "error", err)
				os.Exit(1)
			}
		} else {
			err := infoSingleNetworkRoutes(client, cfg)
			if err != nil {
				slog.Error("Failed to get info for route tables", "error", err)
				os.Exit(1)
			}
		}
		return

	case "licenses":
		if cfg.InfoAll {
			err := infoAllNetworkLicenses(client, cfg)
			if err != nil {
				slog.Error("Failed to get info for all network licenses", "error", err)
				os.Exit(1)
			}
		} else {
			err := infoSingleNetworkLicenses(client, cfg)
			if err != nil {
				slog.Error("Failed to collect license info", "error", err)
				os.Exit(1)
			}
		}
		return

	case "down":
		if cfg.InfoAll {
			err := infoAllNetworkDownDevices(client, cfg)
			if err != nil {
				slog.Error("Failed to get info for all network down devices", "error", err)
				os.Exit(1)
			}
		} else {
			err := infoSingleNetworkDownDevices(client, cfg)
			if err != nil {
				slog.Error("Failed to collect down device info", "error", err)
				os.Exit(1)
			}
		}
		return

	case "alerting":
		if cfg.InfoAll {
			if err := infoAllNetworkAlertingDevices(client, cfg); err != nil {
				slog.Error("Failed to get info for all network alerting devices", "error", err)
				os.Exit(1)
			}
		} else {
			if err := infoSingleNetworkAlertingDevices(client, cfg); err != nil {
				slog.Error("Failed to collect alerting device info", "error", err)
				os.Exit(1)
			}
		}
		return

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'. Use access, route-tables, licenses, down, or alerting.\n", cfg.Command)
		os.Exit(1)
	}
}

// infoSingleNetworkRoutes collects routes for a single network
func infoSingleNetworkRoutes(client *meraki.Client, cfg *config.Config) error {
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
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(routes, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Route tables info collection completed successfully", "output_file", outputFile)

	return nil
}

// infoSingleNetworkLicenses collects info for licenses for a single network/organization
func infoSingleNetworkLicenses(client *meraki.Client, cfg *config.Config) error {
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
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(licenses, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Licenses info collection completed successfully", "output_file", outputFile)

	return nil
}

// infoSingleNetworkDownDevices collects info for down devices for a single network
func infoSingleNetworkDownDevices(client *meraki.Client, cfg *config.Config) error {
	// Fetch down devices for single network
	downDevices, err := client.GetDownDevices(cfg.Organization, cfg.Network)
	if err != nil {
		return fmt.Errorf("failed to fetch down devices: %w", err)
	}

	slog.Info("Retrieved down devices", "count", len(downDevices))

	// Determine output filename
	outputFile := cfg.OutputFile
	if outputFile == "" || outputFile == "-" {
		// Send to stdout when not provided or explicitly set to "-"
		outputWriter := output.NewWriter(cfg.OutputType)
		if err := outputWriter.WriteTo(downDevices, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Down devices sent to stdout", "device_count", len(downDevices))
		return nil
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(downDevices, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Down devices info collection completed successfully", "output_file", outputFile)

	return nil
}

// infoSingleNetworkAlertingDevices retrieves and outputs alerting device information for a single network
func infoSingleNetworkAlertingDevices(client *meraki.Client, cfg *config.Config) error {
	// Fetch alerting devices for single network
	alertingDevices, err := client.GetAlertingDevices(cfg.Organization, cfg.Network)
	if err != nil {
		return fmt.Errorf("failed to fetch alerting devices: %w", err)
	}

	slog.Info("Retrieved alerting devices", "count", len(alertingDevices))

	// Determine output filename
	outputFile := cfg.OutputFile
	if outputFile == "" || outputFile == "-" {
		// Send to stdout when not provided or explicitly set to "-"
		outputWriter := output.NewWriter(cfg.OutputType)
		if err := outputWriter.WriteTo(alertingDevices, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Alerting devices sent to stdout", "device_count", len(alertingDevices))
		return nil
	}

	// Output to file
	outputWriter := output.NewWriter(cfg.OutputType)
	if err := outputWriter.WriteToFile(alertingDevices, outputFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}
	slog.Info("Alerting devices info collection completed successfully", "output_file", outputFile)

	return nil
}

// infoAllNetworkAlertingDevices collects info for alerting devices for all networks in the organization(s) to separate files
func infoAllNetworkAlertingDevices(client *meraki.Client, cfg *config.Config) error {
	// Check if output should go to stdout (consolidated format)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		return infoAllNetworkAlertingDevicesConsolidated(client, cfg)
	}

	// Otherwise use separate files for each network
	if cfg.Organization != "" {
		// Get info for all networks in a specific organization
		return infoOrganizationNetworkAlertingDevices(cfg, client, cfg.Organization)
	} else {
		// Get info for all networks in all organizations
		return infoAllOrganizationAlertingDevices(cfg, client)
	}
}

// infoAllNetworkAlertingDevicesConsolidated collects alerting device info for all networks and outputs in a consolidated format to stdout
func infoAllNetworkAlertingDevicesConsolidated(client *meraki.Client, cfg *config.Config) error {
	// Get all organizations
	orgs, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("failed to get organizations: %w", err)
	}

	var allAlertingDevices []meraki.DeviceWithNetwork

	for _, org := range orgs {
		// Get all networks in the organization
		networks, err := client.GetOrganizationNetworks(org.ID)
		if err != nil {
			slog.Error("Failed to get networks for organization", "orgID", org.ID, "orgName", org.Name, "error", err)
			continue
		}

		for _, network := range networks {
			// Get alerting devices for this network
			alertingDevices, err := client.GetAlertingDevices(org.ID, network.ID)
			if err != nil {
				slog.Error("Failed to get alerting devices for network", "networkID", network.ID, "networkName", network.Name, "error", err)
				continue
			}

			// Add network and organization information to each device
			for _, device := range alertingDevices {
				deviceWithNetwork := meraki.DeviceWithNetwork{
					Device:       device,
					NetworkName:  network.Name,
					Organization: org.Name,
				}
				allAlertingDevices = append(allAlertingDevices, deviceWithNetwork)
			}
		}
	}

	slog.Info("Collected all alerting devices", "totalDevices", len(allAlertingDevices))

	// Output to stdout or file
	writer := output.NewWriter(cfg.OutputType)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		if err := writer.WriteTo(allAlertingDevices, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Alerting devices info sent to stdout", "total_devices", len(allAlertingDevices))
	} else {
		if err := writer.WriteToFile(allAlertingDevices, cfg.OutputFile); err != nil {
			return fmt.Errorf("failed to write output to file: %w", err)
		}
		slog.Info("Alerting devices info written to file", "total_devices", len(allAlertingDevices), "file", cfg.OutputFile)
	}

	return nil
}

// infoAllNetworkLicensesConsolidated collects license info for all networks and outputs in a consolidated format to stdout
func infoAllNetworkLicensesConsolidated(client *meraki.Client, cfg *config.Config) error {
	// Get all organizations
	orgs, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("failed to get organizations: %w", err)
	}

	var allLicenses []meraki.LicenseWithNetwork

	for _, org := range orgs {
		// Get licenses for this organization
		licenses, err := client.GetLicenses(org.ID)
		if err != nil {
			slog.Error("Failed to get licenses for organization", "orgID", org.ID, "orgName", org.Name, "error", err)
			continue
		}

		// Add organization information to each license
		for _, license := range licenses {
			licenseWithNetwork := meraki.LicenseWithNetwork{
				License:      license,
				Organization: org.Name,
			}
			allLicenses = append(allLicenses, licenseWithNetwork)
		}
	}

	slog.Info("Collected all licenses", "totalLicenses", len(allLicenses))

	// Output to stdout or file
	writer := output.NewWriter(cfg.OutputType)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		if err := writer.WriteTo(allLicenses, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("License info sent to stdout", "total_licenses", len(allLicenses))
	} else {
		if err := writer.WriteToFile(allLicenses, cfg.OutputFile); err != nil {
			return fmt.Errorf("failed to write output to file: %w", err)
		}
		slog.Info("License info written to file", "total_licenses", len(allLicenses), "file", cfg.OutputFile)
	}

	return nil
}

// infoAllNetworkDownDevicesConsolidated collects down device info for all networks and outputs in a consolidated format to stdout
func infoAllNetworkDownDevicesConsolidated(client *meraki.Client, cfg *config.Config) error {
	// Get all organizations
	orgs, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("failed to get organizations: %w", err)
	}

	var allDownDevices []meraki.DeviceWithNetwork

	for _, org := range orgs {
		// Get all networks in the organization
		networks, err := client.GetOrganizationNetworks(org.ID)
		if err != nil {
			slog.Error("Failed to get networks for organization", "orgID", org.ID, "orgName", org.Name, "error", err)
			continue
		}

		for _, network := range networks {
			// Get down devices for this network
			downDevices, err := client.GetDownDevices(org.ID, network.ID)
			if err != nil {
				slog.Error("Failed to get down devices for network", "networkID", network.ID, "networkName", network.Name, "error", err)
				continue
			}

			// Add network and organization information to each device
			for _, device := range downDevices {
				deviceWithNetwork := meraki.DeviceWithNetwork{
					Device:       device,
					NetworkName:  network.Name,
					Organization: org.Name,
				}
				allDownDevices = append(allDownDevices, deviceWithNetwork)
			}
		}
	}

	slog.Info("Collected all down devices", "totalDevices", len(allDownDevices))

	// Output to stdout or file
	writer := output.NewWriter(cfg.OutputType)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		if err := writer.WriteTo(allDownDevices, os.Stdout); err != nil {
			return fmt.Errorf("failed to write output to stdout: %w", err)
		}
		slog.Info("Down devices info sent to stdout", "total_devices", len(allDownDevices))
	} else {
		if err := writer.WriteToFile(allDownDevices, cfg.OutputFile); err != nil {
			return fmt.Errorf("failed to write output to file: %w", err)
		}
		slog.Info("Down devices info written to file", "total_devices", len(allDownDevices), "file", cfg.OutputFile)
	}

	return nil
}

// infoAllNetworkDownDevices collects info for down devices for all networks in the organization(s)
func infoAllNetworkDownDevices(client *meraki.Client, cfg *config.Config) error {
	// Check if output should go to stdout (consolidated format)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		return infoAllNetworkDownDevicesConsolidated(client, cfg)
	}

	// Otherwise use separate files for each network
	if cfg.Organization != "" {
		// Get info for all networks in a specific organization
		return infoOrganizationNetworkDownDevices(cfg, client, cfg.Organization)
	} else {
		// Get info for all networks in all organizations
		return infoAllOrganizationDownDevices(cfg, client)
	}
}

// infoAllNetworkRoutes collects info for routes for all networks in the organization(s)
func infoAllNetworkRoutes(client *meraki.Client, cfg *config.Config) error {
	// When using --all, always use consolidated output
	return infoAllNetworkRoutesConsolidated(client, cfg)
}

// infoAllNetworkRoutesConsolidated collects info for routes for all networks and outputs to stdout in consolidated format
func infoAllNetworkRoutesConsolidated(client *meraki.Client, cfg *config.Config) error {
	if cfg.Organization != "" {
		// Get routes for all networks in a specific organization
		networkRoutes, err := client.GetAllNetworkRoutes(cfg.Organization)
		if err != nil {
			return fmt.Errorf("failed to fetch network routes: %w", err)
		}

		// Create consolidated output with network information
		allRoutes := make([]meraki.RouteWithNetwork, 0)
		for _, nr := range networkRoutes {
			orgName := cfg.Organization // Could be resolved to name if needed
			for _, route := range nr.Routes {
				allRoutes = append(allRoutes, meraki.RouteWithNetwork{
					Route:        route,
					NetworkID:    nr.Network.ID,
					NetworkName:  nr.Network.Name,
					Organization: orgName,
				})
			}
		}

		// Output to stdout or file
		outputWriter := output.NewWriter(cfg.OutputType)
		if cfg.OutputFile == "" || cfg.OutputFile == "-" {
			if err := outputWriter.WriteTo(allRoutes, os.Stdout); err != nil {
				return fmt.Errorf("failed to write output to stdout: %w", err)
			}
			slog.Info("Route tables info sent to stdout", "total_routes", len(allRoutes))
		} else {
			if err := outputWriter.WriteToFile(allRoutes, cfg.OutputFile); err != nil {
				return fmt.Errorf("failed to write output to file: %w", err)
			}
			slog.Info("Route tables info written to file", "total_routes", len(allRoutes), "file", cfg.OutputFile)
		}
		return nil
	} else {
		// Get routes for all networks in all organizations
		organizations, err := client.GetOrganizations()
		if err != nil {
			return fmt.Errorf("error getting organizations: %w", err)
		}

		allRoutes := make([]meraki.RouteWithNetwork, 0)
		for _, org := range organizations {
			networkRoutes, err := client.GetAllNetworkRoutes(org.ID)
			if err != nil {
				slog.Error("Failed to get routes for organization", "org", org.Name, "error", err)
				continue
			}

			for _, nr := range networkRoutes {
				for _, route := range nr.Routes {
					allRoutes = append(allRoutes, meraki.RouteWithNetwork{
						Route:        route,
						NetworkID:    nr.Network.ID,
						NetworkName:  nr.Network.Name,
						Organization: org.Name,
					})
				}
			}
		}

		// Output to stdout or file
		outputWriter := output.NewWriter(cfg.OutputType)
		if cfg.OutputFile == "" || cfg.OutputFile == "-" {
			if err := outputWriter.WriteTo(allRoutes, os.Stdout); err != nil {
				return fmt.Errorf("failed to write output to stdout: %w", err)
			}
			slog.Info("Route tables info sent to stdout", "total_routes", len(allRoutes))
		} else {
			if err := outputWriter.WriteToFile(allRoutes, cfg.OutputFile); err != nil {
				return fmt.Errorf("failed to write output to file: %w", err)
			}
			slog.Info("Route tables info written to file", "total_routes", len(allRoutes), "file", cfg.OutputFile)
		}
		return nil
	}
}

// infoAllNetworkLicenses collects info for licenses for all networks in the organization(s)
func infoAllNetworkLicenses(client *meraki.Client, cfg *config.Config) error {
	// Check if output should go to stdout (consolidated format)
	if cfg.OutputFile == "" || cfg.OutputFile == "-" {
		return infoAllNetworkLicensesConsolidated(client, cfg)
	}

	// Otherwise use separate files for each network
	if cfg.Organization != "" {
		// Get info for all networks in a specific organization
		return infoOrganizationNetworkLicenses(cfg, client, cfg.Organization)
	} else {
		// Get info for all networks in all organizations
		return infoAllOrganizationLicenses(cfg, client)
	}
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
		fmt.Fprintf(os.Stderr, "❌ Error fetching organizations: %v\n", err)
		os.Exit(1)
	}

	if len(orgs) == 0 {
		fmt.Fprintln(os.Stderr, "⚠️  No organizations found. Please check your API key permissions.")
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
			fmt.Fprintf(os.Stderr, "⚠️  No organization found matching '%s'\n", orgFilter)
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
			fmt.Fprintf(os.Stderr, "  ⚠️  Error fetching networks for %s: %v\n", org.Name, err)
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
			fmt.Printf("   # Get route info from organization '%s':\n", orgs[0].Name)
			fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)

			networks, err := client.GetOrganizationNetworks(orgs[0].ID)
			if err == nil && len(networks) > 0 {
				fmt.Printf("   # Get route info from specific network in '%s':\n", orgs[0].Name)
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --network \"%s\"\n", orgs[0].ID, networks[0].ID)
				fmt.Printf("   # Get info for all networks in '%s' to separate files:\n", orgs[0].Name)
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --all\n", orgs[0].ID)
				fmt.Printf("   # View access info for this organization only:\n")
				fmt.Printf("   ./meraki-routes-backup --access --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)
			}
		} else {
			// General examples when showing all orgs
			fmt.Printf("   # Get route info from a specific organization:\n")
			fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\"\n", orgs[0].ID)

			networks, err := client.GetOrganizationNetworks(orgs[0].ID)
			if err == nil && len(networks) > 0 {
				fmt.Printf("   # Get route info from a specific network:\n")
				fmt.Printf("   ./meraki-routes-backup --apikey \"your-key\" --org \"%s\" --network \"%s\"\n", orgs[0].ID, networks[0].ID)
			}
			fmt.Printf("   # Get info for all networks to separate files:\n")
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

// infoOrganizationNetworkRoutes collects info for routes for all networks in an organization
func infoOrganizationNetworkRoutes(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Only proceed if a specific output file is provided
		if networkCfg.OutputFile == "" {
			return fmt.Errorf("no output file specified for separate file generation")
		}

		err := infoSingleNetworkRoutes(client, &networkCfg)
		if err != nil {
			slog.Error("Failed to collect route info for network", "network", network.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoOrganizationNetworkLicenses collects info for licenses for all networks in an organization
func infoOrganizationNetworkLicenses(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Only proceed if a specific output file is provided
		if networkCfg.OutputFile == "" {
			return fmt.Errorf("no output file specified for separate file generation")
		}

		err := infoSingleNetworkLicenses(client, &networkCfg)
		if err != nil {
			slog.Error("Failed to collect license info for network", "network", network.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoAllOrganizationRoutes collects info for routes for all organizations
func infoAllOrganizationRoutes(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for route tables", "org", org.Name, "id", org.ID)
		err := infoOrganizationNetworkRoutes(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to collect route info for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoAllOrganizationLicenses collects info for licenses for all organizations
func infoAllOrganizationLicenses(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for licenses", "org", org.Name, "id", org.ID)
		err := infoOrganizationNetworkLicenses(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to collect license info for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoOrganizationNetworkDownDevices collects info for down devices for all networks in an organization
func infoOrganizationNetworkDownDevices(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Only proceed if a specific output file is provided
		if networkCfg.OutputFile == "" {
			return fmt.Errorf("no output file specified for separate file generation")
		}

		err := infoSingleNetworkDownDevices(client, &networkCfg)
		if err != nil {
			slog.Error("Failed to collect down device info for network", "network", network.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoAllOrganizationDownDevices collects info for down devices for all organizations
func infoAllOrganizationDownDevices(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for down devices", "org", org.Name, "id", org.ID)
		err := infoOrganizationNetworkDownDevices(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to collect down device info for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}

// infoOrganizationNetworkAlertingDevices collects info for alerting devices for all networks in an organization
func infoOrganizationNetworkAlertingDevices(cfg *config.Config, client *meraki.Client, organizationID string) error {
	networks, err := client.GetOrganizationNetworks(organizationID)
	if err != nil {
		return fmt.Errorf("error getting organization networks: %w", err)
	}

	for _, network := range networks {
		// Create a copy of config for this network
		networkCfg := *cfg
		networkCfg.Organization = organizationID
		networkCfg.Network = network.ID

		// Only proceed if a specific output file is provided
		if networkCfg.OutputFile == "" {
			return fmt.Errorf("no output file specified for separate file generation")
		}

		// Fetch alerting devices for this network
		alertingDevices, err := client.GetAlertingDevices(organizationID, network.ID)
		if err != nil {
			slog.Error("Failed to get alerting devices for network", "network", network.Name, "error", err)
			continue
		}

		if len(alertingDevices) > 0 {
			// Write to file only if there are alerting devices
			outputWriter := output.NewWriter(cfg.OutputType)
			if err := outputWriter.WriteToFile(alertingDevices, networkCfg.OutputFile); err != nil {
				slog.Error("Failed to write alerting devices info", "network", network.Name, "error", err)
				continue
			}
			slog.Info("Alerting devices info collection completed", "network", network.Name, "device_count", len(alertingDevices), "output_file", networkCfg.OutputFile)
		} else {
			slog.Info("No alerting devices found", "network", network.Name)
		}
	}

	return nil
}

// infoAllOrganizationAlertingDevices collects info for alerting devices for all organizations
func infoAllOrganizationAlertingDevices(cfg *config.Config, client *meraki.Client) error {
	organizations, err := client.GetOrganizations()
	if err != nil {
		return fmt.Errorf("error getting organizations: %w", err)
	}

	for _, org := range organizations {
		slog.Info("Processing organization for alerting devices", "org", org.Name, "id", org.ID)
		err := infoOrganizationNetworkAlertingDevices(cfg, client, org.ID)
		if err != nil {
			slog.Error("Failed to collect alerting device info for organization", "org", org.Name, "error", err)
			continue
		}
	}

	return nil
}
