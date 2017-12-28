// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package etcdserver

import (
	"net"

	"github.com/coreos/etcd/etcdserver"
)

// EtcdServer data object.
type EtcdServer struct {
	*etcdserver.EtcdServer
	config       *etcdserver.ServerConfig
	clientListen net.Listener
}

