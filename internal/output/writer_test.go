package output

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"os"
	"strings"
	"testing"

	"meraki-info/internal/meraki"
)

// TestNewWriter checks that each format name maps to the right Writer type.
func TestNewWriter(t *testing.T) {
	tests := []struct {
		outputType   string
		expectedType interface{}
	}{
		{"json", &JSONWriter{}},
		{"JSON", &JSONWriter{}},
		{"xml", &XMLWriter{}},
		{"XML", &XMLWriter{}},
		{"csv", &CSVWriter{}},
		{"CSV", &CSVWriter{}},
		{"text", &TextWriter{}},
		{"unknown", &TextWriter{}},
		{"", &TextWriter{}},
	}

	for _, tt := range tests {
		t.Run(tt.outputType, func(t *testing.T) {
			writer := NewWriter(tt.outputType)

			switch tt.expectedType.(type) {
			case *JSONWriter:
				if _, ok := writer.(*JSONWriter); !ok {
					t.Errorf("Expected JSONWriter, got %T", writer)
				}
			case *XMLWriter:
				if _, ok := writer.(*XMLWriter); !ok {
					t.Errorf("Expected XMLWriter, got %T", writer)
				}
			case *CSVWriter:
				if _, ok := writer.(*CSVWriter); !ok {
					t.Errorf("Expected CSVWriter, got %T", writer)
				}
			case *TextWriter:
				if _, ok := writer.(*TextWriter); !ok {
					t.Errorf("Expected TextWriter, got %T", writer)
				}
			}
		})
	}
}

// TestTextWriter_WriteToFile checks that routes written as text to a file
// contain the expected fields.
func TestTextWriter_WriteToFile(t *testing.T) {
	routes := []meraki.Route{
		{
			ID:          "route1",
			Name:        "Test Route 1",
			Subnet:      "192.168.1.0/24",
			GatewayIP:   "192.168.1.1",
			GatewayVlan: 100,
			Enabled:     true,
			FixedIP:     false,
		},
		{
			ID:          "route2",
			Name:        "Test Route 2",
			Subnet:      "10.0.0.0/8",
			GatewayIP:   "10.0.0.1",
			GatewayVlan: 200,
			Enabled:     false,
			FixedIP:     true,
		},
	}

	writer := &TextWriter{}
	filename := "test_routes.txt"

	// Clean up
	defer os.Remove(filename)

	err := writer.WriteToFile(routes, filename)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read file and verify content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Meraki Route Tables") {
		t.Error("Expected header not found")
	}
	if !strings.Contains(contentStr, "Total Routes: 2") {
		t.Error("Expected route count not found")
	}
	if !strings.Contains(contentStr, "Test Route 1") {
		t.Error("Expected route name not found")
	}
}

// TestJSONWriter_WriteToFile checks that routes written as JSON to a file
// decode back into a route list.
func TestJSONWriter_WriteToFile(t *testing.T) {
	routes := []meraki.Route{
		{
			ID:          "route1",
			Name:        "Test Route 1",
			Subnet:      "192.168.1.0/24",
			GatewayIP:   "192.168.1.1",
			GatewayVlan: 100,
			Enabled:     true,
			FixedIP:     false,
		},
	}

	writer := &JSONWriter{}
	filename := "test_routes.json"

	// Clean up
	defer os.Remove(filename)

	err := writer.WriteToFile(routes, filename)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read file and verify JSON content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsedRoutes []meraki.Route
	err = json.Unmarshal(content, &parsedRoutes)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if len(parsedRoutes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(parsedRoutes))
	}

	if parsedRoutes[0].Name != "Test Route 1" {
		t.Errorf("Expected route name 'Test Route 1', got '%s'", parsedRoutes[0].Name)
	}
}

// TestXMLWriter_WriteToFile checks that routes written as XML to a file
// contain the expected elements.
func TestXMLWriter_WriteToFile(t *testing.T) {
	routes := []meraki.Route{
		{
			ID:          "route1",
			Name:        "Test Route 1",
			Subnet:      "192.168.1.0/24",
			GatewayIP:   "192.168.1.1",
			GatewayVlan: 100,
			Enabled:     true,
			FixedIP:     false,
		},
	}

	writer := &XMLWriter{}
	filename := "test_routes.xml"

	// Clean up
	defer os.Remove(filename)

	err := writer.WriteToFile(routes, filename)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read file and verify XML content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var parsedRoutes RoutesXML
	err = xml.Unmarshal(content, &parsedRoutes)
	if err != nil {
		t.Fatalf("Failed to parse XML: %v", err)
	}

	if len(parsedRoutes.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(parsedRoutes.Routes))
	}

	if parsedRoutes.Routes[0].Name != "Test Route 1" {
		t.Errorf("Expected route name 'Test Route 1', got '%s'", parsedRoutes.Routes[0].Name)
	}
}

// TestCSVWriter_WriteToFile checks that routes written as CSV to a file have
// the expected header row followed by data rows.
func TestCSVWriter_WriteToFile(t *testing.T) {
	routes := []meraki.Route{
		{
			ID:          "route1",
			Name:        "Test Route 1",
			Subnet:      "192.168.1.0/24",
			GatewayIP:   "192.168.1.1",
			GatewayVlan: 100,
			Enabled:     true,
			FixedIP:     false,
		},
	}

	writer := &CSVWriter{}
	filename := "test_routes.csv"

	// Clean up
	defer os.Remove(filename)

	err := writer.WriteToFile(routes, filename)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}

	// Read file and verify CSV content
	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		t.Error("Expected at least 2 lines (header + data)")
	}

	// Check header
	if !strings.Contains(lines[0], "ID,Name,Subnet") {
		t.Error("Expected CSV header not found")
	}

	// Check data
	if !strings.Contains(lines[1], "route1,Test Route 1") {
		t.Error("Expected CSV data not found")
	}
}

// failingWriter is an io.Writer whose every Write fails, standing in for a
// full disk or a closed pipe.
type failingWriter struct{}

// Write always fails.
//
// Returns 0 and an error.
func (failingWriter) Write(p []byte) (int, error) {
	return 0, errors.New("disk full")
}

// TestCSVWriter_WriteTo_ReportsWriteError checks that a failed write reaches
// the caller. csv.Writer buffers, so a small table only reaches the
// destination at the final flush, and that flush's error used to be dropped:
// a CSV file could be left empty or truncated while the command reported
// success.
func TestCSVWriter_WriteTo_ReportsWriteError(t *testing.T) {
	inputs := map[string]interface{}{
		"routes":              []meraki.Route{{ID: "r1", Subnet: "10.0.0.0/8"}},
		"routes with network": []meraki.RouteWithNetwork{{Route: meraki.Route{ID: "r1"}, NetworkID: "N_1"}},
		"licenses":            []meraki.License{{ID: "l1"}},
		"devices":             []meraki.Device{{Serial: "Q2XX"}},
	}
	for name, data := range inputs {
		t.Run(name, func(t *testing.T) {
			err := (&CSVWriter{}).WriteTo(data, failingWriter{})
			if err == nil {
				t.Fatal("expected the write error to be returned, got nil")
			}
		})
	}
}
