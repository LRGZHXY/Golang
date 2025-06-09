package raft

import (
	"math/rand"
	"time"
)

/*
	领导者选举
*/

// resetElectionTimerLocked 重置选举计时器
func (rf *Raft) resetElectionTimerLocked() {
	rf.electionStart = time.Now()
	randRange := int64(electionTimeoutMax - electionTimeoutMin)                     //150
	rf.electionTimeout = electionTimeoutMin + time.Duration(rand.Int63()%randRange) //保证生成值在0~150之间(250~400)
}

// isElectionTimeoutLocked 检查是否超时
func (rf *Raft) isElectionTimeoutLocked() bool {
	return time.Since(rf.electionStart) > rf.electionTimeout
}

// 比较规则：1.Term高者更新	2.Term相同，Index大者更新
// isMoreUpToDate 检查自己本身的日志和候选者日志谁更新
func (rf *Raft) isMoreUpToDateLocked(candidateIndex, candidateTerm int) bool {
	l := len(rf.log)
	lastIndex, lastTerm := l-1, rf.log[l-1].Term

	LOG(rf.me, rf.currentTerm, DVote, "Compore last log,Me:[%d]T%d,Candidate:[%d]T%d", lastIndex, lastTerm, candidateIndex, candidateTerm)
	if lastTerm != candidateTerm {
		return lastTerm > candidateTerm //true:自己任期更大->不投票
	}
	return lastIndex > candidateIndex //true:自己日志更长->不投票
}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!
//投票请求参数
type RequestVoteArgs struct {
	// Your data here (PartA, PartB).
	Term        int //候选人任期
	CandidateId int

	LastLogIndex int //候选人最后一条日志的索引
	LastLogTerm  int //候选人最后一条日志的任期
}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
//投票回复参数
type RequestVoteReply struct {
	// Your data here (PartA).
	Term        int
	VoteGranted bool //是否投票
}

// example RequestVote RPC handler.
// RequestVote 候选人请求投票
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (PartA, PartB).
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	if args.Term < rf.currentTerm { //候选者任期<当前节点任期
		LOG(rf.me, rf.currentTerm, DVote, "-> S%d,Reject voted,Higher term,T%d>T%d", args.CandidateId, rf.currentTerm, args.Term)
		return
	}
	if args.Term > rf.currentTerm {
		rf.becomeFollowerLocked(args.Term) //将自己转换为follower
	}

	if rf.votedFor != -1 { //当前任期已经投过票
		LOG(rf.me, rf.currentTerm, DVote, "-> S%d,Reject voted,Already voted to S%d", args.CandidateId, rf.votedFor)
		return
	}

	if rf.isMoreUpToDateLocked(args.LastLogIndex, args.LastLogTerm) { //候选人日志不如自己新，拒绝投票
		LOG(rf.me, rf.currentTerm, DVote, "-> S%d,Reject voted,Candidate's log less up-to-date", args.CandidateId)
		return
	}

	reply.VoteGranted = true
	rf.votedFor = args.CandidateId
	rf.resetElectionTimerLocked()
	LOG(rf.me, rf.currentTerm, DVote, "-> S%d,Vote granted", args.CandidateId)
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// startElection 发起选举
func (rf *Raft) startElection(term int) {
	votes := 0
	askVoteFromPeer := func(peer int, args *RequestVoteArgs) {
		reply := &RequestVoteReply{}
		ok := rf.sendRequestVote(peer, args, reply) //向其他节点请求投票

		//handle the response
		rf.mu.Lock()
		defer rf.mu.Unlock()
		if !ok {
			LOG(rf.me, rf.currentTerm, DDebug, "Ask vote from S%d,Lost or error", peer)
			return
		}

		if reply.Term > rf.currentTerm { //对方任期更大，自己变成follower
			rf.becomeFollowerLocked(reply.Term)
			return
		}

		if rf.contextLostLocked(Candidate, term) { //上下文失效
			LOG(rf.me, rf.currentTerm, DVote, "Lost context,abort RequestVoteReply for S%d", peer)
			return
		}

		if reply.VoteGranted {
			votes++
			if votes > len(rf.peers)/2 { //票数超过多数节点
				rf.becomeLeaderLocked() //变成leader
				go rf.replicationTicker(term)
			}
		}
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.contextLostLocked(Candidate, term) {
		LOG(rf.me, rf.currentTerm, DVote, "Lost Candidate[T%d] to %s[T%d], abort RequestVote", rf.role, term, rf.currentTerm) ///
		return
	}

	l := len(rf.log)
	for peer := 0; peer < len(rf.peers); peer++ {
		if peer == rf.me {
			votes++ //自己默认给自己投票
			continue
		}

		args := &RequestVoteArgs{
			Term:         rf.currentTerm,
			CandidateId:  rf.me,
			LastLogIndex: l - 1,
			LastLogTerm:  rf.log[l-1].Term,
		}
		go askVoteFromPeer(peer, args) //向其他节点发出投票请求
	}
}

// electionticker 定时检查是否需要发起选举
func (rf *Raft) electionTicker() {
	for !rf.killed() {

		// Your code here (PartA)
		// Check if a leader election should be started.
		rf.mu.Lock()
		if rf.role != Leader && rf.isElectionTimeoutLocked() {
			//不是leader并且已经超时，变成候选者
			rf.becomeCandidateLocked()
			go rf.startElection(rf.currentTerm) //开始选举
		}
		rf.mu.Unlock()

		// pause for a random amount of time between 50 and 350
		// milliseconds.
		ms := 50 + (rand.Int63() % 300)
		time.Sleep(time.Duration(ms) * time.Millisecond) //休眠一段随机时间(选举超时是一个随机值)
	}
}
