package raft

import "time"

/*
	日志复制
*/

type LogEntry struct {
	Term         int
	CommandValid bool        //命令是不是需要被应用
	Command      interface{} //操作日志（命令的内容）
}

type AppendEntriesArgs struct {
	Term     int
	LeaderId int

	PrevLogIndex int        //新日志条目的前一个日志条目的索引
	PrevLogTerm  int        //PrevLogIndex对应日志条目的任期号
	Entries      []LogEntry //需要被复制的日志条目（如果为空，则是心跳信号）
}

type AppendEntriesReply struct {
	Term    int  //跟随者任期
	Success bool //表示跟随者是否成功接收日志条目（这里是心跳，所以可忽略）
}

// 回调函数 AppendEntries 处理领导者对跟随者发送心跳或日志复制请求
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	// For debug ///
	LOG(rf.me, rf.currentTerm, DDebug, "<- S%d, Receive log, Prev=[%d]T%d, Len()=%d", args.LeaderId, args.PrevLogIndex, args.PrevLogTerm, len(args.Entries))

	reply.Term = rf.currentTerm
	reply.Success = false

	if args.Term < rf.currentTerm { //过时的领导者，直接拒绝
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject log,Higher term,T%d<T%d", args.LeaderId, args.Term, rf.currentTerm)
		return
	}
	if args.Term >= rf.currentTerm {
		rf.becomeFollowerLocked(args.Term)
	}

	if args.PrevLogIndex >= len(rf.log) { //follower日志长度不够 /// >=!!!
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject Log,Follower log too short,Len:%d <= Prev:%d", args.LeaderId, len(rf.log), args.PrevLogIndex)
		return
	}
	if rf.log[args.PrevLogIndex].Term != args.PrevLogTerm { //PrevLogIndex位置的日志项任期不一致
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject Log,Prev log not match,[%d]:T%d!=T%d", args.LeaderId, args.PrevLogIndex, rf.log[args.PrevLogIndex].Term, args.PrevLogTerm)
		return
	}

	rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...) //追加新日志（将Follower日志中匹配点之后的内容全部替换为Leader的内容）
	reply.Success = true
	LOG(rf.me, rf.currentTerm, DLog2, "Follower accept logs:(%d,%d]", args.PrevLogIndex, args.PrevLogIndex+len(args.Entries))

	// TODO(qtmuniao):handle LeaderCommit

	rf.resetElectionTimerLocked() //重置选举计时器
}

// sendAppendEntries 调用Raft.AppendEntries方法 向follower发送日志或心跳请求
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// startReplication 单轮心跳：对除自己外的所有Peer发送一个心跳RPC
func (rf *Raft) startReplication(term int) bool {
	//单次RPC：对某个Peer来发送心跳，并且处理RPC返回值
	replicateToPeer := func(peer int, args *AppendEntriesArgs) {
		reply := &AppendEntriesReply{}
		ok := rf.sendAppendEntries(peer, args, reply)
		rf.mu.Lock()
		defer rf.mu.Unlock()
		if !ok {
			LOG(rf.me, rf.currentTerm, DLog, "-> S%d,Lost or crashed", peer)
			return
		}
		if reply.Term > rf.currentTerm {
			rf.becomeFollowerLocked(reply.Term)
			return
		}

		if !reply.Success { //日志不一致
			idx, term := args.PrevLogIndex, args.PrevLogTerm ///
			for idx > 0 && rf.log[idx].Term == term {
				idx-- //往前找第一个与当前term不同的位置
			}
			rf.nextIndex[peer] = idx + 1
			LOG(rf.me, rf.currentTerm, DLog, "Not match with S%d in %d,try next=%d", peer, args.PrevLogIndex, rf.nextIndex[peer])
			return
		}

		//日志复制成功,更新match/next index
		rf.matchIndex[peer] = args.PrevLogIndex + len(args.Entries)
		rf.nextIndex[peer] = rf.matchIndex[peer] + 1

		// TODO(qtmuniao):update the commitIndex
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.contextLostLocked(Leader, term) {
		LOG(rf.me, rf.currentTerm, DLeader, "Lost Leader[%d] to %s[T%d]", term, rf.role, rf.currentTerm)
		return false
	}
	for peer := 0; peer < len(rf.peers); peer++ {
		if peer == rf.me {
			rf.matchIndex[peer] = len(rf.log) - 1 //设置为最新日志索引
			rf.nextIndex[peer] = len(rf.log)
			continue
		}

		prevIdx := rf.nextIndex[peer] - 1 //要发送日志的起始位置前的一个位置
		prevTerm := rf.log[prevIdx].Term
		args := &AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			PrevLogIndex: prevIdx,
			PrevLogTerm:  prevTerm,
			Entries:      rf.log[prevIdx+1:],
			///
		}
		///
		go replicateToPeer(peer, args)
	}
	return true
}

// 心跳Ticker：在当选Leader后起一个后台线程，等间隔的发送心跳/复制日志
// replicationTicker 心跳 只有在当前term内能进行日志同步
func (rf *Raft) replicationTicker(term int) {
	for !rf.killed() {
		ok := rf.startReplication(term) //发起一次日志复制（心跳）
		if !ok {
			break
		}
		time.Sleep(replicateInterval)
	}
}
