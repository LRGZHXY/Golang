package raft

import "fmt"

/*
	日志压缩
*/

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
// Snapshot 生成快照
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (PartD).
	rf.mu.Lock()
	defer rf.mu.Unlock()
	///
	if index > rf.commitIndex {
		LOG(rf.me, rf.currentTerm, DSnap, "Couldn't snapshot before CommitIdx: %d>%d", index, rf.commitIndex)
		return
	}
	if index <= rf.log.snapLastIdx {
		LOG(rf.me, rf.currentTerm, DSnap, "Already snapshot in %d<=%d", index, rf.log.snapLastIdx)
		return
	}
	///
	rf.log.doSnapshot(index, snapshot)
	rf.persistLocked()
}

type InstallSnapshotArgs struct {
	Term     int
	LeaderId int

	LastIncludedIndex int //快照中包含的最后日志索引（压缩断点）
	LastIncludedTerm  int

	Snapshot []byte
}

/*
	Leader-1,T5,Last:[100]T4
	领导者id:1，任期:5，快照包含的最后日志索引:100，任期:4
*/
func (args *InstallSnapshotArgs) String() string {
	return fmt.Sprintf("Leader-%d,T%d,Last:[%d]T%d", args.LeaderId, args.Term, args.LastIncludedIndex, args.LastIncludedTerm)
}

type InstallSnapshotReply struct {
	Term int
}

/*
	T5  Follower回复自己当前的Term
*/
func (reply *InstallSnapshotReply) String() string {
	return fmt.Sprintf("T%d", reply.Term)
}

// InstallSnapshot Leader->Follower 同步快照
func (rf *Raft) InstallSnapshot(args *InstallSnapshotArgs, reply *InstallSnapshotReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	LOG(rf.me, rf.currentTerm, DDebug, "<- S%d,RecvSnapshot,Args=%v", args.LeaderId, args.String())

	reply.Term = rf.currentTerm
	if args.Term < rf.currentTerm {
		LOG(rf.me, rf.currentTerm, DSnap, "<- S%d,Reject Snap,Higher Term,T%d>T%d", args.LeaderId, rf.currentTerm, args.Term)
		return
	}
	if args.Term >= rf.currentTerm { // = 处理节点是candidate的情况
		rf.becomeFollowerLocked(args.Term)
	}

	if rf.log.snapLastIdx >= args.LastIncludedIndex { //当前节点日志已经有了更“新”的快照
		LOG(rf.me, rf.currentTerm, DSnap, "<- S%d,Reject Snap,Already installed:%d>%d", args.LeaderId, rf.log.snapLastIdx, args.LastIncludedIndex)
		return
	}

	rf.log.installSnapshot(args.LastIncludedIndex, args.LastIncludedTerm, args.Snapshot)
	rf.persistLocked()
	rf.snapPending = true
	rf.applyCond.Signal()
}

//  Leader->Follower 同步快照
func (rf *Raft) sendInstallSnapshot(server int, args *InstallSnapshotArgs, reply *InstallSnapshotReply) bool {
	ok := rf.peers[server].Call("Raft.InstallSnapshot", args, reply)
	return ok
}

// installToPeer 向Follower发送快照请求
func (rf *Raft) installToPeer(peer, term int, args *InstallSnapshotArgs) {
	reply := &InstallSnapshotReply{}
	ok := rf.sendInstallSnapshot(peer, args, reply)
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if !ok {
		LOG(rf.me, rf.currentTerm, DLog, "-> S%d,Lost or crashed", peer)
		return
	}
	LOG(rf.me, rf.currentTerm, DDebug, "-> S%d,Append,Reply=%v", peer, reply.String())

	if reply.Term > rf.currentTerm {
		rf.becomeFollowerLocked(reply.Term)
		return
	}

	//检查上下文
	if rf.contextLostLocked(Leader, term) {
		LOG(rf.me, rf.currentTerm, DLog, "-> S%d,Context Lost,T%d:Leader->T%d:%d", peer, term, rf.currentTerm, rf.role)
		return
	}

	if args.LastIncludedIndex > rf.matchIndex[peer] {
		rf.matchIndex[peer] = args.LastIncludedIndex
		rf.nextIndex[peer] = rf.matchIndex[peer] + 1 //下一个待发送的日志条目
	}

	//不需要更新commitIndex，快照已经包含了所有已提交日志的状态，commitIndex只关注快照之后的日志
}
