package meraki

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		shouldErr bool
	}{
		{
			name:      "valid API key",
			apiKey:    "test-api-key",
			shouldErr: false,
		},
		{
			name:      "empty API key",
			apiKey:    "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey)

			if tt.shouldErr && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.shouldErr && client == nil {
				t.Error("Expected client but got nil")
			}
		})
	}
}

func TestClient_makeRequest(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check headers
		if r.Header.Get("X-Cisco-Meraki-API-Key") != "test-api-key" {
			t.Error("Expected API key header not found")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Expected Content-Type header not found")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"test": "response"}`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	resp, err := client.makeRequest("GET", "/test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_makeRequest_Error(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	_, err := client.makeRequest("GET", "/test")
	if err == nil {
		t.Error("Expected error but got none")
	}
}

func TestClient_getOrganizationNetworks(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations/org123/networks" {
			t.Errorf("Expected path /organizations/org123/networks, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[{"id": "net1", "name": "Network 1"}, {"id": "net2", "name": "Network 2"}]`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	networks, err := client.getOrganizationNetworks("org123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(networks) != 2 {
		t.Errorf("Expected 2 networks, got %d", len(networks))
	}

	if networks[0].ID != "net1" {
		t.Errorf("Expected network ID 'net1', got '%s'", networks[0].ID)
	}
}

func TestClient_getNetworkRoutes(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/networks/net123/appliance/staticRoutes":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": "route1",
					"name": "Test Route",
					"subnet": "192.168.1.0/24",
					"gatewayIp": "192.168.1.1",
					"gatewayVlanId": 100,
					"enabled": true
				}
			]`))
		case "/networks/net123/appliance/vpn/siteToSiteVpn":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"subnets": [
					{
						"localSubnet": "10.0.0.0/8",
						"useVpn": true
					}
				]
			}`))
		case "/networks/net123/appliance/vlans":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 1,
					"name": "Default",
					"applianceIp": "172.16.1.1",
					"subnet": "172.16.1.0/24"
				}
			]`))
		case "/networks/net123/switch/routing/interfaces":
			// Return empty array for switch routing interfaces (no switch Layer 3 interfaces)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net123/switch/routing/staticRoutes":
			// Return empty array for switch static routes
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net123/switch/stacks":
			// Return empty array for switch stacks
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	routes, err := client.getNetworkRoutes("net123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(routes))
	}

	// Check static route
	if routes[0].ID != "route1" {
		t.Errorf("Expected route ID 'route1', got '%s'", routes[0].ID)
	}

	if routes[0].Subnet != "192.168.1.0/24" {
		t.Errorf("Expected subnet '192.168.1.0/24', got '%s'", routes[0].Subnet)
	}

	// Check VPN route
	if routes[1].Subnet != "10.0.0.0/8" {
		t.Errorf("Expected VPN subnet '10.0.0.0/8', got '%s'", routes[1].Subnet)
	}

	// Check VLAN route
	if routes[2].Subnet != "172.16.1.0/24" {
		t.Errorf("Expected VLAN subnet '172.16.1.0/24', got '%s'", routes[2].Subnet)
	}
}

func TestClient_GetOrganizations(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations" {
			t.Errorf("Expected path /organizations, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "123456",
				"name": "Test Organization",
				"url": "https://dashboard.meraki.com/test",
				"api": {"enabled": true},
				"licensing": {"model": "co-term"},
				"cloud": {
					"region": {
						"name": "North America",
						"host": {"name": "United States"}
					}
				}
			}
		]`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	orgs, err := client.GetOrganizations()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(orgs) != 1 {
		t.Errorf("Expected 1 organization, got %d", len(orgs))
	}

	if orgs[0].ID != "123456" {
		t.Errorf("Expected organization ID '123456', got '%s'", orgs[0].ID)
	}

	if orgs[0].Name != "Test Organization" {
		t.Errorf("Expected organization name 'Test Organization', got '%s'", orgs[0].Name)
	}

	if !orgs[0].API.Enabled {
		t.Error("Expected API to be enabled")
	}
}

func TestClient_ResolveNetworkID(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/organizations/org123/networks" {
			t.Errorf("Expected path /organizations/org123/networks, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[
			{
				"id": "net1",
				"name": "Main Office",
				"productTypes": ["appliance", "wireless"]
			},
			{
				"id": "net2", 
				"name": "Branch Office",
				"productTypes": ["wireless"]
			},
			{
				"id": "net3",
				"name": "Main Office",
				"productTypes": ["camera"]
			}
		]`))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	tests := []struct {
		name              string
		networkIdentifier string
		expectedID        string
		shouldError       bool
		errorContains     string
	}{
		{
			name:              "resolve by existing ID",
			networkIdentifier: "net1",
			expectedID:        "net1",
			shouldError:       false,
		},
		{
			name:              "resolve by unique name",
			networkIdentifier: "Branch Office",
			expectedID:        "net2",
			shouldError:       false,
		},
		{
			name:              "resolve by unique name case insensitive",
			networkIdentifier: "BRANCH OFFICE",
			expectedID:        "net2",
			shouldError:       false,
		},
		{
			name:              "resolve by duplicate name should error",
			networkIdentifier: "Main Office",
			shouldError:       true,
			errorContains:     "multiple networks found",
		},
		{
			name:              "resolve by duplicate name case insensitive should error",
			networkIdentifier: "main office",
			shouldError:       true,
			errorContains:     "multiple networks found",
		},
		{
			name:              "resolve non-existent network should error",
			networkIdentifier: "Non-existent Network",
			shouldError:       true,
			errorContains:     "not found",
		},
		{
			name:              "empty identifier returns empty",
			networkIdentifier: "",
			expectedID:        "",
			shouldError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolvedID, err := client.ResolveNetworkID("org123", tt.networkIdentifier)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if resolvedID != tt.expectedID {
					t.Errorf("Expected resolved ID '%s', got '%s'", tt.expectedID, resolvedID)
				}
			}
		})
	}
}

func TestClient_GetAllNetworkRoutes(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/organizations/org123/networks":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": "net1",
					"name": "Network 1",
					"productTypes": ["appliance"]
				},
				{
					"id": "net2", 
					"name": "Network 2",
					"productTypes": ["appliance"]
				}
			]`))
		case "/networks/net1/appliance/staticRoutes":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": "route1",
					"name": "Test Route 1",
					"subnet": "192.168.1.0/24",
					"gatewayIp": "192.168.1.1",
					"enabled": true
				}
			]`))
		case "/networks/net1/appliance/vpn/siteToSiteVpn":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"subnets": [
					{
						"localSubnet": "10.1.0.0/16",
						"useVpn": true
					}
				]
			}`))
		case "/networks/net2/appliance/staticRoutes":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": "route2",
					"name": "Test Route 2",
					"subnet": "192.168.2.0/24",
					"gatewayIp": "192.168.2.1",
					"enabled": true
				}
			]`))
		case "/networks/net2/appliance/vpn/siteToSiteVpn":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"subnets": [
					{
						"localSubnet": "10.2.0.0/16",
						"useVpn": true
					}
				]
			}`))
		case "/networks/net1/appliance/vlans":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 1,
					"name": "Default-1",
					"applianceIp": "172.16.1.1",
					"subnet": "172.16.1.0/24"
				}
			]`))
		case "/networks/net2/appliance/vlans":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"id": 1,
					"name": "Default-2",
					"applianceIp": "172.16.2.1",
					"subnet": "172.16.2.0/24"
				}
			]`))
		// Switch routing endpoints for net1
		case "/networks/net1/switch/routing/interfaces":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net1/switch/routing/staticRoutes":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net1/switch/stacks":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		// Switch routing endpoints for net2
		case "/networks/net2/switch/routing/interfaces":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net2/switch/routing/staticRoutes":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		case "/networks/net2/switch/stacks":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-api-key",
	}

	allNetworkRoutes, err := client.GetAllNetworkRoutes("org123")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(allNetworkRoutes) != 2 {
		t.Errorf("Expected 2 network route sets, got %d", len(allNetworkRoutes))
	}

	// Check first network
	if allNetworkRoutes[0].Network.ID != "net1" {
		t.Errorf("Expected network ID 'net1', got '%s'", allNetworkRoutes[0].Network.ID)
	}
	if len(allNetworkRoutes[0].Routes) != 3 {
		t.Errorf("Expected 3 routes for net1, got %d", len(allNetworkRoutes[0].Routes))
	}
	if allNetworkRoutes[0].Routes[0].ID != "route1" {
		t.Errorf("Expected route ID 'route1', got '%s'", allNetworkRoutes[0].Routes[0].ID)
	}

	// Check second network
	if allNetworkRoutes[1].Network.ID != "net2" {
		t.Errorf("Expected network ID 'net2', got '%s'", allNetworkRoutes[1].Network.ID)
	}
	if len(allNetworkRoutes[1].Routes) != 3 {
		t.Errorf("Expected 3 routes for net2, got %d", len(allNetworkRoutes[1].Routes))
	}
	if allNetworkRoutes[1].Routes[0].ID != "route2" {
		t.Errorf("Expected route ID 'route2', got '%s'", allNetworkRoutes[1].Routes[0].ID)
	}
}

func TestResolveOrganizationID(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/organizations" {
			// Mock organizations response
			response := `[
				{"id": "org1", "name": "Test Org 1"},
				{"id": "org2", "name": "Test Org 2"},
				{"id": "org3", "name": "Duplicate Name"},
				{"id": "org4", "name": "Duplicate Name"}
			]`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := &Client{
		httpClient: &http.Client{},
		baseURL:    server.URL,
		apiKey:     "test-key",
	}

	tests := []struct {
		name          string
		orgIdentifier string
		expectedID    string
		shouldErr     bool
		errorContains string
	}{
		{
			name:          "resolve by ID",
			orgIdentifier: "org1",
			expectedID:    "org1",
			shouldErr:     false,
		},
		{
			name:          "resolve by name",
			orgIdentifier: "Test Org 2",
			expectedID:    "org2",
			shouldErr:     false,
		},
		{
			name:          "resolve by name case insensitive",
			orgIdentifier: "TEST ORG 2",
			expectedID:    "org2",
			shouldErr:     false,
		},
		{
			name:          "empty identifier",
			orgIdentifier: "",
			expectedID:    "",
			shouldErr:     false,
		},
		{
			name:          "non-existent organization",
			orgIdentifier: "Non-existent Org",
			shouldErr:     true,
			errorContains: "not found",
		},
		{
			name:          "duplicate organization names",
			orgIdentifier: "Duplicate Name",
			shouldErr:     true,
			errorContains: "multiple organizations found",
		},
		{
			name:          "duplicate organization names case insensitive",
			orgIdentifier: "duplicate name",
			shouldErr:     true,
			errorContains: "multiple organizations found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgID, err := client.ResolveOrganizationID(tt.orgIdentifier)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %s", err.Error())
				}
				if orgID != tt.expectedID {
					t.Errorf("Expected organization ID '%s', got '%s'", tt.expectedID, orgID)
				}
			}
		})
	}
}
