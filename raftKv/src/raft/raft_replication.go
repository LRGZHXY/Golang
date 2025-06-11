package raft

import (
	"fmt"
	"sort"
	"time"
)

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

	LeaderCommit int //已经提交到状态机的最大日志索引
}

/*
	Leader-2,T5,Prev:[10]T4,(10,13],CommitIdx:12
	领导者id:2，任期:5，前一个日志索引及任期:[10]T4，本次追加日志索引范围(10,13]，领导者已提交的日志索引:12
*/
func (args *AppendEntriesArgs) String() string {
	return fmt.Sprintf("Leader-%d,T%d,Prev:[%d]T%d,(%d,%d],CommitIdx:%d",
		args.LeaderId, args.Term, args.PrevLogIndex, args.PrevLogTerm,
		args.PrevLogIndex, args.PrevLogIndex+len(args.Entries), args.LeaderCommit)
}

type AppendEntriesReply struct {
	Term    int  //跟随者任期
	Success bool //表示跟随者是否成功接收日志条目（这里是心跳，所以可忽略）

	ConfilictIndex int
	ConfilictTerm  int
}

/*
	T5,Success:false,ConflictTerm:[8]T3
	跟随者任期:5，是否成功接收日志:false，冲突日志索引及任期:[8]T3
*/
func (reply *AppendEntriesReply) String() string {
	return fmt.Sprintf("T%d,Success:%v,ConflictTerm:[%d]T%d",
		reply.Term, reply.Success, reply.ConfilictIndex, reply.ConfilictTerm)
}

// 回调函数 AppendEntries 处理领导者对跟随者发送心跳或日志复制请求
func (rf *Raft) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	LOG(rf.me, rf.currentTerm, DDebug, "<- S%d,Appended,Args=%v", args.LeaderId, args.String())

	reply.Term = rf.currentTerm
	reply.Success = false

	if args.Term < rf.currentTerm { //过时的领导者，直接拒绝
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject log,Higher term,T%d<T%d", args.LeaderId, args.Term, rf.currentTerm)
		return
	}
	if args.Term >= rf.currentTerm {
		rf.becomeFollowerLocked(args.Term)
	}

	defer func() {
		rf.resetElectionTimerLocked() //重置选举计时器
		if !reply.Success {
			LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Follower Conflict:[%d]T%d", args.LeaderId, reply.ConfilictIndex, reply.ConfilictTerm)
			LOG(rf.me, rf.currentTerm, DDebug, "<- S%d,Follower Log=%v", args.LeaderId, rf.logString())
		}
	}()

	if args.PrevLogIndex >= len(rf.log) { //follower日志长度不够
		reply.ConfilictTerm = InvalidTerm
		//冲突索引：为了告诉Leader冲突日志任期从哪个位置开始，Leader可以直接回退到这个索引，提高同步效率
		reply.ConfilictIndex = len(rf.log)
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject Log,Follower log too short,Len:%d < Prev:%d", args.LeaderId, len(rf.log), args.PrevLogIndex)
		return
	}
	if rf.log[args.PrevLogIndex].Term != args.PrevLogTerm { //PrevLogIndex位置的日志项任期不一致
		reply.ConfilictTerm = rf.log[args.PrevLogIndex].Term       //设置冲突任期为跟随者日志该位置的任期
		reply.ConfilictIndex = rf.firstLogFor(reply.ConfilictTerm) //找到冲突任期第一次出现的日志索引
		LOG(rf.me, rf.currentTerm, DLog2, "<- S%d,Reject Log,Prev log not match,[%d]:T%d!=T%d", args.LeaderId, args.PrevLogIndex, rf.log[args.PrevLogIndex].Term, args.PrevLogTerm)
		return
	}

	rf.log = append(rf.log[:args.PrevLogIndex+1], args.Entries...) //追加新日志（将Follower日志中匹配点之后的内容全部替换为Leader的内容）
	rf.persistLocked()
	reply.Success = true
	LOG(rf.me, rf.currentTerm, DLog2, "Follower accept logs:(%d,%d]", args.PrevLogIndex, args.PrevLogIndex+len(args.Entries))

	if args.LeaderCommit > rf.commitIndex {
		LOG(rf.me, rf.currentTerm, DApply, "Follower update the commit index %d->%d", rf.commitIndex, args.LeaderCommit)
		rf.commitIndex = args.LeaderCommit
		//通知applicationTicker()线程：“日志有新提交了，可以把它们发给状态机了”
		rf.applyCond.Signal() //唤醒应用线程
	}
}

// sendAppendEntries 调用Raft.AppendEntries方法 向follower发送日志或心跳请求
func (rf *Raft) sendAppendEntries(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	ok := rf.peers[server].Call("Raft.AppendEntries", args, reply)
	return ok
}

// getMajorityIndexLocked Leader计算可以安全提交的日志索引（大多数节点已复制的最大日志索引）
func (rf *Raft) getMajorityIndexLocked() int {
	tmpIndexes := make([]int, len(rf.peers)) ///
	copy(tmpIndexes, rf.matchIndex)
	sort.Ints(sort.IntSlice(tmpIndexes)) //升序排序
	/*
			tmpIndexes = [4, 7, 8, 9, 9]
			majorityIdx = (5 - 1) / 2 = 4 / 2 = 2
			函数返回 8
		    表示多数节点（节点索引2、3、4）已经复制了日志索引8 --> Leader可以把索引 ≤ 8 的日志项标记为已提交
	*/
	majorityIdx := (len(rf.peers) - 1) / 2 //计算多数节点对应的下标
	LOG(rf.me, rf.currentTerm, DDebug, "Match index after sort:%v,majority[%d]=%d", tmpIndexes, majorityIdx, tmpIndexes[majorityIdx])
	return tmpIndexes[majorityIdx]
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

		if !reply.Success { //日志不一致
			prevIndex := rf.nextIndex[peer]         //下一条尝试发送给Follower的日志索引
			if reply.ConfilictTerm == InvalidTerm { //Follower日志长度不够
				rf.nextIndex[peer] = reply.ConfilictIndex //从冲突索引日志开始同步
			} else {
				firstIndex := rf.firstLogFor(reply.ConfilictTerm)
				if firstIndex != InvalidIndex {
					rf.nextIndex[peer] = firstIndex //Leader从第一次出现冲突任期的位置开始发送 ///
				} else { //Leader日志里没有这个任期的日志
					rf.nextIndex[peer] = reply.ConfilictIndex
				}
			}

			//nextIndex是往前回退，逐步减少索引，确保所有缺失或冲突的日志能被重新发送
			if rf.nextIndex[peer] > prevIndex {
				rf.nextIndex[peer] = prevIndex //保证nextIndex不会增加，只能保持或减小
			}
			LOG(rf.me, rf.currentTerm, DLog, "-> S%d,Not matched at Prev=[%d]T%d,Try next Prev=[%d]T%d",
				peer, args.PrevLogIndex, rf.log[args.PrevLogIndex].Term, rf.nextIndex[peer]-1, rf.log[rf.nextIndex[peer]-1].Term) //
			LOG(rf.me, rf.currentTerm, DDebug, "-> S%d,Leader log=%v", peer, rf.logString())
			return
		}

		//日志复制成功,更新match/next index
		rf.matchIndex[peer] = args.PrevLogIndex + len(args.Entries)
		rf.nextIndex[peer] = rf.matchIndex[peer] + 1

		// 更新commitIndex
		majorityMatched := rf.getMajorityIndexLocked()
		if majorityMatched > rf.commitIndex && rf.log[majorityMatched].Term == rf.currentTerm { //只能提交当前任期的日志
			LOG(rf.me, rf.currentTerm, DApply, "Leader update the commit index %d->%d", rf.commitIndex, majorityMatched)
			rf.commitIndex = majorityMatched
			rf.applyCond.Signal()
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
			LeaderCommit: rf.commitIndex,
		}
		LOG(rf.me, rf.currentTerm, DDebug, "-> S%d, Append,Args=%v", peer, args.String())
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
