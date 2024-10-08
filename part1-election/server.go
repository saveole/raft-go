package raft

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"sync"
	"time"
)

// Server wraps a raft.ConsensusModule along with a rpc.Server that exposes its
// methods as RPC endpoints. It also manages the peers of the Raft server. The
// main goal of this type is to simplify the code of raft.Server for
// presentation purposes. raft.ConsensusModule has a *Server to do its peer
// communication and doesn't have to worry about the specifics of running an
// RPC server.
type Server struct {
	mu sync.Mutex

	serverId int
	// 集群内其他机器 id
	peerIds []int

	cm       *ConsensusModule
	rpcProxy *RPCProxy

	rpcServer *rpc.Server
	listener  net.Listener

	peerClients map[int]*rpc.Client

	ready <-chan interface{}
	quit  chan interface{}
	wg    sync.WaitGroup
}

// ConsensusModule RPC 请求的代理封装
// 用于：
// 1. 模拟 RPC 请求过程中的延迟
// 2. 避免 https://github.com/golang/go/issues/19957
// 3. 模拟因网络延迟导致的请求乱序以及丢包的情况(设置了 RAFT_UNRELIABLE_RPC 时)
type RPCProxy struct {
	cm *ConsensusModule
}

func NewServer(serverId int, peerIds []int, ready <-chan interface{}) *Server {
	return &Server{
		serverId:    serverId,
		peerIds:     peerIds,
		ready:       ready,
		quit:        make(chan interface{}),
		peerClients: map[int]*rpc.Client{},
	}
}

func (s *Server) Serve() {
	s.mu.Lock()
	s.cm = NewConsensusModule(s.serverId, s.peerIds, s, s.ready)

	// Create a new RPC server and register a RPCProxy that forwards all methods
	// to n.cm
	s.rpcServer = rpc.NewServer()
	s.rpcProxy = &RPCProxy{cm: s.cm}
	s.rpcServer.RegisterName("ConsensusModule", s.rpcProxy)

	var err error
	s.listener, err = net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[%v] listening at %s\n", s.serverId, s.listener.Addr())
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			conn, err := s.listener.Accept()
			if err != nil {
				select {
				case <-s.quit:
					return
				default:
					log.Fatal("accept err:", err)
				}
			}
			s.wg.Add(1)
			go func() {
				s.rpcServer.ServeConn(conn)
				s.wg.Done()
			}()
		}
	}()
}

// DisconnectAll closes all the client connections to peers for this server.
func (s *Server) DisconnectAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.peerClients {
		if s.peerClients[id] != nil {
			s.peerClients[id].Close()
			s.peerClients[id] = nil
		}
	}
}

// Shutdown closes the server and waits for it to shut down properly.
func (s *Server) Shutdown() {
	s.cm.Stop()
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}

func (s *Server) GetListenAddr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.listener.Addr()
}

func (s *Server) ConnectToPeer(peerId int, addr net.Addr) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.peerClients[peerId] == nil {
		client, err := rpc.Dial(addr.Network(), addr.String())
		if err != nil {
			return err
		}
		s.peerClients[peerId] = client
	}
	return nil
}

func (s *Server) DisconnectPeer(peerId int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.peerClients[peerId] != nil {
		err := s.peerClients[peerId].Close()
		s.peerClients[peerId] = nil
		return err
	}
	return nil
}

func (s *Server) Call(id int, serviceMethod string, args, reply interface{}) error {
	s.mu.Lock()
	peer := s.peerClients[id]
	s.mu.Unlock()

	// If this is called after shutdown (where client.Close is called), it will
	// return an error.
	if peer == nil {
		return fmt.Errorf("call client %d after it's closed", id)
	} else {
		return peer.Call(serviceMethod, args, reply)
	}
}

func (rpc *RPCProxy) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	// 模拟 RPC 请求过程中的延迟和丢包情况
	if len(os.Getenv("RAFT_UNRELIABLE_RPC")) > 0 {
		dice := rand.Intn(10)
		if dice == 9 {
			rpc.cm.dlog("dropping RequestVote")
			return fmt.Errorf("RPC failed")
		} else if dice == 8 {
			rpc.cm.dlog("delaying RequestVote")
			time.Sleep(75 * time.Millisecond)
		}
	} else {
		// 模拟网络延迟
		time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)
	}
	return rpc.cm.RequestVote(args, reply)
}

func (rpc *RPCProxy) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	// 模拟 RPC 请求过程中的延迟和丢包情况
	if len(os.Getenv("RAFT_UNRELIABLE_RPC")) > 0 {
		dice := rand.Intn(10)
		if dice == 9 {
			rpc.cm.dlog("dropping RequestVote")
			return fmt.Errorf("RPC failed")
		} else if dice == 8 {
			rpc.cm.dlog("delaying RequestVote")
			time.Sleep(75 * time.Millisecond)
		}
	} else {
		// 模拟网络延迟
		time.Sleep(time.Duration(1+rand.Intn(5)) * time.Millisecond)
	}
	return rpc.cm.AppendEntries(args, reply)
}
