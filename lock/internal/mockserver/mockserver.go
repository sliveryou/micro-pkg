// Copyright 2018 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mockserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"

	pb "go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
)

// MockServer provides a mocked out grpc server of the etcdserver interface.
type MockServer struct {
	ln         net.Listener
	Network    string
	Address    string
	GrpcServer *grpc.Server
}

func (ms *MockServer) ResolverAddress() resolver.Address {
	switch ms.Network {
	case "unix":
		return resolver.Address{Addr: fmt.Sprintf("unix://%s", ms.Address)}
	case "tcp":
		return resolver.Address{Addr: ms.Address}
	default:
		panic("illegal network type: " + ms.Network)
	}
}

// MockServers provides a cluster of mocket out gprc servers of the etcdserver interface.
type MockServers struct {
	mu      sync.RWMutex
	Servers []*MockServer
	wg      sync.WaitGroup
}

// StartMockServers creates the desired count of mock servers
// and starts them.
func StartMockServers(count int) (ms *MockServers, err error) {
	return StartMockServersOnNetwork(count, "tcp")
}

// StartMockServersOnNetwork creates mock servers on either 'tcp' or 'unix' sockets.
func StartMockServersOnNetwork(count int, network string) (ms *MockServers, err error) {
	switch network {
	case "tcp":
		return startMockServersTcp(count)
	case "unix":
		return startMockServersUnix(count)
	default:
		return nil, fmt.Errorf("unsupported network type: %s", network)
	}
}

func startMockServersTcp(count int) (ms *MockServers, err error) {
	addrs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		addrs = append(addrs, "localhost:0")
	}
	return startMockServers("tcp", addrs)
}

func startMockServersUnix(count int) (ms *MockServers, err error) {
	dir := os.TempDir()
	addrs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		f, err := os.CreateTemp(dir, "etcd-unix-so-")
		if err != nil {
			return nil, fmt.Errorf("failed to allocate temp file for unix socket: %v", err)
		}
		fn := f.Name()
		err = os.Remove(fn)
		if err != nil {
			return nil, fmt.Errorf("failed to remove temp file before creating unix socket: %v", err)
		}
		addrs = append(addrs, fn)
	}
	return startMockServers("unix", addrs)
}

func startMockServers(network string, addrs []string) (ms *MockServers, err error) {
	ms = &MockServers{
		Servers: make([]*MockServer, len(addrs)),
		wg:      sync.WaitGroup{},
	}
	defer func() {
		if err != nil {
			ms.Stop()
		}
	}()
	for idx, addr := range addrs {
		ln, err := net.Listen(network, addr)
		if err != nil {
			return nil, fmt.Errorf("failed to listen %v", err)
		}
		ms.Servers[idx] = &MockServer{ln: ln, Network: network, Address: ln.Addr().String()}
		ms.StartAt(idx)
	}
	return ms, nil
}

// StartAt restarts mock server at given index.
func (ms *MockServers) StartAt(idx int) (err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Servers[idx].ln == nil {
		ms.Servers[idx].ln, err = net.Listen(ms.Servers[idx].Network, ms.Servers[idx].Address)
		if err != nil {
			return fmt.Errorf("failed to listen %v", err)
		}
	}

	svr := grpc.NewServer()
	pb.RegisterKVServer(svr, &mockKVServer{})
	pb.RegisterLeaseServer(svr, &mockLeaseServer{})
	ms.Servers[idx].GrpcServer = svr

	ms.wg.Add(1)
	go func(svr *grpc.Server, l net.Listener) {
		svr.Serve(l)
	}(ms.Servers[idx].GrpcServer, ms.Servers[idx].ln)
	return nil
}

// StopAt stops mock server at given index.
func (ms *MockServers) StopAt(idx int) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.Servers[idx].ln == nil {
		return
	}

	ms.Servers[idx].GrpcServer.Stop()
	ms.Servers[idx].GrpcServer = nil
	ms.Servers[idx].ln = nil
	ms.wg.Done()
}

// Stop stops the mock server, immediately closing all open connections and listeners.
func (ms *MockServers) Stop() {
	for idx := range ms.Servers {
		ms.StopAt(idx)
	}
	ms.wg.Wait()
}

type mockKVServer struct{}

func (m *mockKVServer) Range(context.Context, *pb.RangeRequest) (*pb.RangeResponse, error) {
	return &pb.RangeResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (m *mockKVServer) Put(context.Context, *pb.PutRequest) (*pb.PutResponse, error) {
	return &pb.PutResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (m *mockKVServer) DeleteRange(context.Context, *pb.DeleteRangeRequest) (*pb.DeleteRangeResponse, error) {
	return &pb.DeleteRangeResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (m *mockKVServer) Txn(context.Context, *pb.TxnRequest) (*pb.TxnResponse, error) {
	return &pb.TxnResponse{
		Header:    &pb.ResponseHeader{},
		Succeeded: true,
		Responses: []*pb.ResponseOp{
			{
				Response: &pb.ResponseOp_ResponseRange{
					ResponseRange: &pb.RangeResponse{
						Header: &pb.ResponseHeader{},
						Kvs:    []*mvccpb.KeyValue{{}},
					},
				},
			},
			{
				Response: &pb.ResponseOp_ResponseRange{
					ResponseRange: &pb.RangeResponse{
						Header: &pb.ResponseHeader{},
						Kvs:    []*mvccpb.KeyValue{{}},
					},
				},
			},
		},
	}, nil
}

func (m *mockKVServer) Compact(context.Context, *pb.CompactionRequest) (*pb.CompactionResponse, error) {
	return &pb.CompactionResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (m *mockKVServer) Lease(context.Context, *pb.LeaseGrantRequest) (*pb.LeaseGrantResponse, error) {
	return &pb.LeaseGrantResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

type mockLeaseServer struct{}

func (s mockLeaseServer) LeaseGrant(context.Context, *pb.LeaseGrantRequest) (*pb.LeaseGrantResponse, error) {
	return &pb.LeaseGrantResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (s *mockLeaseServer) LeaseRevoke(context.Context, *pb.LeaseRevokeRequest) (*pb.LeaseRevokeResponse, error) {
	return &pb.LeaseRevokeResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (s *mockLeaseServer) LeaseKeepAlive(pb.Lease_LeaseKeepAliveServer) error {
	return nil
}

func (s *mockLeaseServer) LeaseTimeToLive(context.Context, *pb.LeaseTimeToLiveRequest) (*pb.LeaseTimeToLiveResponse, error) {
	return &pb.LeaseTimeToLiveResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}

func (s *mockLeaseServer) LeaseLeases(context.Context, *pb.LeaseLeasesRequest) (*pb.LeaseLeasesResponse, error) {
	return &pb.LeaseLeasesResponse{
		Header: &pb.ResponseHeader{},
	}, nil
}
