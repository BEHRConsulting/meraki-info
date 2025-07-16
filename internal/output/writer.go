// Package output handles different output formats for route data
package output

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"meraki-info/internal/meraki"
)

// Writer interface for different output formats
type Writer interface {
	WriteToFile(data interface{}, filename string) error
	WriteTo(data interface{}, writer io.Writer) error
}

// TextWriter writes routes in plain text format
type TextWriter struct{}

// JSONWriter writes routes in JSON format
type JSONWriter struct{}

// XMLWriter writes routes in XML format
type XMLWriter struct{}

// CSVWriter writes routes in CSV format
type CSVWriter struct{}

// RoutesXML represents routes in XML format
type RoutesXML struct {
	XMLName xml.Name   `xml:"routes"`
	Routes  []RouteXML `xml:"route"`
}

// RouteXML represents a single route in XML format
type RouteXML struct {
	ID          string `xml:"id,omitempty"`
	Name        string `xml:"name,omitempty"`
	Subnet      string `xml:"subnet"`
	GatewayIP   string `xml:"gatewayIp"`
	GatewayVlan int    `xml:"gatewayVlanId,omitempty"`
	Enabled     bool   `xml:"enabled"`
	FixedIP     string `xml:"fixedIpAssignments,omitempty"`
}

// LicensesXML represents licenses in XML format
type LicensesXML struct {
	XMLName  xml.Name     `xml:"licenses"`
	Licenses []LicenseXML `xml:"license"`
}

// LicenseXML represents a single license in XML format
type LicenseXML struct {
	ID                string `xml:"id,omitempty"`
	OrganizationID    string `xml:"organizationId,omitempty"`
	DeviceSerial      string `xml:"deviceSerial,omitempty"`
	NetworkID         string `xml:"networkId,omitempty"`
	State             string `xml:"state,omitempty"`
	Edition           string `xml:"edition,omitempty"`
	Mode              string `xml:"mode,omitempty"`
	ExpirationDate    string `xml:"expirationDate,omitempty"`
	LicenseType       string `xml:"licenseType,omitempty"`
	LicenseKey        string `xml:"licenseKey,omitempty"`
	OrderNumber       string `xml:"orderNumber,omitempty"`
	PermanentlyQueued bool   `xml:"permanentlyQueued,omitempty"`
	DurationInDays    int    `xml:"durationInDays,omitempty"`
}

// NewWriter creates a new writer based on the output type
func NewWriter(outputType string) Writer {
	switch strings.ToLower(outputType) {
	case "json":
		return &JSONWriter{}
	case "xml":
		return &XMLWriter{}
	case "csv":
		return &CSVWriter{}
	default:
		return &TextWriter{}
	}
}

// WriteToFile writes data to a file in text format
func (w *TextWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(data, file)
}

// WriteTo writes data to an io.Writer in text format
func (w *TextWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutes(v, writer)
	case []meraki.License:
		return w.writeLicenses(v, writer)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutes writes routes to an io.Writer in text format
func (w *TextWriter) writeRoutes(routes []meraki.Route, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki Route Tables\n")
	fmt.Fprintf(writer, "===================\n\n")
	fmt.Fprintf(writer, "Total Routes: %d\n\n", len(routes))

	// Write routes
	for i, route := range routes {
		fmt.Fprintf(writer, "Route %d:\n", i+1)
		fmt.Fprintf(writer, "  ID: %s\n", route.ID)
		fmt.Fprintf(writer, "  Name: %s\n", route.Name)
		fmt.Fprintf(writer, "  Subnet: %s\n", route.Subnet)
		fmt.Fprintf(writer, "  Gateway IP: %s\n", route.GatewayIP)
		fmt.Fprintf(writer, "  Gateway VLAN: %d\n", route.GatewayVlan)
		fmt.Fprintf(writer, "  Enabled: %t\n", route.Enabled)
		fmt.Fprintf(writer, "  Fixed IP: %v\n", route.FixedIP)
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// writeLicenses writes licenses to an io.Writer in text format
func (w *TextWriter) writeLicenses(licenses []meraki.License, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki License Information\n")
	fmt.Fprintf(writer, "==========================\n\n")
	fmt.Fprintf(writer, "Total Licenses: %d\n\n", len(licenses))

	// Write licenses
	for i, license := range licenses {
		fmt.Fprintf(writer, "License %d:\n", i+1)
		fmt.Fprintf(writer, "  ID: %s\n", license.ID)
		fmt.Fprintf(writer, "  Organization ID: %s\n", license.OrganizationID)
		fmt.Fprintf(writer, "  Device Serial: %s\n", license.DeviceSerial)
		fmt.Fprintf(writer, "  Network ID: %s\n", license.NetworkID)
		fmt.Fprintf(writer, "  State: %s\n", license.State)
		fmt.Fprintf(writer, "  Edition: %s\n", license.Edition)
		fmt.Fprintf(writer, "  Mode: %s\n", license.Mode)
		fmt.Fprintf(writer, "  License Type: %s\n", license.LicenseType)
		fmt.Fprintf(writer, "  License Key: %s\n", license.LicenseKey)
		fmt.Fprintf(writer, "  Order Number: %s\n", license.OrderNumber)
		fmt.Fprintf(writer, "  Duration (Days): %d\n", license.DurationInDays)
		fmt.Fprintf(writer, "  Expiration Date: %s\n", license.ExpirationDate)
		fmt.Fprintf(writer, "  Permanently Queued: %t\n", license.PermanentlyQueued)
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// WriteToFile writes data to a file in JSON format
func (w *JSONWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(data, file)
}

// WriteTo writes data to an io.Writer in JSON format
func (w *JSONWriter) WriteTo(data interface{}, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// WriteToFile writes data to a file in XML format
func (w *XMLWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(data, file)
}

// WriteTo writes data to an io.Writer in XML format
func (w *XMLWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutesXML(v, writer)
	case []meraki.License:
		return w.writeLicensesXML(v, writer)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutesXML writes routes to an io.Writer in XML format
func (w *XMLWriter) writeRoutesXML(routes []meraki.Route, writer io.Writer) error {
	// Convert routes to XML-compatible format
	xmlRoutes := make([]RouteXML, len(routes))
	for i, route := range routes {
		fixedIPStr := ""
		if route.FixedIP != nil {
			// Convert interface{} to string representation
			fixedIPBytes, err := json.Marshal(route.FixedIP)
			if err == nil {
				fixedIPStr = string(fixedIPBytes)
			}
		}

		xmlRoutes[i] = RouteXML{
			ID:          route.ID,
			Name:        route.Name,
			Subnet:      route.Subnet,
			GatewayIP:   route.GatewayIP,
			GatewayVlan: route.GatewayVlan,
			Enabled:     route.Enabled,
			FixedIP:     fixedIPStr,
		}
	}

	routesXML := RoutesXML{Routes: xmlRoutes}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")

	// Write XML header
	fmt.Fprint(writer, xml.Header)

	if err := encoder.Encode(routesXML); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// writeLicensesXML writes licenses to an io.Writer in XML format
func (w *XMLWriter) writeLicensesXML(licenses []meraki.License, writer io.Writer) error {
	// Convert licenses to XML-compatible format
	xmlLicenses := make([]LicenseXML, len(licenses))
	for i, license := range licenses {
		xmlLicenses[i] = LicenseXML{
			ID:                license.ID,
			OrganizationID:    license.OrganizationID,
			DeviceSerial:      license.DeviceSerial,
			NetworkID:         license.NetworkID,
			State:             license.State,
			Edition:           license.Edition,
			Mode:              license.Mode,
			ExpirationDate:    license.ExpirationDate,
			LicenseType:       license.LicenseType,
			LicenseKey:        license.LicenseKey,
			OrderNumber:       license.OrderNumber,
			PermanentlyQueued: license.PermanentlyQueued,
			DurationInDays:    license.DurationInDays,
		}
	}

	licensesXML := LicensesXML{Licenses: xmlLicenses}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")

	// Write XML header
	fmt.Fprint(writer, xml.Header)

	if err := encoder.Encode(licensesXML); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// WriteToFile writes data to a file in CSV format
func (w *CSVWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return w.WriteTo(data, file)
}

// WriteTo writes data to an io.Writer in CSV format
func (w *CSVWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutesCSV(v, writer)
	case []meraki.License:
		return w.writeLicensesCSV(v, writer)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutesCSV writes routes to an io.Writer in CSV format
func (w *CSVWriter) writeRoutesCSV(routes []meraki.Route, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Name", "Subnet", "Gateway IP", "Gateway VLAN", "Enabled", "Fixed IP"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write routes
	for _, route := range routes {
		record := []string{
			route.ID,
			route.Name,
			route.Subnet,
			route.GatewayIP,
			fmt.Sprintf("%d", route.GatewayVlan),
			fmt.Sprintf("%t", route.Enabled),
			fmt.Sprintf("%v", route.FixedIP),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// writeLicensesCSV writes licenses to an io.Writer in CSV format
func (w *CSVWriter) writeLicensesCSV(licenses []meraki.License, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write header
	header := []string{"ID", "Organization ID", "Device Serial", "Network ID", "State", "Edition", "Mode", "License Type", "License Key", "Order Number", "Duration (Days)", "Expiration Date", "Permanently Queued"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write licenses
	for _, license := range licenses {
		record := []string{
			license.ID,
			license.OrganizationID,
			license.DeviceSerial,
			license.NetworkID,
			license.State,
			license.Edition,
			license.Mode,
			license.LicenseType,
			license.LicenseKey,
			license.OrderNumber,
			fmt.Sprintf("%d", license.DurationInDays),
			license.ExpirationDate,
			fmt.Sprintf("%t", license.PermanentlyQueued),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}
