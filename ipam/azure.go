// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package ipam

import (
	"encoding/xml"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-container-networking/common"
)

const (
	// Host URL to query.
	azureQueryUrl = "http://169.254.169.254/machine/plugins?comp=nmagent&type=getinterfaceinfov1"

	// Minimum time interval between consecutive queries.
	azureQueryInterval = 10 * time.Second
)

// Microsoft Azure IPAM configuration source.
type azureSource struct {
	name          string
	sink          addressConfigSink
	queryUrl      string
	queryInterval time.Duration
	lastRefresh   time.Time
}

// Azure host agent XML document format.
type xmlDocument struct {
	XMLName   xml.Name `xml:"Interfaces"`
	Interface []struct {
		XMLName    xml.Name `xml:"Interface"`
		MacAddress string   `xml:"MacAddress,attr"`
		IsPrimary  bool     `xml:"IsPrimary,attr"`

		IPSubnet []struct {
			XMLName xml.Name `xml:"IPSubnet"`
			Prefix  string   `xml:"Prefix,attr"`

			IPAddress []struct {
				XMLName   xml.Name `xml:"IPAddress"`
				Address   string   `xml:"Address,attr"`
				IsPrimary bool     `xml:"IsPrimary,attr"`
			}
		}
	}
}

// Creates the Azure source.
func newAzureSource(options map[string]interface{}) (*azureSource, error) {
	queryUrl, _ := options[common.OptIpamQueryUrl].(string)
	if queryUrl == "" {
		queryUrl = azureQueryUrl
	}

	i, _ := options[common.OptIpamQueryInterval].(int)
	queryInterval := time.Duration(i) * time.Second
	if queryInterval == 0 {
		queryInterval = azureQueryInterval
	}

	return &azureSource{
		name:          "Azure",
		queryUrl:      queryUrl,
		queryInterval: queryInterval,
	}, nil
}

// Starts the Azure source.
func (s *azureSource) start(sink addressConfigSink) error {
	s.sink = sink
	return nil
}

// Stops the Azure source.
func (s *azureSource) stop() {
	s.sink = nil
	return
}

// Refreshes configuration.
func (s *azureSource) refresh() error {

	// Refresh only if enough time has passed since the last query.
	if time.Since(s.lastRefresh) < s.queryInterval {
		return nil
	}
	s.lastRefresh = time.Now()

	// Query the list of local interfaces.
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	// Configure the local default address space.
	local, err := s.sink.newAddressSpace(LocalDefaultAddressSpaceId, LocalScope)
	if err != nil {
		return err
	}

	// Fetch configuration.
	resp, err := http.Get(s.queryUrl)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// Decode XML document.
	var doc xmlDocument
	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(&doc)
	if err != nil {
		return err
	}

	// For each interface...
	for _, i := range doc.Interface {
		ifName := ""
		priority := 0
		i.MacAddress = strings.ToLower(i.MacAddress)

		// Find the interface with the matching MacAddress.
		for _, iface := range interfaces {
			macAddr := strings.Replace(iface.HardwareAddr.String(), ":", "", -1)
			macAddr = strings.ToLower(macAddr)
			if macAddr == i.MacAddress || i.MacAddress == "*" {
				ifName = iface.Name

				// Prioritize secondary interfaces.
				if !i.IsPrimary {
					priority = 1
				}
				break
			}
		}

		// Skip if interface is not found.
		if ifName == "" {
			continue
		}

		// For each subnet on the interface...
		for _, s := range i.IPSubnet {
			_, subnet, err := net.ParseCIDR(s.Prefix)
			if err != nil {
				return err
			}

			ap, err := local.newAddressPool(ifName, "", priority, subnet)
			if err != nil && err != errAddressExists {
				return err
			}

			// For each address in the subnet...
			for _, a := range s.IPAddress {
				// Primary addresses are reserved for the host.
				if a.IsPrimary {
					continue
				}

				address := net.ParseIP(a.Address)

				_, err = ap.newAddressRecord(&address)
				if err != nil {
					return err
				}
			}
		}
	}

	// Set the local address space as active.
	s.sink.setAddressSpace(local)

	return nil
}
