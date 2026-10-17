// Package meraki provides Meraki API client functionality
//
// It is a small, read-only client for the Cisco Meraki Dashboard API v1
// (https://api.meraki.com/api/v1). Every call is an HTTP GET made through
// [Client.makeRequest], which authenticates with the X-Cisco-Meraki-API-Key
// header and retries rate-limit (429) and server (5xx) failures with
// exponential backoff.
//
// The package also defines the data types the rest of the program prints:
// [Route], [License] and [Device], plus the *WithNetwork variants that add
// the network and organization a record came from for consolidated
// multi-network output.
//
// Route collection is best-effort by design: a network has routes in up to
// five places (appliance static routes, site-to-site VPN subnets, appliance
// VLANs, switch L3 interfaces and switch stacks) and most networks only have
// some of them, so a source that fails or is absent contributes no routes
// rather than an error.
//
// Responses are not paginated: each list endpoint is read as a single page.
package meraki

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Route represents a Meraki network route
//
// Its JSON shape is that of the appliance static-routes endpoint, which
// decodes into it directly. Routes from the other sources (VPN, VLAN,
// switch and switch-stack) are synthesized into the same shape, with an ID
// prefix such as "vpn-", "vlan-" or "stack-" saying where they came from.
type Route struct {
	// ID is the Meraki route ID, or a synthesized ID for non-static sources.
	ID string `json:"id,omitempty"`
	// Name is the route's name; a generated label when the API gave none.
	Name string `json:"name,omitempty"`
	// Subnet is the destination subnet in CIDR notation.
	Subnet string `json:"subnet"`
	// GatewayIP is the next-hop address (or interface address for
	// directly connected subnets).
	GatewayIP string `json:"gatewayIp"`
	// GatewayVlan is the next-hop VLAN ID, when the API supplies one.
	GatewayVlan int `json:"gatewayVlanId,omitempty"`
	// Enabled reports whether the route is active.
	Enabled bool `json:"enabled"`
	// FixedIP holds the API's fixedIpAssignments object unchanged; its
	// shape is not modeled, so it is kept as decoded JSON.
	FixedIP interface{} `json:"fixedIpAssignments,omitempty"`
}

// NetworkRoutes represents routes for a specific network
type NetworkRoutes struct {
	// Network is the network the routes belong to.
	Network Network `json:"network"`
	// Routes holds every route found for Network; empty, not nil, when
	// collection failed.
	Routes []Route `json:"routes"`
}

// License represents a Meraki license
//
// Its JSON shape is that of GET /organizations/{organizationId}/licenses.
type License struct {
	// ID is the Meraki license ID.
	ID string `json:"id,omitempty"`
	// OrganizationID is the organization that owns the license.
	OrganizationID string `json:"organizationId,omitempty"`
	// DeviceSerial is the serial of the device the license is assigned to.
	DeviceSerial string `json:"deviceSerial,omitempty"`
	// NetworkID is the network the license is assigned to, if any.
	NetworkID string `json:"networkId,omitempty"`
	// State is the license state as Meraki reports it (for example active).
	State string `json:"state,omitempty"`
	// Edition is the license edition, when applicable.
	Edition string `json:"edition,omitempty"`
	// Mode is the license mode, when applicable.
	Mode string `json:"mode,omitempty"`
	// ExpirationDate is the expiry timestamp exactly as the API returns it.
	ExpirationDate string `json:"expirationDate,omitempty"`
	// LicenseType is the license type (product/model family).
	LicenseType string `json:"licenseType,omitempty"`
	// LicenseKey is the license key.
	LicenseKey string `json:"licenseKey,omitempty"`
	// OrderNumber is the order the license was purchased on.
	OrderNumber string `json:"orderNumber,omitempty"`
	// PermanentlyQueued reports whether the license is permanently queued.
	PermanentlyQueued bool `json:"permanentlyQueued,omitempty"`
	// DurationInDays is the license term in days.
	DurationInDays int `json:"durationInDays,omitempty"`
}

// NetworkLicenses represents licenses for a specific network
type NetworkLicenses struct {
	// Network is the network the licenses were matched to.
	Network Network `json:"network"`
	// Licenses holds the licenses assigned to Network plus every
	// organization-wide license (those with no NetworkID).
	Licenses []License `json:"licenses"`
}

// Device represents a Meraki device
//
// Its JSON shape is that of GET /networks/{networkId}/devices. Status is
// filled from that response when present; [Client.GetAlertingDevices]
// overwrites it from the organization device-statuses endpoint.
type Device struct {
	// Serial is the device serial number, Meraki's unique device key.
	Serial string `json:"serial"`
	// Name is the device's dashboard name.
	Name string `json:"name,omitempty"`
	// Model is the hardware model, for example MX64.
	Model string `json:"model"`
	// NetworkID is the network the device belongs to.
	NetworkID string `json:"networkId"`
	// MAC is the device's MAC address.
	MAC string `json:"mac,omitempty"`
	// Status is the device status (online, offline, alerting, dormant...).
	Status string `json:"status"`
	// LastReportedAt is when the device last checked in, as the API gives it.
	LastReportedAt string `json:"lastReportedAt,omitempty"`
	// ProductType is the product family, for example appliance or switch.
	ProductType string `json:"productType,omitempty"`
	// Tags are the device's dashboard tags.
	Tags []string `json:"tags,omitempty"`
	// Address is the device's configured street address.
	Address string `json:"address,omitempty"`
	// Lat is the device's configured latitude.
	Lat float64 `json:"lat,omitempty"`
	// Lng is the device's configured longitude.
	Lng float64 `json:"lng,omitempty"`
	// Notes are the device's dashboard notes.
	Notes string `json:"notes,omitempty"`
	// BeaconIdParams holds the device's Bluetooth beacon identifiers. It is
	// decoded but not printed by any writer.
	BeaconIdParams struct {
		// UUID is the beacon UUID.
		UUID string `json:"uuid,omitempty"`
		// Major is the beacon major number.
		Major int `json:"major,omitempty"`
		// Minor is the beacon minor number.
		Minor int `json:"minor,omitempty"`
	} `json:"beaconIdParams,omitempty"`
}

// RouteWithNetwork extends the Route struct to include network and organization information
//
// It is the record type of consolidated (-all to stdout) route output. The
// xml and csv struct tags are not read by any writer in this program; the
// output package builds its XML and CSV by hand.
type RouteWithNetwork struct {
	Route
	// NetworkID is the ID of the network the route came from.
	NetworkID string `json:"network_id" xml:"NetworkID" csv:"network_id"`
	// NetworkName is the name of the network the route came from.
	NetworkName string `json:"network_name" xml:"NetworkName" csv:"network_name"`
	// Organization is the organization the route came from. It holds the
	// organization's name when every organization is listed, but its ID
	// when -org was given (see main.infoAllNetworkRoutesConsolidated).
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
}

// DeviceWithNetwork extends the Device struct to include network and organization information
//
// It is the record type of consolidated down and alerting device output.
type DeviceWithNetwork struct {
	Device
	// NetworkName is the name of the network the device belongs to.
	NetworkName string `json:"network_name" xml:"NetworkName" csv:"network_name"`
	// NetworkID is the ID of the network the device belongs to.
	NetworkID string `json:"network_id" xml:"NetworkID" csv:"network_id"`
	// Organization is the name of the device's organization.
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
	// OrganizationID is the ID of the device's organization.
	OrganizationID string `json:"organization_id" xml:"OrganizationID" csv:"organization_id"`
}

// LicenseWithNetwork extends the License struct to include organization information
//
// It is the record type of consolidated license output. Despite the name it
// carries no network fields: licenses are listed per organization.
type LicenseWithNetwork struct {
	License
	// Organization is the name of the organization that owns the license.
	Organization string `json:"organization" xml:"Organization" csv:"organization"`
	// OrganizationID is the ID of the organization that owns the license.
	OrganizationID string `json:"organization_id" xml:"OrganizationID" csv:"organization_id"`
}

// DeviceStatus represents the status information for a device from the organization statuses endpoint
//
// It is decoded from GET /organizations/{organizationId}/devices/statuses,
// the endpoint that reliably reports live status.
type DeviceStatus struct {
	// Serial is the device serial, used to join statuses to devices.
	Serial string `json:"serial"`
	// Name is the device's dashboard name.
	Name string `json:"name,omitempty"`
	// Status is the live status (online, offline, alerting, dormant...).
	Status string `json:"status"`
}

// NetworkDevices represents devices for a specific network
type NetworkDevices struct {
	// Network is the network the devices belong to.
	Network Network `json:"network"`
	// Devices holds the network's devices; empty, not nil, when fetching
	// them failed.
	Devices []Device `json:"devices"`
}

// RetryConfig holds retry configuration parameters
//
// See [Client.calculateBackoff] for how the fields combine.
type RetryConfig struct {
	// MaxRetries is how many times a request is retried after the first
	// attempt, so a request is tried at most MaxRetries+1 times.
	MaxRetries int
	// InitialInterval is the wait before the first retry.
	InitialInterval time.Duration
	// MaxInterval caps any single wait.
	MaxInterval time.Duration
	// Multiplier is the growth factor between successive waits.
	Multiplier float64
}

// DefaultRetryConfig returns a default retry configuration
//
// Returns 3 retries, starting at 1s, doubling, capped at 30s. Because
// [Client.calculateBackoff] maps attempts 0 and 1 to the same wait, the
// three waits a failing request actually sees are 1s, 1s and 2s.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialInterval: 1 * time.Second,
		MaxInterval:     30 * time.Second,
		Multiplier:      2.0,
	}
}

// Client represents a Meraki API client
//
// A Client is created by [NewClient] (API key) or [NewClientWithOAuth2]. It
// holds no per-request state; it is not documented as safe for concurrent
// use, and the program only ever uses it from one goroutine.
type Client struct {
	// httpClient sends the requests; it carries the timeout and TLS policy.
	httpClient *http.Client
	// baseURL is the API root, https://api.meraki.com/api/v1 in production.
	// Tests in this package point it at an httptest server.
	baseURL string
	// apiKey is sent as X-Cisco-Meraki-API-Key; empty for OAuth2 clients.
	apiKey string
	// retryConfig controls makeRequest's retry and backoff behavior.
	retryConfig RetryConfig
}

// NewClient creates a new Meraki API client
//
// The client talks to the global https://api.meraki.com/api/v1 endpoint
// with a 30-second per-request timeout, TLS 1.2 or later, and
// [DefaultRetryConfig].
//
// Parameters:
//   - apiKey: the Meraki Dashboard API key.
//
// Returns the client, or an error if apiKey is empty. No request is made,
// so a wrong key is only discovered on the first call.
func NewClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key cannot be empty")
	}

	// For production, you might want to implement proper OAuth2 flow
	// This is a simplified version using API key authentication
	client := &http.Client{
		Timeout: time.Second * 30,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				// Could add certificate pinning for extra security
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &Client{
		httpClient:  client,
		baseURL:     "https://api.meraki.com/api/v1",
		apiKey:      apiKey,
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// NewClientWithOAuth2 creates a new Meraki API client with OAuth2 authentication
//
// It wraps an already-issued access token in an oauth2 HTTP client; it
// does not run an authorization flow or refresh the token. Nothing in the
// program calls it today: the CLI only accepts an API key.
//
// Parameters:
//   - clientID, clientSecret: the OAuth2 application credentials.
//   - token: an access token obtained elsewhere.
//
// Returns the client. The error result is always nil.
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
		httpClient:  client,
		baseURL:     "https://api.meraki.com/api/v1",
		retryConfig: DefaultRetryConfig(),
	}, nil
}

// isRetryableError determines if an error should be retried
//
// Parameters:
//   - err: the transport error from the HTTP client, or nil.
//   - statusCode: the HTTP status, consulted only when err is nil.
//
// Returns true for any transport error (timeouts, resets, DNS) and for
// 429, 500, 502, 503 and 504; false for everything else, including other
// 4xx codes, which retrying cannot fix.
func isRetryableError(err error, statusCode int) bool {
	// Retry on network errors
	if err != nil {
		return true
	}

	// Retry on specific HTTP status codes
	switch statusCode {
	case 429: // Too Many Requests (rate limiting)
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	default:
		return false
	}
}

// calculateBackoff calculates the backoff duration for a retry attempt
//
// The wait is InitialInterval * Multiplier^(attempt-1), capped at
// MaxInterval. Attempts 0 and 1 both wait InitialInterval. No jitter is
// applied, despite the comment in the body.
//
// Parameters:
//   - attempt: the zero-based number of the attempt that just failed.
//
// Returns how long to sleep before the next attempt.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return c.retryConfig.InitialInterval
	}

	// Exponential backoff with jitter
	backoff := time.Duration(float64(c.retryConfig.InitialInterval) * math.Pow(c.retryConfig.Multiplier, float64(attempt-1)))

	// Cap at maximum interval
	if backoff > c.retryConfig.MaxInterval {
		backoff = c.retryConfig.MaxInterval
	}

	return backoff
}

// makeRequest makes an authenticated HTTP request to the Meraki API with retry logic
//
// Every API call in this package goes through it. A request is retried,
// after the [Client.calculateBackoff] wait, while [isRetryableError] says
// so and retries remain; any other non-2xx status fails at once. The
// response body of a failed attempt is closed and not read, so Meraki's
// error message is not included in the returned error.
//
// Parameters:
//   - method: the HTTP method; always "GET" in this program.
//   - endpoint: the path below baseURL, starting with "/".
//
// Returns the successful (2xx) response, whose Body the caller must close,
// or an error after the last failed attempt.
func (c *Client) makeRequest(method, endpoint string) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	var lastErr error
	var lastStatusCode int

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
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

		slog.Debug("Making API request", "method", method, "url", url, "attempt", attempt+1)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			lastStatusCode = 0

			if attempt < c.retryConfig.MaxRetries && isRetryableError(err, 0) {
				backoff := c.calculateBackoff(attempt)
				slog.Info("Request failed, retrying", "error", err, "attempt", attempt+1, "backoff", backoff)
				time.Sleep(backoff)
				continue
			}

			return nil, fmt.Errorf("failed to make request after %d attempts: %w", attempt+1, err)
		}

		// Check for HTTP errors
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastStatusCode = resp.StatusCode
			lastErr = fmt.Errorf("API request failed with status: %d", resp.StatusCode)
			resp.Body.Close()

			if attempt < c.retryConfig.MaxRetries && isRetryableError(nil, resp.StatusCode) {
				backoff := c.calculateBackoff(attempt)
				slog.Info("Request failed with retryable status, retrying", "status", resp.StatusCode, "attempt", attempt+1, "backoff", backoff)
				time.Sleep(backoff)
				continue
			}

			return nil, fmt.Errorf("API request failed with status %d after %d attempts", resp.StatusCode, attempt+1)
		}

		// Success - return the response
		slog.Debug("API request successful", "method", method, "url", url, "attempt", attempt+1)
		return resp, nil
	}

	// This should not be reached, but included for completeness
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryConfig.MaxRetries+1, lastErr)
	}
	return nil, fmt.Errorf("request failed after %d attempts with status: %d", c.retryConfig.MaxRetries+1, lastStatusCode)
}

// SetRetryConfig allows customization of retry behavior
//
// Parameters:
//   - config: the new retry policy, used by every later request.
//
// It is used by tests to shorten waits; the CLI keeps the default.
func (c *Client) SetRetryConfig(config RetryConfig) {
	c.retryConfig = config
}

// GetRetryConfig returns the current retry configuration
//
// Returns a copy of the client's retry policy.
func (c *Client) GetRetryConfig() RetryConfig {
	return c.retryConfig
}

// GetRoutes fetches all routes for the specified organization and network
//
// Parameters:
//   - organizationID: the organization to search.
//   - networkIdentifier: a network ID or name (see [Client.ResolveNetworkID]).
//     Empty means every network in the organization, concatenated.
//
// Returns the routes (never nil), or an error if the network cannot be
// resolved or the organization's network list cannot be read. When every
// network is read, a network whose routes fail is logged and skipped.
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
//
// Its JSON shape is that of GET /organizations/{organizationId}/networks.
type Network struct {
	// ID is the Meraki network ID, for example L_123456789012345678.
	ID string `json:"id"`
	// Name is the network's dashboard name; not guaranteed unique.
	Name string `json:"name"`
	// ProductTypes lists the product families in the network.
	ProductTypes []string `json:"productTypes,omitempty"`
	// TimeZone is the network's configured time zone.
	TimeZone string `json:"timeZone,omitempty"`
	// Tags are the network's dashboard tags.
	Tags []string `json:"tags,omitempty"`
}

// Organization represents a Meraki organization
//
// Its JSON shape is that of GET /organizations. The nested fields are only
// used by the access command's report.
type Organization struct {
	// ID is the Meraki organization ID.
	ID string `json:"id"`
	// Name is the organization's name; not guaranteed unique.
	Name string `json:"name"`
	// URL is the organization's dashboard URL.
	URL string `json:"url,omitempty"`
	// API reports whether Dashboard API access is enabled.
	API struct {
		// Enabled is true when the organization allows API access.
		Enabled bool `json:"enabled"`
	} `json:"api"`
	// Licensing describes the organization's licensing.
	Licensing struct {
		// Model is the licensing model, for example co-term or per-device.
		Model string `json:"model"`
	} `json:"licensing"`
	// Cloud describes where the organization is hosted.
	Cloud struct {
		// Region is the hosting region.
		Region struct {
			// Name is the region name.
			Name string `json:"name"`
			// Host is the region's host, as the API reports it.
			Host struct {
				// Name is the host name.
				Name string `json:"name"`
			} `json:"host"`
		} `json:"region"`
	} `json:"cloud"`
}

// getOrganizationNetworks fetches all networks in an organization
//
// Parameters:
//   - organizationID: the organization whose networks to list.
//
// Returns the networks from GET /organizations/{id}/networks (first page
// only), or the request or decode error.
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
//
// It concatenates, in order: appliance static routes, site-to-site VPN
// subnets, appliance VLAN subnets, switch L3 interfaces and static routes,
// and switch-stack routes. A source that fails is logged and skipped,
// because a network rarely has all five.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the combined routes (nil if none). The error result is always
// nil today; it is kept so callers handle a future hard failure.
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
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the routes from GET /networks/{id}/appliance/staticRoutes, with
// unnamed routes labeled "Static Route N", or the request or decode error.
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
//
// Each local subnet that site-to-site VPN advertises (useVpn true) becomes
// a synthesized enabled route with ID "vpn-<index>" and no gateway.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the synthesized routes. Request and decode failures are treated
// as "no VPN configured" and return an empty slice with a nil error.
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
//
// Each appliance VLAN with a subnet becomes a route with ID "vlan-<id>"
// whose gateway is the appliance's address on that VLAN.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the synthesized routes. Request and decode failures are treated
// as "no VLANs" and return an empty slice with a nil error.
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
//
// It combines [Client.getNetworkSwitchInterfaces] and
// [Client.getNetworkSwitchStaticRoutes] for a network's L3 switches.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the combined routes; the error result is always nil because both
// helpers swallow their own failures.
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
//
// Each interface with a subnet becomes a route with ID
// "switch-iface-<interfaceId>" whose gateway is the interface address.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the synthesized routes, or an empty slice and nil error on any
// request or decode failure.
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
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the routes from GET /networks/{id}/switch/routing/staticRoutes,
// unnamed ones labeled "Switch Static Route N", or an empty slice and nil
// error on any request or decode failure.
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
//
// For every stack in the network it collects routing interfaces, static
// routes and DHCP relay servers.
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the combined routes; failures for a stack or source are logged
// at debug level and skipped, so the error result is always nil.
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
//
// Its JSON shape is that of GET /networks/{networkId}/switch/stacks.
type SwitchStack struct {
	// ID is the switch stack ID.
	ID string `json:"id"`
	// Name is the switch stack name.
	Name string `json:"name"`
}

// getNetworkSwitchStacks gets all switch stacks in a network
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the stacks, or an empty slice and nil error on any request or
// decode failure (most networks have no stacks).
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
//
// Each interface with a subnet becomes a route with ID
// "stack-<stackId>-iface-<interfaceId>".
//
// Parameters:
//   - networkID: the network that owns the stack.
//   - stackID: the switch stack to read.
//
// Returns the synthesized routes, or an empty slice and nil error on any
// request or decode failure.
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
//
// The stack endpoint has no "enabled" field, so Enabled is taken from
// preferOverOspfRoutesEnabled as a stand-in (see the comment in the body).
//
// Parameters:
//   - networkID: the network that owns the stack.
//   - stackID: the switch stack to read.
//
// Returns the routes, or an empty slice and nil error on any request or
// decode failure.
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
//
// Only the DHCP relay server addresses are turned into output: each one
// becomes a synthesized 0.0.0.0/0 "route" whose gateway is the relay
// server. DHCP options and reserved ranges are decoded but not used.
//
// Parameters:
//   - networkID: the network that owns the stack.
//   - stackID: the switch stack to read.
//
// Returns the synthesized routes, or an empty slice and nil error on any
// request or decode failure.
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
//
// Returns the organizations from GET /organizations, or the request or
// decode error.
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
//
// Parameters:
//   - organizationID: the organization whose networks to list.
//
// Returns the networks, or the request or decode error.
func (c *Client) GetOrganizationNetworks(organizationID string) ([]Network, error) {
	return c.getOrganizationNetworks(organizationID)
}

// getOrganizationDeviceStatuses fetches device statuses for all devices in an organization
//
// Parameters:
//   - organizationID: the organization to read.
//
// Returns the statuses from GET /organizations/{id}/devices/statuses
// (first page only), or the request or decode error.
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
//
// Parameters:
//   - organizationID: the organization to read.
//
// Returns one [NetworkRoutes] per network, in API order; a network whose
// routes fail is still included with an empty route list. Returns an error
// only if the network list cannot be read.
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
//
// An exact ID match wins; otherwise the name is matched case-insensitively
// and must be unique, so users can type the name they see in the dashboard.
//
// Parameters:
//   - organizationID: the organization to search.
//   - networkIdentifier: a network ID or name. Empty returns "" and nil.
//
// Returns the network ID, or an error that lists the available networks
// (no match) or the candidate IDs (several networks share the name).
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
//
// An exact ID match wins; otherwise the name is matched case-insensitively
// and must be unique.
//
// Parameters:
//   - organizationIdentifier: an organization ID or name. Empty returns ""
//     and nil.
//
// Returns the organization ID, or an error that lists the available
// organizations (no match) or the duplicates (several share the name).
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
//
// Parameters:
//   - organizationID: the organization to read.
//
// Returns the licenses from GET /organizations/{id}/licenses, or the
// request or decode error.
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
//
// Nothing in the program calls it today.
//
// Parameters:
//   - organizationID: the organization to read.
//
// Returns one [NetworkLicenses] per network: the licenses assigned to that
// network plus every organization-wide license, so org-wide licenses are
// repeated under every network. If the license list fails it is logged and
// treated as empty. Returns an error only if the network list fails.
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
//
// Parameters:
//   - organizationID: the organization to search.
//   - networkIdentifier: a network ID or name; empty means every network in
//     the organization, concatenated.
//
// Returns the devices (never nil), or an error if the network cannot be
// resolved or listed. When every network is read, a network whose devices
// fail is logged and skipped.
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
//
// It filters [Client.GetDevices] with [isDeviceDown], using the Status
// field of the network devices response as-is.
//
// Parameters:
//   - organizationID: the organization to search.
//   - networkIdentifier: a network ID or name; empty means every network.
//
// Returns the down devices (never nil), or GetDevices' error.
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
//
// Device status is taken from the organization device-statuses endpoint and
// joined on serial, because the network devices response does not reliably
// carry it. If that endpoint fails, it falls back to whatever Status the
// devices response had.
//
// Parameters:
//   - organizationID: the organization to search.
//   - networkIdentifier: a network ID or name; empty means every network.
//
// Returns the alerting devices (never nil), or GetDevices' error.
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
//
// Parameters:
//   - networkID: the network to read.
//
// Returns the devices from GET /networks/{id}/devices, or the request or
// decode error.
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
//
// Nothing in the program calls it today.
//
// Parameters:
//   - organizationID: the organization to read.
//
// Returns one [NetworkDevices] per network; a network whose devices fail
// is included with an empty list. Returns an error only if the network
// list fails.
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
//
// Parameters:
//   - status: a Meraki device status, compared case-insensitively.
//
// Returns true for offline, alerting, dormant, down, unreachable and
// disconnected. Note that "alerting" counts as down, so the down command's
// output includes every alerting device.
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
//
// Parameters:
//   - status: a Meraki device status, compared case-insensitively.
//
// Returns true only for "alerting".
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
