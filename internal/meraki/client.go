// Package meraki provides Meraki API client functionality
package meraki

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Route represents a Meraki network route
type Route struct {
	ID          string      `json:"id,omitempty"`
	Name        string      `json:"name,omitempty"`
	Subnet      string      `json:"subnet"`
	GatewayIP   string      `json:"gatewayIp"`
	GatewayVlan int         `json:"gatewayVlanId,omitempty"`
	Enabled     bool        `json:"enabled"`
	FixedIP     interface{} `json:"fixedIpAssignments,omitempty"`
}

// NetworkRoutes represents routes for a specific network
type NetworkRoutes struct {
	Network Network `json:"network"`
	Routes  []Route `json:"routes"`
}

// License represents a Meraki license
type License struct {
	ID                string `json:"id,omitempty"`
	OrganizationID    string `json:"organizationId,omitempty"`
	DeviceSerial      string `json:"deviceSerial,omitempty"`
	NetworkID         string `json:"networkId,omitempty"`
	State             string `json:"state,omitempty"`
	Edition           string `json:"edition,omitempty"`
	Mode              string `json:"mode,omitempty"`
	ExpirationDate    string `json:"expirationDate,omitempty"`
	LicenseType       string `json:"licenseType,omitempty"`
	LicenseKey        string `json:"licenseKey,omitempty"`
	OrderNumber       string `json:"orderNumber,omitempty"`
	PermanentlyQueued bool   `json:"permanentlyQueued,omitempty"`
	DurationInDays    int    `json:"durationInDays,omitempty"`
}

// NetworkLicenses represents licenses for a specific network
type NetworkLicenses struct {
	Network  Network   `json:"network"`
	Licenses []License `json:"licenses"`
}

// Client represents a Meraki API client
type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// NewClient creates a new Meraki API client
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// For production, you might want to implement proper OAuth2 flow
	// This is a simplified version using API key authentication
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	return &Client{
		httpClient: client,
		baseURL:    "https://api.meraki.com/api/v1",
		apiKey:     apiKey,
	}, nil
}

// NewClientWithOAuth2 creates a new Meraki API client with OAuth2 authentication
func NewClientWithOAuth2(clientID, clientSecret, token string) (*Client, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.meraki.com/oauth/authorize",
			TokenURL: "https://api.meraki.com/oauth/token",
		},
	}

	// Create OAuth2 token
	tok := &oauth2.Token{
		AccessToken: token,
	}

	client := config.Client(context.Background(), tok)

	return &Client{
		httpClient: client,
		baseURL:    "https://api.meraki.com/api/v1",
	}, nil
}

// makeRequest makes an authenticated HTTP request to the Meraki API
func (c *Client) makeRequest(method, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key authentication if available
	if c.apiKey != "" {
		req.Header.Set("X-Cisco-Meraki-API-Key", c.apiKey)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "meraki-routes-backup/1.0.0")

	slog.Debug("Making API request", "method", method, "url", url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	return resp, nil
}

// GetRoutes fetches all routes for the specified organization and network
func (c *Client) GetRoutes(organizationID, networkIdentifier string) ([]Route, error) {
	var routes []Route

	// If network is specified, resolve it to an ID and get routes for that specific network
	if networkIdentifier != "" {
		// Resolve network name/ID to actual network ID
		networkID, err := c.ResolveNetworkID(organizationID, networkIdentifier)
		if err != nil {
			return nil, err
		}

		networkRoutes, err := c.getNetworkRoutes(networkID)
		if err != nil {
			return nil, fmt.Errorf("failed to get routes for network %s: %w", networkID, err)
		}
		routes = append(routes, networkRoutes...)
		slog.Info("Retrieved routes from specific network", "network_id", networkID, "route_count", len(networkRoutes))
	} else {
		// Get all networks in the organization and fetch routes for each
		networks, err := c.getOrganizationNetworks(organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
		}

		slog.Info("Found networks", "count", len(networks))

		for _, network := range networks {
			networkRoutes, err := c.getNetworkRoutes(network.ID)
			if err != nil {
				slog.Warn("Failed to get routes for network", "network_id", network.ID, "network_name", network.Name, "error", err)
				continue
			}
			routes = append(routes, networkRoutes...)
		}
	}

	slog.Info("Retrieved routes", "count", len(routes))
	return routes, nil
}

// Network represents a Meraki network
type Network struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ProductTypes []string `json:"productTypes,omitempty"`
	TimeZone     string   `json:"timeZone,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// Organization represents a Meraki organization
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
	API  struct {
		Enabled bool `json:"enabled"`
	} `json:"api"`
	Licensing struct {
		Model string `json:"model"`
	} `json:"licensing"`
	Cloud struct {
		Region struct {
			Name string `json:"name"`
			Host struct {
				Name string `json:"name"`
			} `json:"host"`
		} `json:"region"`
	} `json:"cloud"`
}

// getOrganizationNetworks fetches all networks in an organization
func (c *Client) getOrganizationNetworks(organizationID string) ([]Network, error) {
	endpoint := fmt.Sprintf("/organizations/%s/networks", organizationID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var networks []Network
	if err := json.NewDecoder(resp.Body).Decode(&networks); err != nil {
		return nil, fmt.Errorf("failed to decode networks response: %w", err)
	}

	return networks, nil
}

// getNetworkRoutes fetches routes for a specific network
func (c *Client) getNetworkRoutes(networkID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/appliance/staticRoutes", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, fmt.Errorf("failed to decode routes response: %w", err)
	}

	return routes, nil
}

// GetOrganizations fetches all organizations accessible with the API key
func (c *Client) GetOrganizations() ([]Organization, error) {
	resp, err := c.makeRequest("GET", "/organizations")
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer resp.Body.Close()

	var organizations []Organization
	if err := json.NewDecoder(resp.Body).Decode(&organizations); err != nil {
		return nil, fmt.Errorf("failed to decode organizations response: %w", err)
	}

	return organizations, nil
}

// GetOrganizationNetworks fetches all networks in an organization (public method)
func (c *Client) GetOrganizationNetworks(organizationID string) ([]Network, error) {
	return c.getOrganizationNetworks(organizationID)
}

// GetAllNetworkRoutes fetches routes for all networks in an organization
func (c *Client) GetAllNetworkRoutes(organizationID string) ([]NetworkRoutes, error) {
	// Get all networks in the organization
	networks, err := c.getOrganizationNetworks(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
	}

	slog.Info("Found networks for backup", "count", len(networks))

	var allNetworkRoutes []NetworkRoutes

	for _, network := range networks {
		networkRoutes, err := c.getNetworkRoutes(network.ID)
		if err != nil {
			slog.Warn("Failed to get routes for network", "network_id", network.ID, "network_name", network.Name, "error", err)
			// Still include the network with empty routes for completeness
			allNetworkRoutes = append(allNetworkRoutes, NetworkRoutes{
				Network: network,
				Routes:  []Route{},
			})
			continue
		}

		allNetworkRoutes = append(allNetworkRoutes, NetworkRoutes{
			Network: network,
			Routes:  networkRoutes,
		})

		slog.Info("Retrieved routes for network", "network_id", network.ID, "network_name", network.Name, "route_count", len(networkRoutes))
	}

	return allNetworkRoutes, nil
}

// ResolveNetworkID resolves a network name or ID to a network ID within an organization
func (c *Client) ResolveNetworkID(organizationID, networkIdentifier string) (string, error) {
	if networkIdentifier == "" {
		return "", nil
	}

	// Get all networks in the organization
	networks, err := c.getOrganizationNetworks(organizationID)
	if err != nil {
		return "", fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
	}

	// First check if it's already a valid network ID
	for _, network := range networks {
		if network.ID == networkIdentifier {
			return networkIdentifier, nil
		}
	}

	// If not found by ID, try to find by name (case-insensitive)
	lowerIdentifier := strings.ToLower(networkIdentifier)
	var matchedNetworks []Network
	for _, network := range networks {
		if strings.ToLower(network.Name) == lowerIdentifier {
			matchedNetworks = append(matchedNetworks, network)
		}
	}

	if len(matchedNetworks) == 0 {
		// Provide helpful error with available networks
		var availableNetworks []string
		for _, network := range networks {
			availableNetworks = append(availableNetworks, fmt.Sprintf("%s (ID: %s)", network.Name, network.ID))
		}
		if len(availableNetworks) > 0 {
			return "", fmt.Errorf("network '%s' not found in organization %s. Available networks: %v", networkIdentifier, organizationID, availableNetworks)
		}
		return "", fmt.Errorf("network '%s' not found in organization %s (no networks found)", networkIdentifier, organizationID)
	}

	if len(matchedNetworks) > 1 {
		// Multiple networks with the same name - provide IDs for disambiguation
		var matchedIDs []string
		for _, network := range matchedNetworks {
			matchedIDs = append(matchedIDs, network.ID)
		}
		return "", fmt.Errorf("multiple networks found with name '%s' in organization %s. Please use network ID instead. Network IDs: %v", networkIdentifier, organizationID, matchedIDs)
	}

	slog.Info("Resolved network name to ID", "name", networkIdentifier, "id", matchedNetworks[0].ID)
	return matchedNetworks[0].ID, nil
}

// ResolveOrganizationID resolves an organization name or ID to an organization ID
func (c *Client) ResolveOrganizationID(organizationIdentifier string) (string, error) {
	if organizationIdentifier == "" {
		return "", nil
	}

	// Get all organizations
	organizations, err := c.GetOrganizations()
	if err != nil {
		return "", fmt.Errorf("failed to get organizations: %w", err)
	}

	// First check if it's already a valid organization ID
	for _, org := range organizations {
		if org.ID == organizationIdentifier {
			return organizationIdentifier, nil
		}
	}

	// If not found by ID, try to find by name (case-insensitive)
	lowerIdentifier := strings.ToLower(organizationIdentifier)
	var matchedOrganizations []Organization
	for _, org := range organizations {
		if strings.ToLower(org.Name) == lowerIdentifier {
			matchedOrganizations = append(matchedOrganizations, org)
		}
	}

	if len(matchedOrganizations) == 0 {
		// Provide helpful error with available organizations
		var availableOrganizations []string
		for _, org := range organizations {
			availableOrganizations = append(availableOrganizations, fmt.Sprintf("%s (ID: %s)", org.Name, org.ID))
		}
		if len(availableOrganizations) > 0 {
			return "", fmt.Errorf("organization '%s' not found. Available organizations: %v", organizationIdentifier, availableOrganizations)
		}
		return "", fmt.Errorf("organization '%s' not found (no organizations found)", organizationIdentifier)
	}

	if len(matchedOrganizations) > 1 {
		// Multiple organizations with the same name - provide IDs for disambiguation
		var duplicateOrgs []string
		for _, org := range matchedOrganizations {
			duplicateOrgs = append(duplicateOrgs, fmt.Sprintf("%s (ID: %s)", org.Name, org.ID))
		}
		return "", fmt.Errorf("multiple organizations found with name '%s': %v. Please use organization ID for disambiguation", organizationIdentifier, duplicateOrgs)
	}

	// Single match found
	slog.Info("Resolved organization name to ID", "name", organizationIdentifier, "id", matchedOrganizations[0].ID)
	return matchedOrganizations[0].ID, nil
}

// GetLicenses fetches license information for the specified organization
func (c *Client) GetLicenses(organizationID string) ([]License, error) {
	endpoint := fmt.Sprintf("/organizations/%s/licenses", organizationID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch licenses: %w", err)
	}
	defer resp.Body.Close()

	var licenses []License
	if err := json.NewDecoder(resp.Body).Decode(&licenses); err != nil {
		return nil, fmt.Errorf("failed to decode licenses response: %w", err)
	}

	return licenses, nil
}

// GetAllNetworkLicenses fetches licenses for all networks in an organization
func (c *Client) GetAllNetworkLicenses(organizationID string) ([]NetworkLicenses, error) {
	// Get all networks in the organization
	networks, err := c.getOrganizationNetworks(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
	}

	slog.Info("Found networks for license backup", "count", len(networks))

	var allNetworkLicenses []NetworkLicenses

	// Get organization-level licenses first
	orgLicenses, err := c.GetLicenses(organizationID)
	if err != nil {
		slog.Warn("Failed to get organization licenses", "org_id", organizationID, "error", err)
		orgLicenses = []License{} // Continue with empty licenses if org licenses fail
	}

	// For each network, collect relevant licenses
	for _, network := range networks {
		// Filter licenses relevant to this network (if they have networkId specified)
		var networkLicenses []License
		for _, license := range orgLicenses {
			// Include license if it's for this network specifically or if it's organization-wide
			if license.NetworkID == network.ID || license.NetworkID == "" {
				networkLicenses = append(networkLicenses, license)
			}
		}

		allNetworkLicenses = append(allNetworkLicenses, NetworkLicenses{
			Network:  network,
			Licenses: networkLicenses,
		})

		slog.Info("Retrieved licenses for network", "network_id", network.ID, "network_name", network.Name, "license_count", len(networkLicenses))
	}

	return allNetworkLicenses, nil
}
