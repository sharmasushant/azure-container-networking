// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package restserver

// Container Network Service remote API Contract.
const (
	Success                 = 0
	UnsupportedNetworkType  = 1
	InvalidParameter        = 2
	UnsupportedEnvironment  = 3
	UnreachableHost         = 4
	ReservationNotFound     = 5
	MalformedSubnet         = 8
	UnreachableDockerDaemon = 9
	UnspecifiedNetworkName  = 10
	NotFound                = 14
	AddressUnavailable      = 15
	UnexpectedError         = 99

	OptDisableSnat = "DisableSNAT"
	NetworkMode    = "com.microsoft.azure.network.mode"
)
