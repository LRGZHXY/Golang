package raft

import (
	"course/labrpc"
	"sync"
	"sync/atomic"
	"time"
)

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

/*
	Raft结构体定义和一些公共逻辑
*/

const (
	//超时上下界 250~400之间可以保证响应速度快，同时误判和冲突概率低
	electionTimeoutMin time.Duration = 250 * time.Millisecond
	electionTimeoutMax time.Duration = 400 * time.Millisecond
	replicateInterval  time.Duration = 30 * time.Millisecond
)

const (
	InvalidTerm  int = 0
	InvalidIndex int = 0
)

type Role string

const (
	Follower  Role = "Follower"
	Candidate Role = "Candidate"
	Leader    Role = "Leader"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part PartD you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For PartD:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (PartA, PartB, PartC).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.
	role        Role
	currentTerm int
	votedFor    int //-1表示没有投过票

	log *RaftLog

	//Leader独有视图字段
	nextIndex  []int //下一条要发的日志索引
	matchIndex []int //已成功匹配的最大日志索引

	//日志应用
	commitIndex int           //已提交的最大日志索引（所有提交的日志条目都已被多数派复制）
	lastApplied int           //最后被应用到状态机的日志索引（已经对外生效的命令）
	applyCh     chan ApplyMsg //将已经提交的日志条目推送给状态机的通道
	snapPending bool          //是否在处理快照
	applyCond   *sync.Cond    //让日志应用线程在没有新日志时能睡眠等待，避免空转浪费资源

	electionStart   time.Time     //选举周期控制起始点
	electionTimeout time.Duration // 选举超时间隔 随机

}

// 状态转换
func (rf *Raft) becomeFollowerLocked(term int) {
	if term < rf.currentTerm { //当前term更大，不用做状态转换
		LOG(rf.me, rf.currentTerm, DError, "Can't become Follower,lower term: T%d", term)
		return
	}

	LOG(rf.me, rf.currentTerm, DLog, "%s->Follower,For T%v->T%v", rf.role, rf.currentTerm, term)
	rf.role = Follower
	shouldPersist := rf.currentTerm != term //任期变了→ 执行持久化
	if term > rf.currentTerm {
		rf.votedFor = -1 //进入新term 将投票置为空
	}
	rf.currentTerm = term
	if shouldPersist {
		rf.persistLocked()
	}
}

func (rf *Raft) becomeCandidateLocked() {
	if rf.role == Leader {
		LOG(rf.me, rf.currentTerm, DError, "Leader can't become Candidate")
		return
	}

	LOG(rf.me, rf.currentTerm, DVote, "%s->Candidate,For T%d", rf.role, rf.currentTerm+1)
	rf.currentTerm++
	rf.role = Candidate
	rf.votedFor = rf.me
	rf.persistLocked()
}

func (rf *Raft) becomeLeaderLocked() {
	if rf.role != Candidate {
		LOG(rf.me, rf.currentTerm, DError, "Only Candidate can become Leader")
		return
	}

	LOG(rf.me, rf.currentTerm, DLeader, "Become Leader in T%d", rf.currentTerm)
	rf.role = Leader
	for peer := 0; peer < len(rf.peers); peer++ {
		rf.nextIndex[peer] = rf.log.size()
		rf.matchIndex[peer] = 0
	}
}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	// Your code here (PartA).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.role == Leader
}

// GetStateSize 获取持久化状态大小（字节数）
func (rf *Raft) GetRaftStateSize() int {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.persister.RaftStateSize()
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
// Start Leader接收来自上层服务（如键值存储）提交的命令
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.role != Leader { //不是Leader，不能接收新的客户端命令
		return 0, 0, false
	}
	rf.log.append(LogEntry{
		CommandValid: true,
		Command:      command,
		Term:         rf.currentTerm,
	})
	LOG(rf.me, rf.currentTerm, DLeader, "Leader accept log [%d]T%d", rf.log.size()-1, rf.currentTerm)
	rf.persistLocked()

	return rf.log.size() - 1, rf.currentTerm, true
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// 防止过期异步操作污染当前状态
func (rf *Raft) contextLostLocked(role Role, term int) bool { //上下文：角色和任期
	return !(rf.currentTerm == term && rf.role == role) // true：上下文丢失，不能继续处理; false：上下文还在，处理可以继续。
}

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{}
	rf.peers = peers
	rf.persister = persister
	rf.me = me

	// Your initialization code here (PartA, PartB, PartC).
	//初始化
	rf.role = Follower
	rf.currentTerm = 1
	rf.votedFor = -1

	/*
		插入一个 dummy entry（哑元日志项）：
		避免访问日志时处理index == 0的边界情况
		实际日志从索引1开始
	*/
	rf.log = NewLog(InvalidIndex, InvalidTerm, nil, nil)

	rf.nextIndex = make([]int, len(rf.peers))
	rf.matchIndex = make([]int, len(rf.peers))

	rf.applyCh = applyCh
	rf.commitIndex = 0
	rf.lastApplied = 0
	rf.applyCond = sync.NewCond(&rf.mu)
	rf.snapPending = false

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.electionTicker()
	go rf.applicationTicker()

	return rf
}
