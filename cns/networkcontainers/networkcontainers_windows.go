// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package networkcontainers

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/Azure/azure-container-networking/cns"
	"github.com/Azure/azure-container-networking/log"
)

func createOrUpdateInterface(createNetworkContainerRequest cns.CreateNetworkContainerRequest) error {
	exists, _ := interfaceExists(createNetworkContainerRequest.NetworkContainerid)

	if !exists {
		return createOrUpdateWithOperation(createNetworkContainerRequest, "CREATE")
	}

	return createOrUpdateWithOperation(createNetworkContainerRequest, "UPDATE")
}

func createOrUpdateWithOperation(createNetworkContainerRequest cns.CreateNetworkContainerRequest, operation string) error {
	if _, err := os.Stat("./AzureNetworkContainer.exe"); err != nil {
		if os.IsNotExist(err) {
			return errors.New("[Azure CNS] Unable to find AzureNetworkContainer.exe. Cannot continue")
		}
	}

	if createNetworkContainerRequest.IPConfiguration.IPSubnet.IPAddress == "" {
		return errors.New("[Azure CNS] IPAddress in IPConfiguration of createNetworkContainerRequest is nil")
	}

	var dnsServers string

	for _, element := range createNetworkContainerRequest.IPConfiguration.DNSServers {
		dnsServers += element
	}

	ipv4AddrCidr := fmt.Sprintf("%v/%d", createNetworkContainerRequest.IPConfiguration.IPSubnet.IPAddress, createNetworkContainerRequest.IPConfiguration.IPSubnet.PrefixLength)
	log.Printf("[Azure CNS] Created ipv4Cidr as %v", ipv4AddrCidr)
	ipv4Addr, _, err := net.ParseCIDR(ipv4AddrCidr)
	ipv4NetInt := net.CIDRMask((int)(createNetworkContainerRequest.IPConfiguration.IPSubnet.PrefixLength), 32)
	log.Printf("[Azure CNS] Created netmask as %v", ipv4NetInt)
	ipv4NetStr := fmt.Sprintf("%d.%d.%d.%d", ipv4NetInt[0], ipv4NetInt[1], ipv4NetInt[2], ipv4NetInt[3])
	log.Printf("[Azure CNS] Created netmask in string format %v", ipv4NetStr)

	args := []string{"/C", "AzureNetworkContainer.exe", "/logpath", "./",
		"/name",
		createNetworkContainerRequest.NetworkContainerid,
		"/operation",
		operation,
		"/ip",
		ipv4Addr.String(),
		"/netmask",
		ipv4NetStr,
		"/gateway",
		createNetworkContainerRequest.IPConfiguration.GatewayIPAddress,
		"/dns",
		dnsServers,
		"/weakhostsend true /weakhostreceive true"}

	log.Printf("[Azure CNS] Going to create/update network loopback adapter: %v", args)
	c := exec.Command("cmd", args...)
	bytes, err := c.Output()

	if err == nil {
		log.Printf("[Azure CNS] Successfully created network loopback adapter %v.\n", string(bytes))
	} else {
		log.Printf("Received error while Creating a Network Container %v %v", err.Error(), string(bytes))
	}

	return err
}

func deleteInterface(networkContainerID string) error {

	if _, err := os.Stat("./AzureNetworkContainer.exe"); err != nil {
		if os.IsNotExist(err) {
			return errors.New("[Azure CNS] Unable to find AzureNetworkContainer.exe. Cannot continue")
		}
	}

	if networkContainerID == "" {
		return errors.New("[Azure CNS] networkContainerID is nil")
	}

	args := []string{"/C", "AzureNetworkContainer.exe", "/logpath", "./",
		"/name",
		networkContainerID,
		"/operation",
		"DELETE"}

	log.Printf("[Azure CNS] Going to delete network loopback adapter: %v", args)
	c := exec.Command("cmd", args...)
	bytes, err := c.Output()

	if err == nil {
		log.Printf("[Azure CNS] Successfully deleted network container %v.\n", string(bytes))
	} else {
		log.Printf("Received error while deleting a Network Container %v %v", err.Error(), string(bytes))
		return err
	}
	return nil
}
