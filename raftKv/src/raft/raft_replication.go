package raft

import "time"

/*
	日志复制
*/

type AppendEntriesArgs struct {
	Term     int
	LeaderId int
}

type AppendEntriesReply struct {
	Term    int  //跟随者任期
	Success bool //表示跟随者是否成功接收日志条目（这里是心跳，所以可忽略）
}

// 回调函数 AppendEntries 处理领导者对跟随者发送心跳或日志复制请求
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm     ///
	reply.Success = false           ///
	if args.Term < rf.currentTerm { //过时的领导者，直接拒绝
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject log,Higher term,T%d<T%d", args.LeaderId, args.Term, rf.currentTerm)
		return
	}
	if args.Term >= rf.currentTerm {
		rf.becomeFollowerLocked(args.Term)
	}
	rf.resetElectionTimerLocked() //重置选举计时器
	reply.Success = true          ///
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
	}

	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.contextLostLocked(Leader, term) {
		LOG(rf.me, rf.currentTerm, DLeader, "Lost Leader[%d] to %s[T%d]", term, rf.role, rf.currentTerm)
		return false
	}
	for peer := 0; peer < len(rf.peers); peer++ {
		if peer == rf.me {
			continue
		}
		args := &AppendEntriesArgs{
			Term:     rf.currentTerm,
			LeaderId: rf.me,
		}
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
