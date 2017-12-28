// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package etcdserver

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/etcd/etcdserver"
	"github.com/coreos/etcd/etcdserver/api/v2http"
	"github.com/coreos/etcd/pkg/transport"
	"github.com/coreos/etcd/pkg/types"
	"github.com/golang/glog"
	"github.com/Azure/azure-container-networking/log"
	"flag"
	"fmt"
)

const(
	snapshotCount        = etcdserver.DefaultSnapCount
	maxSnapshotFiles     = 5
	maxWALFileCount      = 5 // write ahead log file count to retain.
	listnerURLForClients = "http://localhost:4001" // clientURL has listener created and handles etcd API traffic
	listnerURLForPeers   = "http://localhost:7001" // peerURL does't have listener created, it is used to pass Etcd validation
	etcdHealthCheckURL   = listnerURLForClients + "/v2/keys/" // Trailing slash is required,
)

// CreateServer creates etcd server object.
func CreateServer(etcdName string, persistenceDir string) (*EtcdServer, error) {
	log.Printf("[Azure CNS] CreateServer")
	clientURLs, err := types.NewURLs([]string{listnerURLForClients})
	if err != nil {
		log.Printf("Failed to parse listner URL %q: %v", listnerURLForClients, err)
		return nil, err
	}
	
	peerURLs, err := types.NewURLs([]string{listnerURLForPeers})
	if err != nil {
		glog.Fatalf("Failed to parse peer URL %q: %v", listnerURLForPeers, err)
		return nil, err
	}
	
	config := &etcdserver.ServerConfig{
		Name:               etcdName,
		ClientURLs:         clientURLs,
		PeerURLs:           peerURLs,
		DataDir:            persistenceDir,
		InitialPeerURLsMap: map[string]types.URLs{etcdName: peerURLs},
		NewCluster:         true,
		SnapCount:          snapshotCount,
		MaxSnapFiles:       maxSnapshotFiles,
		MaxWALFiles:        maxWALFileCount,
		// TickMs:             keep default, // heartbeat interval default is 100 miliseconds
		// ElectionTicks:      electionTicks, // election timeout default is 1000 miliseconds
	}

	return &EtcdServer{
		config: config,
	}, nil
}

// Start starts the etcd server and listening for client connections.
func (e *EtcdServer) Start(etcdHealthCheckURL string) error { 	
	var err error
	
	e.EtcdServer, err = etcdserver.NewServer(*e.config)
	if err != nil {
		return err
	}

	e.clientListen, err = createListener(e.config.ClientURLs[0])
	if err != nil {
		return err
	}

	e.EtcdServer.Start()

	ch := v2http.NewClientHandler(e.EtcdServer, e.config.ReqTimeout())
	errCh := make(chan error)
	go func(l net.Listener) {
		defer close(errCh)
		srv := &http.Server{
			Handler:     ch,
			ReadTimeout: 5 * time.Minute,
		}
		
		// Serve always returns a non-nil error.
		errCh <- srv.Serve(l)
	}(e.clientListen)

	err = pollErrorChannel("etcd", []string{etcdHealthCheckURL}, errCh)
	if err != nil {
		return err
	}
	return nil
}

// Stop closes all connections and stops the Etcd server
func (e *EtcdServer) Stop() error {
	if e.EtcdServer != nil {
		e.EtcdServer.Stop()
	}

	if e.clientListen != nil {
		err := e.clientListen.Close()
		if err != nil {
			return err
		}
	}
	
	return nil
}

func createListener(url url.URL) (net.Listener, error) {
	l, err := net.Listen("tcp", url.Host)
	if err != nil {
		return nil, err
	}

	l, err = transport.NewKeepAliveListener(l, url.Scheme, &tls.Config{})
	if err != nil {
		return nil, err
	}

	return l, nil
}

func pollErrorChannel(name string, urls []string, errCh <-chan error) error {
	glog.Infof("Running readiness check for service %q", name)
	var serverStartTimeout = flag.Duration("server-start-timeout", time.Second*120, "Time to wait for server to become healthy.")
	endTime := time.Now().Add(*serverStartTimeout)
	blockCh := make(chan error)
	defer close(blockCh)
	for endTime.After(time.Now()) {
		select {

		// Run the health check if there is no error on the channel.		
		case err, ok := <-errCh:
			if ok { // The channel is not closed, this is a real error
				if err != nil { 
					return err
				}				
			} else { 
				// The channel is closed, this is only a zero value.
				// Replace the errCh with blockCh to avoid busy loop,
				// and keep checking readiness.
				errCh = blockCh
			}
		case <-time.After(time.Second):
			ready := true
			for _, url := range urls {
				resp, err := http.Head(url)
				if err != nil || resp.StatusCode != http.StatusOK {
					ready = false
					break
				}
			}
			if ready {
				return nil
			}
		}
	}
	
	return fmt.Errorf("e2e service %q readiness check timeout %v", name, *serverStartTimeout)
}