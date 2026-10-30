// Package output handles different output formats for route data
//
// Despite that first line, it formats every record type the program
// produces: routes, licenses and devices, each in a per-network form
// ([]meraki.Route, []meraki.License, []meraki.Device) and a consolidated
// form that also names the network and organization
// ([]meraki.RouteWithNetwork, []meraki.LicenseWithNetwork,
// []meraki.DeviceWithNetwork).
//
// [NewWriter] picks a [Writer] by format name. Text, XML and CSV are written
// by hand per type; JSON uses encoding/json on whatever it is given. Note
// that the XML and CSV writers fall back to indented JSON for the
// consolidated license and device types, which have no XML/CSV layout yet.
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
//
// data is one of the six slice types listed in the package comment; any
// other type is an error for every writer except [JSONWriter].
type Writer interface {
	// WriteToFile creates (or truncates) filename and writes data to it.
	WriteToFile(data interface{}, filename string) error
	// WriteTo writes data to writer.
	WriteTo(data interface{}, writer io.Writer) error
}

// TextWriter writes routes in plain text format
//
// It writes licenses and devices too, as human-readable numbered blocks
// with a title and total count. It is the default format.
type TextWriter struct{}

// JSONWriter writes routes in JSON format
//
// It encodes any value as two-space-indented JSON, so it accepts every
// record type, including ones the other writers reject.
type JSONWriter struct{}

// XMLWriter writes routes in XML format
//
// It converts routes, licenses and devices to the *XML mirror types in this
// file and encodes them with an XML header. Consolidated licenses and
// devices are written as JSON instead (see the package comment).
type XMLWriter struct{}

// CSVWriter writes routes in CSV format
//
// It writes a header row and one record per route, license or device.
// Consolidated licenses and devices are written as JSON instead (see the
// package comment).
type CSVWriter struct{}

// RoutesXML represents routes in XML format
//
// It is the <routes> document root for per-network route output.
type RoutesXML struct {
	// XMLName sets the root element name.
	XMLName xml.Name `xml:"routes"`
	// Routes holds one <route> element per route.
	Routes []RouteXML `xml:"route"`
}

// RoutesWithNetworkXML represents routes with network information in XML format
//
// It is the <routes> document root for consolidated route output.
type RoutesWithNetworkXML struct {
	// XMLName sets the root element name.
	XMLName xml.Name `xml:"routes"`
	// Routes holds one <route> element per route.
	Routes []RouteWithNetworkXML `xml:"route"`
}

// RouteXML represents a single route in XML format
//
// It mirrors [meraki.Route] with FixedIP flattened to a string, because
// encoding/xml cannot encode the arbitrary JSON value Route holds there.
type RouteXML struct {
	// ID mirrors [meraki.Route] field ID.
	ID string `xml:"id,omitempty"`
	// Name mirrors [meraki.Route] field Name.
	Name string `xml:"name,omitempty"`
	// Subnet mirrors [meraki.Route] field Subnet.
	Subnet string `xml:"subnet"`
	// GatewayIP mirrors [meraki.Route] field GatewayIP.
	GatewayIP string `xml:"gatewayIp"`
	// GatewayVlan mirrors [meraki.Route] field GatewayVlan.
	GatewayVlan int `xml:"gatewayVlanId,omitempty"`
	// Enabled mirrors [meraki.Route] field Enabled.
	Enabled bool `xml:"enabled"`
	// FixedIP mirrors [meraki.Route] field FixedIP, re-encoded as a JSON string.
	FixedIP string `xml:"fixedIpAssignments,omitempty"`
}

// RouteWithNetworkXML represents a single route with network information in XML format
//
// It mirrors [meraki.RouteWithNetwork] flattened into one element, with
// FixedIP as a string for the same reason as [RouteXML].
type RouteWithNetworkXML struct {
	// ID mirrors [meraki.RouteWithNetwork] field ID.
	ID string `xml:"id,omitempty"`
	// Name mirrors [meraki.RouteWithNetwork] field Name.
	Name string `xml:"name,omitempty"`
	// Subnet mirrors [meraki.RouteWithNetwork] field Subnet.
	Subnet string `xml:"subnet"`
	// GatewayIP mirrors [meraki.RouteWithNetwork] field GatewayIP.
	GatewayIP string `xml:"gatewayIp"`
	// GatewayVlan mirrors [meraki.RouteWithNetwork] field GatewayVlan.
	GatewayVlan int `xml:"gatewayVlanId,omitempty"`
	// Enabled mirrors [meraki.RouteWithNetwork] field Enabled.
	Enabled bool `xml:"enabled"`
	// FixedIP mirrors [meraki.RouteWithNetwork] field FixedIP, re-encoded as a JSON string.
	FixedIP string `xml:"fixedIpAssignments,omitempty"`
	// NetworkID mirrors [meraki.RouteWithNetwork] field NetworkID.
	NetworkID string `xml:"networkId"`
	// NetworkName mirrors [meraki.RouteWithNetwork] field NetworkName.
	NetworkName string `xml:"networkName"`
	// Organization mirrors [meraki.RouteWithNetwork] field Organization.
	Organization string `xml:"organization"`
}

// LicensesXML represents licenses in XML format
//
// It is the <licenses> document root for per-organization license output.
type LicensesXML struct {
	// XMLName sets the root element name.
	XMLName xml.Name `xml:"licenses"`
	// Licenses holds one <license> element per license.
	Licenses []LicenseXML `xml:"license"`
}

// LicenseXML represents a single license in XML format
//
// It mirrors [meraki.License] field for field with XML tags.
type LicenseXML struct {
	// ID mirrors [meraki.License] field ID.
	ID string `xml:"id,omitempty"`
	// OrganizationID mirrors [meraki.License] field OrganizationID.
	OrganizationID string `xml:"organizationId,omitempty"`
	// DeviceSerial mirrors [meraki.License] field DeviceSerial.
	DeviceSerial string `xml:"deviceSerial,omitempty"`
	// NetworkID mirrors [meraki.License] field NetworkID.
	NetworkID string `xml:"networkId,omitempty"`
	// State mirrors [meraki.License] field State.
	State string `xml:"state,omitempty"`
	// Edition mirrors [meraki.License] field Edition.
	Edition string `xml:"edition,omitempty"`
	// Mode mirrors [meraki.License] field Mode.
	Mode string `xml:"mode,omitempty"`
	// ExpirationDate mirrors [meraki.License] field ExpirationDate.
	ExpirationDate string `xml:"expirationDate,omitempty"`
	// LicenseType mirrors [meraki.License] field LicenseType.
	LicenseType string `xml:"licenseType,omitempty"`
	// LicenseKey mirrors [meraki.License] field LicenseKey.
	LicenseKey string `xml:"licenseKey,omitempty"`
	// OrderNumber mirrors [meraki.License] field OrderNumber.
	OrderNumber string `xml:"orderNumber,omitempty"`
	// PermanentlyQueued mirrors [meraki.License] field PermanentlyQueued.
	PermanentlyQueued bool `xml:"permanentlyQueued,omitempty"`
	// DurationInDays mirrors [meraki.License] field DurationInDays.
	DurationInDays int `xml:"durationInDays,omitempty"`
}

// DevicesXML represents devices in XML format
//
// It is the <devices> document root for per-network device output.
type DevicesXML struct {
	// XMLName sets the root element name.
	XMLName xml.Name `xml:"devices"`
	// Devices holds one <device> element per device.
	Devices []DeviceXML `xml:"device"`
}

// DeviceXML represents a single device in XML format
//
// It mirrors [meraki.Device] without BeaconIdParams.
type DeviceXML struct {
	// Serial mirrors [meraki.Device] field Serial.
	Serial string `xml:"serial"`
	// Name mirrors [meraki.Device] field Name.
	Name string `xml:"name,omitempty"`
	// Model mirrors [meraki.Device] field Model.
	Model string `xml:"model"`
	// NetworkID mirrors [meraki.Device] field NetworkID.
	NetworkID string `xml:"networkId"`
	// MAC mirrors [meraki.Device] field MAC.
	MAC string `xml:"mac,omitempty"`
	// Status mirrors [meraki.Device] field Status.
	Status string `xml:"status"`
	// LastReportedAt mirrors [meraki.Device] field LastReportedAt.
	LastReportedAt string `xml:"lastReportedAt,omitempty"`
	// ProductType mirrors [meraki.Device] field ProductType.
	ProductType string `xml:"productType,omitempty"`
	// Tags mirrors [meraki.Device] field Tags.
	Tags []string `xml:"tags,omitempty"`
	// Address mirrors [meraki.Device] field Address.
	Address string `xml:"address,omitempty"`
	// Lat mirrors [meraki.Device] field Lat.
	Lat float64 `xml:"lat,omitempty"`
	// Lng mirrors [meraki.Device] field Lng.
	Lng float64 `xml:"lng,omitempty"`
	// Notes mirrors [meraki.Device] field Notes.
	Notes string `xml:"notes,omitempty"`
}

// NewWriter creates a new writer based on the output type
//
// Parameters:
//   - outputType: "json", "xml", "csv" or "text", case-insensitive.
//
// Returns the matching Writer. Any unrecognized value, including "",
// returns a [TextWriter] rather than an error.
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
//
// Parameters:
//   - data: one of the supported record slices.
//   - filename: the file to create or truncate.
//
// Returns an error if the file cannot be created or closed, or data is an
// unsupported type. Write errors are not reported, because
// [TextWriter.WriteTo] does not check them.
func (w *TextWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if err := w.WriteTo(data, file); err != nil {
		_ = file.Close() // the write error is the one worth reporting
		return err
	}
	// A failed Close can mean the data never reached the disk (for example
	// on a network share), so it is reported rather than deferred away.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// WriteTo writes data to an io.Writer in text format
//
// Parameters:
//   - data: one of the supported record slices.
//   - writer: the destination.
//
// Returns an error only for an unsupported data type; write errors on
// writer are not checked.
func (w *TextWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutes(v, writer)
	case []meraki.RouteWithNetwork:
		return w.writeRoutesWithNetwork(v, writer)
	case []meraki.License:
		return w.writeLicenses(v, writer)
	case []meraki.LicenseWithNetwork:
		return w.writeLicensesWithNetwork(v, writer)
	case []meraki.Device:
		return w.writeDevices(v, writer)
	case []meraki.DeviceWithNetwork:
		return w.writeDevicesWithNetwork(v, writer)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutes writes routes to an io.Writer in text format
//
// Parameters:
//   - routes: the routes to print.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
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

// writeRoutesWithNetwork writes routes with network information to an io.Writer in text format
//
// Parameters:
//   - routes: the consolidated routes to print.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
func (w *TextWriter) writeRoutesWithNetwork(routes []meraki.RouteWithNetwork, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki Route Tables - Consolidated View\n")
	fmt.Fprintf(writer, "=======================================\n\n")
	fmt.Fprintf(writer, "Total Routes: %d\n\n", len(routes))

	// Write routes
	for i, route := range routes {
		fmt.Fprintf(writer, "Route %d:\n", i+1)
		fmt.Fprintf(writer, "  Organization: %s\n", route.Organization)
		fmt.Fprintf(writer, "  Network ID: %s\n", route.NetworkID)
		fmt.Fprintf(writer, "  Network Name: %s\n", route.NetworkName)
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
//
// Parameters:
//   - licenses: the licenses to print.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
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

// writeDevices writes devices to an io.Writer in text format
//
// Parameters:
//   - devices: the devices to print. The title says "Down Devices" whatever
//     command produced them, so alerting output is labeled "down" too.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
func (w *TextWriter) writeDevices(devices []meraki.Device, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki Down Devices\n")
	fmt.Fprintf(writer, "===================\n\n")
	fmt.Fprintf(writer, "Total Down Devices: %d\n\n", len(devices))

	if len(devices) == 0 {
		fmt.Fprintf(writer, "No devices are currently down.\n")
		return nil
	}

	// Write devices
	for i, device := range devices {
		fmt.Fprintf(writer, "Device %d:\n", i+1)
		fmt.Fprintf(writer, "  Serial: %s\n", device.Serial)
		fmt.Fprintf(writer, "  Name: %s\n", device.Name)
		fmt.Fprintf(writer, "  Model: %s\n", device.Model)
		fmt.Fprintf(writer, "  Network ID: %s\n", device.NetworkID)
		fmt.Fprintf(writer, "  MAC: %s\n", device.MAC)
		fmt.Fprintf(writer, "  Status: %s\n", device.Status)
		fmt.Fprintf(writer, "  Last Reported: %s\n", device.LastReportedAt)
		fmt.Fprintf(writer, "  Product Type: %s\n", device.ProductType)
		if len(device.Tags) > 0 {
			fmt.Fprintf(writer, "  Tags: %v\n", device.Tags)
		}
		if device.Address != "" {
			fmt.Fprintf(writer, "  Address: %s\n", device.Address)
		}
		if device.Lat != 0 || device.Lng != 0 {
			fmt.Fprintf(writer, "  Location: %.6f, %.6f\n", device.Lat, device.Lng)
		}
		if device.Notes != "" {
			fmt.Fprintf(writer, "  Notes: %s\n", device.Notes)
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// writeLicensesWithNetwork writes licenses with network info to an io.Writer in text format
//
// Parameters:
//   - licenses: the consolidated licenses to print.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
func (w *TextWriter) writeLicensesWithNetwork(licenses []meraki.LicenseWithNetwork, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki License Information\n")
	fmt.Fprintf(writer, "==========================\n\n")
	fmt.Fprintf(writer, "Total Licenses: %d\n\n", len(licenses))

	// Write licenses
	for i, licenseWithNetwork := range licenses {
		license := licenseWithNetwork.License
		fmt.Fprintf(writer, "License %d:\n", i+1)
		fmt.Fprintf(writer, "  ID: %s\n", license.ID)
		fmt.Fprintf(writer, "  Organization: %s\n", licenseWithNetwork.Organization)
		fmt.Fprintf(writer, "  Organization ID: %s\n", licenseWithNetwork.OrganizationID)
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

// writeDevicesWithNetwork writes devices with network info to an io.Writer in text format
//
// Parameters:
//   - devices: the consolidated devices to print.
//   - writer: the destination.
//
// Returns nil; write errors on writer are not checked.
func (w *TextWriter) writeDevicesWithNetwork(devices []meraki.DeviceWithNetwork, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Meraki Devices Information\n")
	fmt.Fprintf(writer, "==========================\n\n")
	fmt.Fprintf(writer, "Total Devices: %d\n\n", len(devices))

	if len(devices) == 0 {
		fmt.Fprintf(writer, "No devices found.\n")
		return nil
	}

	// Write devices
	for i, deviceWithNetwork := range devices {
		device := deviceWithNetwork.Device
		fmt.Fprintf(writer, "Device %d:\n", i+1)
		fmt.Fprintf(writer, "  Serial: %s\n", device.Serial)
		fmt.Fprintf(writer, "  Name: %s\n", device.Name)
		fmt.Fprintf(writer, "  Model: %s\n", device.Model)
		fmt.Fprintf(writer, "  Organization: %s\n", deviceWithNetwork.Organization)
		fmt.Fprintf(writer, "  Organization ID: %s\n", deviceWithNetwork.OrganizationID)
		fmt.Fprintf(writer, "  Network Name: %s\n", deviceWithNetwork.NetworkName)
		fmt.Fprintf(writer, "  Network ID: %s\n", deviceWithNetwork.NetworkID)
		fmt.Fprintf(writer, "  MAC: %s\n", device.MAC)
		fmt.Fprintf(writer, "  Status: %s\n", device.Status)
		fmt.Fprintf(writer, "  Last Reported: %s\n", device.LastReportedAt)
		fmt.Fprintf(writer, "  Product Type: %s\n", device.ProductType)
		if len(device.Tags) > 0 {
			fmt.Fprintf(writer, "  Tags: %v\n", device.Tags)
		}
		if device.Address != "" {
			fmt.Fprintf(writer, "  Address: %s\n", device.Address)
		}
		if device.Lat != 0 || device.Lng != 0 {
			fmt.Fprintf(writer, "  Location: %.6f, %.6f\n", device.Lat, device.Lng)
		}
		if device.Notes != "" {
			fmt.Fprintf(writer, "  Notes: %s\n", device.Notes)
		}
		fmt.Fprintf(writer, "\n")
	}

	return nil
}

// WriteToFile writes data to a file in JSON format
//
// Parameters:
//   - data: any JSON-encodable value.
//   - filename: the file to create or truncate.
//
// Returns an error if the file cannot be created or closed, or encoding
// fails.
func (w *JSONWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if err := w.WriteTo(data, file); err != nil {
		_ = file.Close() // the write error is the one worth reporting
		return err
	}
	// A failed Close can mean the data never reached the disk (for example
	// on a network share), so it is reported rather than deferred away.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// WriteTo writes data to an io.Writer in JSON format
//
// Parameters:
//   - data: any JSON-encodable value; a nil slice encodes as null.
//   - writer: the destination.
//
// Returns an error if encoding or writing fails.
func (w *JSONWriter) WriteTo(data interface{}, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// WriteToFile writes data to a file in XML format
//
// Parameters:
//   - data: one of the supported record slices.
//   - filename: the file to create or truncate.
//
// Returns an error if the file cannot be created or closed, data is
// unsupported, or encoding fails.
func (w *XMLWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if err := w.WriteTo(data, file); err != nil {
		_ = file.Close() // the write error is the one worth reporting
		return err
	}
	// A failed Close can mean the data never reached the disk (for example
	// on a network share), so it is reported rather than deferred away.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// WriteTo writes data to an io.Writer in XML format
//
// Parameters:
//   - data: one of the supported record slices. []meraki.LicenseWithNetwork
//     and []meraki.DeviceWithNetwork are written as indented JSON.
//   - writer: the destination.
//
// Returns an error for an unsupported type or a failed encode.
func (w *XMLWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutesXML(v, writer)
	case []meraki.RouteWithNetwork:
		return w.writeRoutesWithNetworkXML(v, writer)
	case []meraki.License:
		return w.writeLicensesXML(v, writer)
	case []meraki.LicenseWithNetwork:
		// Use JSON encoding for consolidated structures
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	case []meraki.Device:
		return w.writeDevicesXML(v, writer)
	case []meraki.DeviceWithNetwork:
		// Use JSON encoding for consolidated structures
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutesXML writes routes to an io.Writer in XML format
//
// Parameters:
//   - routes: the routes to encode.
//   - writer: the destination.
//
// Returns an error if XML encoding fails.
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

// writeRoutesWithNetworkXML writes routes with network information to an io.Writer in XML format
//
// Parameters:
//   - routes: the consolidated routes to encode.
//   - writer: the destination.
//
// Returns an error if XML encoding fails.
func (w *XMLWriter) writeRoutesWithNetworkXML(routes []meraki.RouteWithNetwork, writer io.Writer) error {
	// Convert routes to XML-compatible format
	xmlRoutes := make([]RouteWithNetworkXML, len(routes))
	for i, route := range routes {
		fixedIPStr := ""
		if route.FixedIP != nil {
			// Convert interface{} to string representation
			fixedIPBytes, err := json.Marshal(route.FixedIP)
			if err == nil {
				fixedIPStr = string(fixedIPBytes)
			}
		}

		xmlRoutes[i] = RouteWithNetworkXML{
			ID:           route.ID,
			Name:         route.Name,
			Subnet:       route.Subnet,
			GatewayIP:    route.GatewayIP,
			GatewayVlan:  route.GatewayVlan,
			Enabled:      route.Enabled,
			FixedIP:      fixedIPStr,
			NetworkID:    route.NetworkID,
			NetworkName:  route.NetworkName,
			Organization: route.Organization,
		}
	}

	routesXML := RoutesWithNetworkXML{Routes: xmlRoutes}

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
//
// Parameters:
//   - licenses: the licenses to encode.
//   - writer: the destination.
//
// Returns an error if XML encoding fails.
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

// writeDevicesXML writes devices to an io.Writer in XML format
//
// Parameters:
//   - devices: the devices to encode.
//   - writer: the destination.
//
// Returns an error if XML encoding fails.
func (w *XMLWriter) writeDevicesXML(devices []meraki.Device, writer io.Writer) error {
	// Convert devices to XML-compatible format
	xmlDevices := make([]DeviceXML, len(devices))
	for i, device := range devices {
		xmlDevices[i] = DeviceXML{
			Serial:         device.Serial,
			Name:           device.Name,
			Model:          device.Model,
			NetworkID:      device.NetworkID,
			MAC:            device.MAC,
			Status:         device.Status,
			LastReportedAt: device.LastReportedAt,
			ProductType:    device.ProductType,
			Tags:           device.Tags,
			Address:        device.Address,
			Lat:            device.Lat,
			Lng:            device.Lng,
			Notes:          device.Notes,
		}
	}

	devicesXML := DevicesXML{Devices: xmlDevices}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")

	// Write XML header
	fmt.Fprint(writer, xml.Header)

	if err := encoder.Encode(devicesXML); err != nil {
		return fmt.Errorf("failed to encode XML: %w", err)
	}

	return nil
}

// WriteToFile writes data to a file in CSV format
//
// Parameters:
//   - data: one of the supported record slices.
//   - filename: the file to create or truncate.
//
// Returns an error if the file cannot be created or closed, data is
// unsupported, or writing fails.
func (w *CSVWriter) WriteToFile(data interface{}, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	if err := w.WriteTo(data, file); err != nil {
		_ = file.Close() // the write error is the one worth reporting
		return err
	}
	// A failed Close can mean the data never reached the disk (for example
	// on a network share), so it is reported rather than deferred away.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// WriteTo writes data to an io.Writer in CSV format
//
// Parameters:
//   - data: one of the supported record slices. []meraki.LicenseWithNetwork
//     and []meraki.DeviceWithNetwork are written as indented JSON.
//   - writer: the destination.
//
// Returns an error for an unsupported type or a failed record write.
func (w *CSVWriter) WriteTo(data interface{}, writer io.Writer) error {
	switch v := data.(type) {
	case []meraki.Route:
		return w.writeRoutesCSV(v, writer)
	case []meraki.RouteWithNetwork:
		return w.writeRoutesWithNetworkCSV(v, writer)
	case []meraki.License:
		return w.writeLicensesCSV(v, writer)
	case []meraki.LicenseWithNetwork:
		// Use JSON encoding for consolidated structures
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	case []meraki.Device:
		return w.writeDevicesCSV(v, writer)
	case []meraki.DeviceWithNetwork:
		// Use JSON encoding for consolidated structures
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(v)
	default:
		return fmt.Errorf("unsupported data type: %T", data)
	}
}

// writeRoutesCSV writes routes to an io.Writer in CSV format
//
// Parameters:
//   - routes: the routes to write.
//   - writer: the destination.
//
// Returns an error if a header or record cannot be written, or if the
// final flush to writer fails.
func (w *CSVWriter) writeRoutesCSV(routes []meraki.Route, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)

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

	// csv.Writer buffers; the final flush is where a short table actually
	// reaches the destination, so its error must not be dropped.
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// writeRoutesWithNetworkCSV writes routes with network information to an io.Writer in CSV format
//
// Parameters:
//   - routes: the consolidated routes to write.
//   - writer: the destination.
//
// Returns an error if a header or record cannot be written, or if the
// final flush to writer fails.
func (w *CSVWriter) writeRoutesWithNetworkCSV(routes []meraki.RouteWithNetwork, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)

	// Write header
	header := []string{"Organization", "Network ID", "Network Name", "ID", "Name", "Subnet", "Gateway IP", "Gateway VLAN", "Enabled", "Fixed IP"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write routes
	for _, route := range routes {
		record := []string{
			route.Organization,
			route.NetworkID,
			route.NetworkName,
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

	// csv.Writer buffers; the final flush is where a short table actually
	// reaches the destination, so its error must not be dropped.
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// writeLicensesCSV writes licenses to an io.Writer in CSV format
//
// Parameters:
//   - licenses: the licenses to write.
//   - writer: the destination.
//
// Returns an error if a header or record cannot be written, or if the
// final flush to writer fails.
func (w *CSVWriter) writeLicensesCSV(licenses []meraki.License, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)

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

	// csv.Writer buffers; the final flush is where a short table actually
	// reaches the destination, so its error must not be dropped.
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}

// writeDevicesCSV writes devices to an io.Writer in CSV format
//
// Parameters:
//   - devices: the devices to write; Tags are joined with ";".
//   - writer: the destination.
//
// Returns an error if a header or record cannot be written, or if the
// final flush to writer fails.
func (w *CSVWriter) writeDevicesCSV(devices []meraki.Device, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)

	// Write header
	header := []string{"Serial", "Name", "Model", "Network ID", "MAC", "Status", "Last Reported At", "Product Type", "Tags", "Address", "Latitude", "Longitude", "Notes"}
	if err := csvWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write devices
	for _, device := range devices {
		// Convert tags slice to comma-separated string
		tagsStr := ""
		if len(device.Tags) > 0 {
			tagsStr = strings.Join(device.Tags, ";")
		}

		record := []string{
			device.Serial,
			device.Name,
			device.Model,
			device.NetworkID,
			device.MAC,
			device.Status,
			device.LastReportedAt,
			device.ProductType,
			tagsStr,
			device.Address,
			fmt.Sprintf("%.6f", device.Lat),
			fmt.Sprintf("%.6f", device.Lng),
			device.Notes,
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	// csv.Writer buffers; the final flush is where a short table actually
	// reaches the destination, so its error must not be dropped.
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to write CSV: %w", err)
	}

	return nil
}
