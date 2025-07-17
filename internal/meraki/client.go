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

// Device represents a Meraki device
type Device struct {
	Serial         string   `json:"serial"`
	Name           string   `json:"name,omitempty"`
	Model          string   `json:"model"`
	NetworkID      string   `json:"networkId"`
	MAC            string   `json:"mac,omitempty"`
	Status         string   `json:"status"`
	LastReportedAt string   `json:"lastReportedAt,omitempty"`
	ProductType    string   `json:"productType,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Address        string   `json:"address,omitempty"`
	Lat            float64  `json:"lat,omitempty"`
	Lng            float64  `json:"lng,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	BeaconIdParams struct {
		UUID  string `json:"uuid,omitempty"`
		Major int    `json:"major,omitempty"`
		Minor int    `json:"minor,omitempty"`
	} `json:"beaconIdParams,omitempty"`
}

// RouteWithNetwork extends the Route struct to include network and organization information
type RouteWithNetwork struct {
	Route
	NetworkID    string `json:"network_id" xml:"NetworkID" csv:"network_id"`
	NetworkName  string `json:"network_name" xml:"NetworkName" csv:"network_name"`
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
}

// DeviceWithNetwork extends the Device struct to include network and organization information
type DeviceWithNetwork struct {
	Device
	NetworkName  string `json:"network_name" xml:"NetworkName" csv:"network_name"`
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
}

// LicenseWithNetwork extends the License struct to include organization information
type LicenseWithNetwork struct {
	License
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
}

// DeviceStatus represents the status information for a device from the organization statuses endpoint
type DeviceStatus struct {
	Serial string `json:"serial"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status"`
}

// NetworkDevices represents devices for a specific network
type NetworkDevices struct {
	Network Network  `json:"network"`
	Devices []Device `json:"devices"`
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
	req.Header.Set("User-Agent", "meraki-info/1.0.0")

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
	routes := make([]Route, 0) // Initialize as empty slice instead of nil slice

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

// getNetworkRoutes fetches all routes for a specific network from multiple sources
func (c *Client) getNetworkRoutes(networkID string) ([]Route, error) {
	var allRoutes []Route

	// Fetch static routes from appliance
	staticRoutes, err := c.getNetworkStaticRoutes(networkID)
	if err != nil {
		slog.Warn("Failed to fetch static routes", "network_id", networkID, "error", err)
		// Don't return error, continue with other route types
	} else {
		allRoutes = append(allRoutes, staticRoutes...)
		slog.Debug("Fetched static routes", "network_id", networkID, "count", len(staticRoutes))
	}

	// Fetch VPN routes if available
	vpnRoutes, err := c.getNetworkVPNRoutes(networkID)
	if err != nil {
		slog.Debug("No VPN routes or error fetching VPN routes", "network_id", networkID, "error", err)
		// VPN routes might not be available for all networks, don't treat as error
	} else {
		allRoutes = append(allRoutes, vpnRoutes...)
		slog.Debug("Fetched VPN routes", "network_id", networkID, "count", len(vpnRoutes))
	}

	// Fetch VLAN/L3 interface routes (directly connected subnets)
	vlanRoutes, err := c.getNetworkVLANRoutes(networkID)
	if err != nil {
		slog.Debug("No VLAN routes or error fetching VLAN routes", "network_id", networkID, "error", err)
		// VLAN routes might not be available for all networks, don't treat as error
	} else {
		allRoutes = append(allRoutes, vlanRoutes...)
		slog.Debug("Fetched VLAN routes", "network_id", networkID, "count", len(vlanRoutes))
	}

	// Fetch switch routing information (for Layer 3 switches)
	switchRoutes, err := c.getNetworkSwitchRoutes(networkID)
	if err != nil {
		slog.Debug("No switch routes or error fetching switch routes", "network_id", networkID, "error", err)
		// Switch routes might not be available for all networks, don't treat as error
	} else {
		allRoutes = append(allRoutes, switchRoutes...)
		slog.Debug("Fetched switch routes", "network_id", networkID, "count", len(switchRoutes))
	}

	// Fetch switch stack routing information (for switch stacks with Layer 3 capabilities)
	switchStackRoutes, err := c.getNetworkSwitchStackRoutes(networkID)
	if err != nil {
		slog.Debug("No switch stack routes or error fetching switch stack routes", "network_id", networkID, "error", err)
		// Switch stack routes might not be available for all networks, don't treat as error
	} else {
		allRoutes = append(allRoutes, switchStackRoutes...)
		slog.Debug("Fetched switch stack routes", "network_id", networkID, "count", len(switchStackRoutes))
	}

	return allRoutes, nil
}

// getNetworkStaticRoutes fetches static routes for a specific network
func (c *Client) getNetworkStaticRoutes(networkID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/appliance/staticRoutes", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return nil, fmt.Errorf("failed to decode static routes response: %w", err)
	}

	// Mark these as static routes
	for i := range routes {
		if routes[i].Name == "" {
			routes[i].Name = fmt.Sprintf("Static Route %d", i+1)
		}
	}

	return routes, nil
}

// getNetworkVPNRoutes fetches VPN routes for a specific network
func (c *Client) getNetworkVPNRoutes(networkID string) ([]Route, error) {
	// Try to fetch site-to-site VPN routes
	endpoint := fmt.Sprintf("/networks/%s/appliance/vpn/siteToSiteVpn", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		// VPN might not be configured, return empty slice
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var vpnConfig struct {
		Mode    string `json:"mode"`
		Subnets []struct {
			LocalSubnet string `json:"localSubnet"`
			UseVpn      bool   `json:"useVpn"`
		} `json:"subnets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&vpnConfig); err != nil {
		return []Route{}, nil // Not an error, just no VPN routes
	}

	var routes []Route
	// Parse VPN subnets regardless of mode
	for i, subnet := range vpnConfig.Subnets {
		if subnet.UseVpn {
			routes = append(routes, Route{
				ID:      fmt.Sprintf("vpn-%d", i),
				Name:    fmt.Sprintf("VPN Route %d", i+1),
				Subnet:  subnet.LocalSubnet,
				Enabled: true, // VPN routes are enabled if useVpn is true
			})
		}
	}

	return routes, nil
}

// getNetworkVLANRoutes fetches VLAN/L3 interface routes (directly connected subnets)
func (c *Client) getNetworkVLANRoutes(networkID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/appliance/vlans", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		// VLANs might not be configured, return empty slice
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var vlans []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		ApplianceIP  string `json:"applianceIp"`
		Subnet       string `json:"subnet"`
		InterfaceID  string `json:"interfaceId"`
		DHCPHandling string `json:"dhcpHandling"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&vlans); err != nil {
		return []Route{}, nil // Not an error, just no VLAN routes
	}

	var routes []Route
	for _, vlan := range vlans {
		if vlan.Subnet != "" {
			routes = append(routes, Route{
				ID:        fmt.Sprintf("vlan-%d", vlan.ID),
				Name:      fmt.Sprintf("VLAN %d - %s", vlan.ID, vlan.Name),
				Subnet:    vlan.Subnet,
				GatewayIP: vlan.ApplianceIP,
				Enabled:   true, // VLAN interfaces are enabled by default
			})
		}
	}

	return routes, nil
}

// getNetworkSwitchRoutes attempts to fetch switch routing information
func (c *Client) getNetworkSwitchRoutes(networkID string) ([]Route, error) {
	// First try to get switch routing interfaces
	switchInterfaces, err := c.getNetworkSwitchInterfaces(networkID)
	if err != nil {
		slog.Debug("No switch routing interfaces available", "network_id", networkID, "error", err)
		return []Route{}, nil
	}

	// Then try to get switch static routes
	switchStaticRoutes, err := c.getNetworkSwitchStaticRoutes(networkID)
	if err != nil {
		slog.Debug("No switch static routes available", "network_id", networkID, "error", err)
	}

	// Combine both types
	var allRoutes []Route
	allRoutes = append(allRoutes, switchInterfaces...)
	allRoutes = append(allRoutes, switchStaticRoutes...)

	return allRoutes, nil
}

// getNetworkSwitchInterfaces attempts to fetch switch Layer 3 interfaces
func (c *Client) getNetworkSwitchInterfaces(networkID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/routing/interfaces", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var interfaces []struct {
		InterfaceID string `json:"interfaceId"`
		Name        string `json:"name"`
		Subnet      string `json:"subnet"`
		InterfaceIP string `json:"interfaceIp"`
		VLAN        int    `json:"vlan"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&interfaces); err != nil {
		return []Route{}, nil
	}

	var routes []Route
	for _, iface := range interfaces {
		if iface.Subnet != "" {
			routes = append(routes, Route{
				ID:        fmt.Sprintf("switch-iface-%s", iface.InterfaceID),
				Name:      fmt.Sprintf("Switch Interface - %s", iface.Name),
				Subnet:    iface.Subnet,
				GatewayIP: iface.InterfaceIP,
				Enabled:   true,
			})
		}
	}

	return routes, nil
}

// getNetworkSwitchStaticRoutes attempts to fetch switch static routes
func (c *Client) getNetworkSwitchStaticRoutes(networkID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/routing/staticRoutes", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return []Route{}, nil
	}

	// Mark these as switch static routes
	for i := range routes {
		if routes[i].Name == "" {
			routes[i].Name = fmt.Sprintf("Switch Static Route %d", i+1)
		}
	}

	return routes, nil
}

// getNetworkSwitchStackRoutes attempts to fetch routing information from switch stacks
func (c *Client) getNetworkSwitchStackRoutes(networkID string) ([]Route, error) {
	// First get all switch stacks in the network
	stacks, err := c.getNetworkSwitchStacks(networkID)
	if err != nil {
		slog.Debug("No switch stacks found", "network_id", networkID, "error", err)
		return []Route{}, nil
	}

	var allRoutes []Route

	// For each stack, get routing interfaces and static routes
	for _, stack := range stacks {
		// Get routing interfaces for this stack
		stackInterfaces, err := c.getSwitchStackRoutingInterfaces(networkID, stack.ID)
		if err != nil {
			slog.Debug("Failed to get routing interfaces for stack", "network_id", networkID, "stack_id", stack.ID, "error", err)
		} else {
			allRoutes = append(allRoutes, stackInterfaces...)
		}

		// Get static routes for this stack
		stackStaticRoutes, err := c.getSwitchStackStaticRoutes(networkID, stack.ID)
		if err != nil {
			slog.Debug("Failed to get static routes for stack", "network_id", networkID, "stack_id", stack.ID, "error", err)
		} else {
			allRoutes = append(allRoutes, stackStaticRoutes...)
		}

		// Get DHCP information for this stack
		stackDHCPRoutes, err := c.getSwitchStackDHCPRoutes(networkID, stack.ID)
		if err != nil {
			slog.Debug("Failed to get DHCP routes for stack", "network_id", networkID, "stack_id", stack.ID, "error", err)
		} else {
			allRoutes = append(allRoutes, stackDHCPRoutes...)
		}
	}

	return allRoutes, nil
}

// SwitchStack represents a switch stack
type SwitchStack struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// getNetworkSwitchStacks gets all switch stacks in a network
func (c *Client) getNetworkSwitchStacks(networkID string) ([]SwitchStack, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/stacks", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []SwitchStack{}, nil
	}
	defer resp.Body.Close()

	var stacks []SwitchStack
	if err := json.NewDecoder(resp.Body).Decode(&stacks); err != nil {
		return []SwitchStack{}, nil
	}

	return stacks, nil
}

// getSwitchStackRoutingInterfaces gets routing interfaces for a specific switch stack
func (c *Client) getSwitchStackRoutingInterfaces(networkID, stackID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/stacks/%s/routing/interfaces", networkID, stackID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var interfaces []struct {
		InterfaceID string `json:"interfaceId"`
		Name        string `json:"name"`
		Subnet      string `json:"subnet"`
		InterfaceIP string `json:"interfaceIp"`
		VLAN        int    `json:"vlan"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&interfaces); err != nil {
		return []Route{}, nil
	}

	var routes []Route
	for _, iface := range interfaces {
		if iface.Subnet != "" {
			routes = append(routes, Route{
				ID:        fmt.Sprintf("stack-%s-iface-%s", stackID, iface.InterfaceID),
				Name:      fmt.Sprintf("Stack Interface - %s", iface.Name),
				Subnet:    iface.Subnet,
				GatewayIP: iface.InterfaceIP,
				Enabled:   true,
			})
		}
	}

	return routes, nil
}

// getSwitchStackStaticRoutes gets static routes for a specific switch stack
func (c *Client) getSwitchStackStaticRoutes(networkID, stackID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/stacks/%s/routing/staticRoutes", networkID, stackID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var apiRoutes []struct {
		StaticRouteID               string `json:"staticRouteId"`
		Name                        string `json:"name"`
		Subnet                      string `json:"subnet"`
		NextHopIP                   string `json:"nextHopIp"`
		AdvertiseViaOspfEnabled     bool   `json:"advertiseViaOspfEnabled"`
		PreferOverOspfRoutesEnabled bool   `json:"preferOverOspfRoutesEnabled"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiRoutes); err != nil {
		return []Route{}, nil
	}

	var routes []Route
	for _, apiRoute := range apiRoutes {
		route := Route{
			ID:        apiRoute.StaticRouteID,
			Name:      apiRoute.Name,
			Subnet:    apiRoute.Subnet,
			GatewayIP: apiRoute.NextHopIP,
			Enabled:   apiRoute.PreferOverOspfRoutesEnabled, // Use this as a proxy for enabled status
		}

		if route.Name == "" {
			route.Name = fmt.Sprintf("Stack %s Static Route", stackID)
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// getSwitchStackDHCPRoutes gets DHCP subnet information for a specific switch stack
func (c *Client) getSwitchStackDHCPRoutes(networkID, stackID string) ([]Route, error) {
	endpoint := fmt.Sprintf("/networks/%s/switch/stacks/%s/routing/dhcp", networkID, stackID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return []Route{}, nil
	}
	defer resp.Body.Close()

	var dhcpConfig struct {
		DHCPRelayServers []string `json:"dhcpRelayServerIps"`
		DHCPOptions      []struct {
			Code  string `json:"code"`
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"dhcpOptions"`
		ReservedIPRanges []struct {
			Start   string `json:"start"`
			End     string `json:"end"`
			Comment string `json:"comment"`
		} `json:"reservedIpRanges"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&dhcpConfig); err != nil {
		return []Route{}, nil
	}

	var routes []Route

	// Add DHCP relay servers as routes if they exist
	for i, relayServer := range dhcpConfig.DHCPRelayServers {
		if relayServer != "" {
			routes = append(routes, Route{
				ID:        fmt.Sprintf("stack-%s-dhcp-relay-%d", stackID, i),
				Name:      fmt.Sprintf("Stack %s DHCP Relay %d", stackID, i+1),
				Subnet:    "0.0.0.0/0", // DHCP relay typically forwards all requests
				GatewayIP: relayServer,
				Enabled:   true,
			})
		}
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

// getOrganizationDeviceStatuses fetches device statuses for all devices in an organization
func (c *Client) getOrganizationDeviceStatuses(organizationID string) ([]DeviceStatus, error) {
	endpoint := fmt.Sprintf("/organizations/%s/devices/statuses", organizationID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var deviceStatuses []DeviceStatus
	if err := json.NewDecoder(resp.Body).Decode(&deviceStatuses); err != nil {
		return nil, fmt.Errorf("failed to decode device statuses: %w", err)
	}

	slog.Debug("Retrieved device statuses from organization", "organization_id", organizationID, "status_count", len(deviceStatuses))
	return deviceStatuses, nil
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

// GetDevices fetches device information for the specified organization and network
func (c *Client) GetDevices(organizationID, networkIdentifier string) ([]Device, error) {
	devices := make([]Device, 0) // Initialize as empty slice instead of nil slice

	// If network is specified, get devices for that specific network
	if networkIdentifier != "" {
		// Resolve network name/ID to actual network ID
		networkID, err := c.ResolveNetworkID(organizationID, networkIdentifier)
		if err != nil {
			return nil, err
		}

		networkDevices, err := c.getNetworkDevices(networkID)
		if err != nil {
			return nil, fmt.Errorf("failed to get devices for network %s: %w", networkID, err)
		}
		devices = append(devices, networkDevices...)
		slog.Info("Retrieved devices from specific network", "network_id", networkID, "device_count", len(networkDevices))
	} else {
		// Get all networks in the organization and fetch devices for each
		networks, err := c.getOrganizationNetworks(organizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
		}

		slog.Info("Found networks", "count", len(networks))

		for _, network := range networks {
			networkDevices, err := c.getNetworkDevices(network.ID)
			if err != nil {
				slog.Warn("Failed to get devices for network", "network_id", network.ID, "network_name", network.Name, "error", err)
				continue
			}
			devices = append(devices, networkDevices...)
		}
	}

	slog.Info("Retrieved devices", "count", len(devices))
	return devices, nil
}

// GetDownDevices fetches devices that are currently down/offline
func (c *Client) GetDownDevices(organizationID, networkIdentifier string) ([]Device, error) {
	// Get all devices first
	allDevices, err := c.GetDevices(organizationID, networkIdentifier)
	if err != nil {
		return nil, err
	}

	// Filter for devices that are down/offline
	downDevices := make([]Device, 0) // Initialize as empty slice instead of nil slice
	for _, device := range allDevices {
		// Check if device is offline/down
		// Meraki API typically uses "offline", "alerting", or similar statuses for down devices
		if isDeviceDown(device.Status) {
			downDevices = append(downDevices, device)
		}
	}

	slog.Info("Filtered down devices", "total_devices", len(allDevices), "down_devices", len(downDevices))
	return downDevices, nil
}

// GetAlertingDevices fetches devices that are currently alerting
func (c *Client) GetAlertingDevices(organizationID, networkIdentifier string) ([]Device, error) {
	// Get all devices first
	allDevices, err := c.GetDevices(organizationID, networkIdentifier)
	if err != nil {
		return nil, err
	}

	// Get device statuses from organization endpoint
	deviceStatuses, err := c.getOrganizationDeviceStatuses(organizationID)
	if err != nil {
		slog.Warn("Failed to get device statuses, using basic device info", "error", err)
		// Fall back to using basic device status info (which might be empty)
		alertingDevices := make([]Device, 0)
		for _, device := range allDevices {
			slog.Debug("Device status check (fallback)", "name", device.Name, "status", device.Status, "is_alerting", isDeviceAlerting(device.Status))
			if isDeviceAlerting(device.Status) {
				alertingDevices = append(alertingDevices, device)
			}
		}
		slog.Info("Filtered alerting devices (fallback)", "total_devices", len(allDevices), "alerting_devices", len(alertingDevices))
		return alertingDevices, nil
	}

	// Create a map of device serial to status for quick lookup
	statusMap := make(map[string]string)
	for _, status := range deviceStatuses {
		statusMap[status.Serial] = status.Status
	}

	// Filter for devices that are alerting using actual status data
	alertingDevices := make([]Device, 0)
	for _, device := range allDevices {
		actualStatus, hasStatus := statusMap[device.Serial]
		if hasStatus {
			device.Status = actualStatus // Update the device with actual status
		}

		slog.Debug("Device status check", "name", device.Name, "serial", device.Serial, "status", device.Status, "is_alerting", isDeviceAlerting(device.Status))

		if isDeviceAlerting(device.Status) {
			alertingDevices = append(alertingDevices, device)
		}
	}

	slog.Info("Filtered alerting devices", "total_devices", len(allDevices), "alerting_devices", len(alertingDevices))
	return alertingDevices, nil
}

// getNetworkDevices fetches all devices in a specific network
func (c *Client) getNetworkDevices(networkID string) ([]Device, error) {
	endpoint := fmt.Sprintf("/networks/%s/devices", networkID)

	resp, err := c.makeRequest("GET", endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var devices []Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("failed to decode devices response: %w", err)
	}

	return devices, nil
}

// GetAllNetworkDevices fetches devices for all networks in an organization
func (c *Client) GetAllNetworkDevices(organizationID string) ([]NetworkDevices, error) {
	// Get all networks in the organization
	networks, err := c.getOrganizationNetworks(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get networks for organization %s: %w", organizationID, err)
	}

	slog.Info("Found networks for device backup", "count", len(networks))

	var allNetworkDevices []NetworkDevices

	for _, network := range networks {
		networkDevices, err := c.getNetworkDevices(network.ID)
		if err != nil {
			slog.Warn("Failed to get devices for network", "network_id", network.ID, "network_name", network.Name, "error", err)
			// Still include the network with empty devices for completeness
			allNetworkDevices = append(allNetworkDevices, NetworkDevices{
				Network: network,
				Devices: []Device{},
			})
			continue
		}

		allNetworkDevices = append(allNetworkDevices, NetworkDevices{
			Network: network,
			Devices: networkDevices,
		})

		slog.Info("Retrieved devices for network", "network_id", network.ID, "network_name", network.Name, "device_count", len(networkDevices))
	}

	return allNetworkDevices, nil
}

// isDeviceDown determines if a device is considered down based on its status
func isDeviceDown(status string) bool {
	downStatuses := []string{
		"offline",
		"alerting",
		"dormant",
		"down",
		"unreachable",
		"disconnected",
	}

	statusLower := strings.ToLower(status)
	for _, downStatus := range downStatuses {
		if statusLower == downStatus {
			return true
		}
	}

	return false
}

// isDeviceAlerting determines if a device is considered alerting based on its status
func isDeviceAlerting(status string) bool {
	alertingStatuses := []string{
		"alerting",
	}

	statusLower := strings.ToLower(status)
	for _, alertingStatus := range alertingStatuses {
		if statusLower == alertingStatus {
			return true
		}
	}

	return false
}
