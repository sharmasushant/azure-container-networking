// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	ipamOverlay "github.com/Azure/azure-container-networking/cnm/ipam_overlay"
	"github.com/Azure/azure-container-networking/cns/common"
	"github.com/Azure/azure-container-networking/cns/restserver"
	acn "github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/platform"
	"github.com/Azure/azure-container-networking/store"
)

const (
	// Service name.
	name     = "azure-cns"
	ipamName = "overlay-ipam"
)

// Version is populated by make during build.
var version string

// Command line arguments for CNM plugin.
var args = acn.ArgumentList{
	{
		Name:         acn.OptEnvironment,
		Shorthand:    acn.OptEnvironmentAlias,
		Description:  "Set the operating environment",
		Type:         "string",
		DefaultValue: "null",
		ValueMap: map[string]interface{}{
			"null": 0,
		},
	},
	{
		Name:         acn.OptAPIServerURL,
		Shorthand:    acn.OptAPIServerURLAlias,
		Description:  "Set the API server URL",
		Type:         "string",
		DefaultValue: "",
	},
	{
		Name:         acn.OptLogLevel,
		Shorthand:    acn.OptLogLevelAlias,
		Description:  "Set the logging level",
		Type:         "int",
		DefaultValue: acn.OptLogLevelInfo,
		ValueMap: map[string]interface{}{
			acn.OptLogLevelInfo:  log.LevelInfo,
			acn.OptLogLevelDebug: log.LevelDebug,
		},
	},
	{
		Name:         acn.OptLogTarget,
		Shorthand:    acn.OptLogTargetAlias,
		Description:  "Set the logging target",
		Type:         "int",
		DefaultValue: acn.OptLogTargetFile,
		ValueMap: map[string]interface{}{
			acn.OptLogTargetSyslog: log.TargetSyslog,
			acn.OptLogTargetStderr: log.TargetStderr,
			acn.OptLogTargetFile:   log.TargetLogfile,
		},
	},
	{
		Name:         acn.OptVersion,
		Shorthand:    acn.OptVersionAlias,
		Description:  "Print version information",
		Type:         "bool",
		DefaultValue: false,
	},
}

// Prints description and version information.
func printVersion() {
	fmt.Printf("Azure Container Network Service\n")
	fmt.Printf("Version %v\n", version)
}

// Main is the entry point for CNS.
func main() {
	// Initialize and parse command line arguments.
	acn.ParseArgs(&args, printVersion)

	url := acn.GetArg(acn.OptAPIServerURL).(string)
	logLevel := acn.GetArg(acn.OptLogLevel).(int)
	logTarget := acn.GetArg(acn.OptLogTarget).(int)
	vers := acn.GetArg(acn.OptVersion).(bool)

	if vers {
		printVersion()
		os.Exit(0)
	}

	// Initialize ipam.

	var pluginConfig acn.PluginConfig
	var err error
	pluginConfig.Version = version

	// Create a channel to receive unhandled errors from the plugins.
	pluginConfig.Store, err = store.NewJsonFileStore(platform.RuntimePath + ipamName + ".json")
	if err != nil {
		fmt.Printf("Failed to create store: %v\n", err)
		return
	}

	// Create IPAM plugin.
	ipamPlugin, err := ipamOverlay.NewPlugin(&pluginConfig)
	if err != nil {
		fmt.Printf("Failed to create IPAM plugin, err:%v.\n", err)
		return
	}

	// Create logging provider.
	log.SetName(name)
	log.SetLevel(logLevel)
	err = log.SetTarget(logTarget)
	if err != nil {
		fmt.Printf("Failed to configure logging: %v\n", err)
		return
	}

	// Log platform information.
	log.Printf("Running on %v", platform.GetOSInfo())

	// Set CNS options.
	environment := acn.GetArg(acn.OptEnvironment).(string)
	ipamQueryInterval, _ := acn.GetArg(acn.OptIpamQueryInterval).(int)

	ipamPlugin.SetOption(acn.OptEnvironment, environment)
	//ipamPlugin.SetOption(acn.OptAPIServerURL, url)
	ipamPlugin.SetOption(acn.OptIpamQueryInterval, ipamQueryInterval)
	if ipamPlugin != nil {
		err = ipamPlugin.Start(&pluginConfig)
		if err != nil {
			fmt.Printf("Failed to start IPAM plugin, err:%v.\n", err)
			return
		}
	}

	//Initialize CNS
	var config common.ServiceConfig
	config.Version = version
	config.Name = name

	// Create a channel to receive unhandled errors from CNS.
	config.ErrChan = make(chan error, 1)

	// Create the key value store.
	config.Store, err = store.NewJsonFileStore(platform.RuntimePath + name + ".json")
	if err != nil {
		fmt.Printf("Failed to create store: %v\n", err)
		return
	}

	// Create CNS object.
	httpRestService, err := restserver.NewHTTPRestService(&config)
	if err != nil {
		fmt.Printf("Failed to create CNS object, err:%v.\n", err)
		return
	}

	httpRestService.SetOption(acn.OptAPIServerURL, url)

	// Start CNS.
	if httpRestService != nil {
		err = httpRestService.Start(&config)
		if err != nil {
			fmt.Printf("Failed to start CNS, err:%v.\n", err)
			//return
		}
	}

	// Relay these incoming signals to OS signal channel.
	osSignalChannel := make(chan os.Signal, 1)
	signal.Notify(osSignalChannel, os.Interrupt, os.Kill, syscall.SIGTERM)

	// Wait until receiving a signal.
	select {
	case sig := <-osSignalChannel:
		log.Printf("CNS Received OS signal <" + sig.String() + ">, shutting down.")
	case err := <-config.ErrChan:
		log.Printf("CNS Received unhandled error %v, shutting down.", err)
	}

	// Cleanup.
	if httpRestService != nil {
		httpRestService.Stop()
	}
	if ipamPlugin != nil {
		ipamPlugin.Stop()
	}
}
