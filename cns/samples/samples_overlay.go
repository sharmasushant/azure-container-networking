// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Azure/azure-container-networking/cns"
)

const (
	defaultCNSServerURL = "http://localhost:10090"
)

// These are example showing how to use CNS APIs by clients

func setEnvironment() error {
	envRequest := cns.SetEnvironmentRequest{Location: "Azure", NetworkType: "Overlay"}
	envRequestJSON := new(bytes.Buffer)
	json.NewEncoder(envRequestJSON).Encode(envRequest)
	res, err :=
		http.Post(
			defaultCNSServerURL+cns.SetEnvironmentPath,
			"application/json; charset=utf-8",
			envRequestJSON)
	if err != nil {
		fmt.Printf("Error received in Set Env: %v\n", err.Error())
		return err
	}
	var setEnvironmentResponse cns.Response
	err = json.NewDecoder(res.Body).Decode(&setEnvironmentResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from SetEnvironment: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for SetEnvironment: %+v\n", setEnvironmentResponse)
	return nil
}

func createNetwork() error {
	overlaySubnet := &cns.Subnet{IPAddress: "10.0.0.0", PrefixLength: 16}
	nodeConfigs := make([]cns.NodeConfiguration, 3)

	node0Subnet := &cns.Subnet{IPAddress: "10.0.0.0", PrefixLength: 24}
	config0 := &cns.NodeConfiguration{NodeID: "id0", NodeIP: "10.0.0.1", NodeSubnet: *node0Subnet}
	nodeConfigs[0] = *config0

	node1Subnet := &cns.Subnet{IPAddress: "10.0.1.0", PrefixLength: 24}
	config1 := &cns.NodeConfiguration{NodeID: "id1", NodeIP: "10.0.1.1", NodeSubnet: *node1Subnet}
	nodeConfigs[1] = *config1

	node2Subnet := &cns.Subnet{IPAddress: "10.0.2.0", PrefixLength: 24}
	config2 := &cns.NodeConfiguration{NodeID: "id2", NodeIP: "10.0.2.1", NodeSubnet: *node2Subnet}
	nodeConfigs[2] = *config2

	overlayConfig := &cns.OverlayConfiguration{
		NodeCount:     3,
		LocalNodeIP:   "10.0.0.1",
		OverlaySubnet: overlaySubnet,
		NodeConfig:    nodeconfigs,
	}

	netRequest := cns.CreateNetworkRequest{NetworkName: "azureovnet", OverlayConfiguration: *overlayConfig}
	netRequestJSON := new(bytes.Buffer)
	json.NewEncoder(netRequestJSON).Encode(netRequest)
	res, err :=
		http.Post(
			defaultCNSServerURL+cns.CreateNetworkPath,
			"application/json; charset=utf-8",
			netRequestJSON)
	if err != nil {
		fmt.Printf("Error received in CreateNetwork post: %v\n", err.Error())
		return err
	}

	var createNetworkResponse cns.Response
	err = json.NewDecoder(res.Body).Decode(&createNetworkResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from CreateNEtwork: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for CreateNetwork: %+v\n", createNetworkResponse)
	return nil
}

func deleteNetwork() error {
	netRequest := cns.DeleteNetworkRequest{NetworkName: "azurenet"}
	netRequestJSON := new(bytes.Buffer)
	json.NewEncoder(netRequestJSON).Encode(netRequest)
	res, err :=
		http.Post(
			defaultCNSServerURL+cns.DeleteNetworkPath,
			"application/json; charset=utf-8",
			netRequestJSON)
	if err != nil {
		fmt.Printf("Error received in DeleteNetwork post: %v\n", err.Error())
		return err
	}

	var deleteNetworkResponse cns.Response
	err = json.NewDecoder(res.Body).Decode(&deleteNetworkResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from DeleteNetwork: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for DeleteNetwork: %+v\n", deleteNetworkResponse)
	return nil
}

func reserveIPAddress(ipID string) error {
	reserveIPRequest := cns.ReserveIPAddressRequest{ReservationID: ipID}
	reserveIPRequestJSON := new(bytes.Buffer)
	json.NewEncoder(reserveIPRequestJSON).Encode(reserveIPRequest)
	res, err :=
		http.Post(
			defaultCNSServerURL+cns.ReserveIPAddressPath,
			"application/json; charset=utf-8",
			reserveIPRequestJSON)
	if err != nil {
		fmt.Printf("Error received in reserveIPAddress post: %v\n", err.Error())
		return err
	}
	var reserveIPAddressResponse cns.ReserveIPAddressResponse
	err = json.NewDecoder(res.Body).Decode(&reserveIPAddressResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from reserveIPAddress: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for reserveIPAddress: %+v\n", reserveIPAddressResponse)
	return nil
}

func releaseIPAddress(ipID string) error {
	releaseIPRequest := cns.ReleaseIPAddressRequest{ReservationID: ipID}
	releaseIPAddressRequestJSON := new(bytes.Buffer)
	json.NewEncoder(releaseIPAddressRequestJSON).Encode(releaseIPRequest)
	res, err :=
		http.Post(
			defaultCNSServerURL+cns.ReleaseIPAddressPath,
			"application/json; charset=utf-8",
			releaseIPAddressRequestJSON)
	if err != nil {
		fmt.Printf("Error received in releaseIPAddress post: %v\n", err.Error())
		return err
	}
	var releaseIPAddressResponse cns.Response
	err = json.NewDecoder(res.Body).Decode(&releaseIPAddressResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from releaseIPAddress: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for releaseIPAddress: %+v\n", releaseIPAddressResponse)
	return nil
}

func getIPAddressUtilization() error {
	res, err :=
		http.Get(defaultCNSServerURL + cns.GetIPAddressUtilizationPath)
	if err != nil {
		fmt.Printf("Error received in getIPAddressUtilization GET: %v\n", err.Error())
		return err
	}
	var iPAddressesUtilizationResponse cns.IPAddressesUtilizationResponse
	err = json.NewDecoder(res.Body).Decode(&iPAddressesUtilizationResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from getIPAddressUtilization: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for getIPAddressUtilization: %+v\n", iPAddressesUtilizationResponse)
	return nil
}

func getHostLocalIP() error {
	res, err :=
		http.Get(defaultCNSServerURL + cns.GetHostLocalIPPath)
	if err != nil {
		fmt.Printf("Error received in getHostLocalIP GET: %v\n", err.Error())
		return err
	}
	var hostLocalIPAddressResponse cns.HostLocalIPAddressResponse
	err = json.NewDecoder(res.Body).Decode(&hostLocalIPAddressResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from getHostLocalIP: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for getHostLocalIP: %+v\n", hostLocalIPAddressResponse)
	return nil
}

func getUnhealthyIPAddresses() error {
	res, err :=
		http.Get(defaultCNSServerURL + cns.GetUnhealthyIPAddressesPath)
	if err != nil {
		fmt.Printf("Error received in GetUnhealthyIPAddresses IP Addresses: %v\n", err.Error())
		return err
	}
	var getIPAddressesResponse cns.GetIPAddressesResponse
	err = json.NewDecoder(res.Body).Decode(&getIPAddressesResponse)
	if err != nil {
		fmt.Printf("Error received in decoding response from GetUnhealthyIPAddresses: %v\n", err.Error())
		return err
	}
	fmt.Printf("Response for GetUnhealthyIPAddresses: %+v\n", getIPAddressesResponse)
	return nil
}

func main() {
	setEnvironment()
	createNetwork()
	reserveIPAddress("ip0")
	reserveIPAddress("ip0")
	reserveIPAddress("ip0")
	reserveIPAddress("ip0")
	reserveIPAddress("ip1")
	reserveIPAddress("ip2")
	getUnhealthyIPAddresses()
	getIPAddressUtilization()
	releaseIPAddress("ip10")
	getIPAddressUtilization()
	releaseIPAddress("ip2")
	getIPAddressUtilization()
	releaseIPAddress("ip1")
	getIPAddressUtilization()
	releaseIPAddress("ip0")
	getIPAddressUtilization()
	releaseIPAddress("ip0")
	getIPAddressUtilization()
	getHostLocalIP()
	deleteNetwork()
	deleteNetwork()
}
