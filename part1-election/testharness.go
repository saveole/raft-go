// Test harness for writing tests for Raft.
package raft

import (
	"log"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
}

type Harness struct {

	// cluster is a list of all the raft servers participating in a cluster.
	cluster []*Server

	// 集群每个节点的连接状态
	connected []bool

	n int
	t *testing.T
}

// NewHarness creates a nre test Harness, initialized with n server
// connected to each other.
func NewHarness(t *testing.T, n int) *Harness {
	ns := make([]*Server, n)
	connected := make([]bool, n)
	ready := make(chan interface{})

	// create all servers in this cluster, assign ids and peer ids.
	for i := 0; i < n; i++ {
		peerIds := make([]int, n)
		for j := 0; j < n; j++ {
			if j != i {
				peerIds = append(peerIds, j)
			}
		}

		ns[i] = NewServer(i, peerIds, ready)
		ns[i].Serve()
	}

	// Connect all peers to each other.
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j {
				addr := ns[j].GetListenAddr()
				ns[i].ConnectToPeer(j, addr)
			}
		}
		connected[i] = true
	}
	close(ready)

	return &Harness{
		cluster:   ns,
		connected: connected,
		n:         n,
		t:         t,
	}
}

// Shutdown shuts down all servers in the harness and waits for them
// to stop running.
func (h *Harness) Shutdown() {
	for i := 0; i < h.n; i++ {
		h.cluster[i].DisconnectAll()
		h.connected[i] = false
	}
	for i := 0; i < h.n; i++ {
		h.cluster[i].Shutdown()
	}
}

// DisconnectPeer disconnects a peer from all other nodes in the cluster.
func (h *Harness) DisconnectPeer(id int) {
	tlog("Disconnect %d", id)
	h.cluster[id].DisconnectAll()
	for i := 0; i < h.n; i++ {
		if i != id {
			h.cluster[i].DisconnectPeer(id)
		}
	}
	h.connected[id] = false
}

// ReconnectPeer connects a disconnected peer to all other nodes in the cluster.
func (h *Harness) ReconnectPeer(id int) {
	tlog("Reconnect %d", id)
	for i := 0; i < h.n; i++ {
		if i != id {
			if err := h.cluster[id].ConnectToPeer(i, h.cluster[i].GetListenAddr()); err != nil {
				h.t.Fatal(err)
			}
			if err := h.cluster[i].ConnectToPeer(id, h.cluster[id].GetListenAddr()); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	h.connected[id] = true
}

// CheckSingleLeader checks that only a single server thinks it's the leader.
// Returns the leader's id and term. It retries several times if no leader is
// identified yet.
func (h *Harness) CheckSingleLeader() (int, int) {
	for i := 0; i < 5; i++ {
		leaderId := -1
		leaderTerm := -1
		for i := 0; i < h.n; i++ {
			if h.connected[i] {
				_, term, isLeader := h.cluster[i].cm.Report()
				if isLeader {
					if leaderId < 0 {
						leaderId = i
						leaderTerm = term
					} else {
						h.t.Fatalf("both %d and %d think they're leaders", leaderId, i)
					}
				}
			}
		}
		if leaderId >= 0 {
			return leaderId, leaderTerm
		}
		time.Sleep(150 * time.Millisecond)
	}
	h.t.Fatal("No leader")
	return -1, -1
}

// CheckNoLeader checks that no connected server considers itself the leader.
func (h *Harness) CheckNoLeader() {
	for i := 0; i < h.n; i++ {
		if h.connected[i] {
			_, _, isLeader := h.cluster[i].cm.Report()
			if isLeader {
				h.t.Fatalf("server %d claims to be leader: want none", i)
			}
		}
	}
}

func tlog(format string, a ...interface{}) {
	format = "[TEST] " + format
	log.Printf(format, a...)
}

func sleepMs(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}
