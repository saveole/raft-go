// Core Raft implementation.
package raft

import (
	"fmt"
	"sync"
	"time"
)

const DebugCM = 1

type LogEntry struct {
	Command interface{}
	Term    int
}

type CMState int

const (
	Follower CMState = iota
	Candidate
	Leader
	Dead
)

func (s CMState) String() string {
	switch s {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	case Dead:
		return "Dead"
	default:
		panic("Unreachable state")
	}
}

// ConsensusModule (CM) implements a single node of Raft consensus.
type ConsensusModule struct {

	// Mutex to protect shared access to this struct.
	mu sync.Mutex

	// server id
	id int

	// peer ids in this cluster
	peerIds []int

	// 包含此 ConsensusModule 模块的 Server 实例,用于 issue Rpc call
	server *Server

	// cluster 中所有节点都保存的 Raft state - Persistent
	currentTerm int
	votedFor    int
	log         []LogEntry

	// cluster 中所有节点都临时保存的 Raft state - Volatile
	state              CMState
	electionResetEvent time.Time
}

// 当 cluster 集群所有节点都 connected 的时候 signal ready channel,
// 然后此 cm 实例就可以重置并运行 election timer.
func NewConsensusModule(id int, peerIds []int, server *Server, ready <-chan interface{}) *ConsensusModule {
	cm := new(ConsensusModule)
	cm.id = id
	cm.peerIds = peerIds
	cm.server = server
	cm.state = Follower
	cm.votedFor = -1

	go func() {
		// The CM is quiescent until ready is signaled; then, it starts a countdown
		// for leader election.
		<-ready
		cm.mu.Lock()
		cm.electionResetEvent = time.Now()
		cm.mu.Unlock()
		cm.runElectionTimer()
	}()

	return cm
}

func (cm *ConsensusModule) runElectionTimer() {

	// TODO: Implement me!
}

func (cm *ConsensusModule) Stop() {
	cm.mu.Unlock()
	defer cm.mu.Unlock()
	cm.state = Dead
	cm.dlog("becomes Dead")
}

// dlog logs a debugging message if DebugCM > 0.
func (cm *ConsensusModule) dlog(format string, args ...interface{}) {
	if DebugCM > 0 {
		format = fmt.Sprintf("[%d] ", cm.id) + format
		fmt.Printf(format, args...)
	}
}

// 选举投票请求参数
type RequestVoteArgs struct {
	// 本次选举轮次
	Term int
	// 候选者 id
	CandidateId  int
	LastLogIndex int
	// 上次记录的选举轮次
	LastLogTerm int
}

type RequestVoteReply struct {
	// 投票是否被采用
	VoteGranted bool
	// 投票轮次
	Term int
}

func (cm *ConsensusModule) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	cm.mu.Unlock()
	defer cm.mu.Unlock()
	// TODO: Implement me!
	return nil
}

type AppendEntriesArgs struct {
	Term     int
	LeaderId int

	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	CommitIndex  int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (cm *ConsensusModule) AppendEntries(args AppendEntriesArgs, reply *AppendEntriesReply) error {
	// TODO: Implement me!
	return nil
}
