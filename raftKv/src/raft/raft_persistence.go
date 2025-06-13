package raft

import (
	"bytes"
	"course/labgob"
	"fmt"
)

/*
   持久化
*/

func (rf *Raft) persistString() string {
	//将当前状态格式化为字符串
	return fmt.Sprintf("T%d,VotedFor:%d,Log:[0:%d)", rf.currentTerm, rf.votedFor, rf.log.size())
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
// before you've implemented snapshots, you should pass nil as the
// second argument to persister.Save().
// after you've implemented snapshots, pass the current snapshot
// (or nil if there's not yet a snapshot).

// persistLocked 将当前任期、投票信息和日志保存到稳定存储
func (rf *Raft) persistLocked() {
	// Your code here (PartC).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// raftstate := w.Bytes()
	// rf.persister.Save(raftstate, nil)
	w := new(bytes.Buffer)
	e := labgob.NewEncoder(w)
	//持久化核心字段
	e.Encode(rf.currentTerm)
	e.Encode(rf.votedFor)
	rf.log.persist(e)
	raftstate := w.Bytes()
	rf.persister.Save(raftstate, rf.log.snapshot) //写入稳定存储
	LOG(rf.me, rf.currentTerm, DPersist, "Persist:%v", rf.persistString())
}

// restore previously persisted state.
// readPersist 在节点崩溃重启后，从持久化存储中恢复状态
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (PartC).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
	r := bytes.NewBuffer(data)
	d := labgob.NewDecoder(r)
	var currentTerm int
	//解码currentTerm votedFor log
	if err := d.Decode(&currentTerm); err != nil {
		LOG(rf.me, rf.currentTerm, DPersist, "Read currentTerm error:%v", err)
		return
	}
	rf.currentTerm = currentTerm

	var votedFor int
	if err := d.Decode(&votedFor); err != nil {
		LOG(rf.me, rf.currentTerm, DPersist, "Read votedFor error:%v", err)
		return
	}
	rf.votedFor = votedFor

	if err := rf.log.readPersist(d); err != nil {
		LOG(rf.me, rf.currentTerm, DPersist, "Read log error:%v", err)
		return
	}
	rf.log.snapshot = rf.persister.ReadSnapshot() //读取快照数据

	if rf.log.snapLastIdx > rf.commitIndex { //防止对旧日志重复提交或应用
		rf.commitIndex = rf.log.snapLastIdx
		rf.lastApplied = rf.log.snapLastIdx
	}
	LOG(rf.me, rf.currentTerm, DPersist, "Read from persist:%v", rf.persistString())
}
