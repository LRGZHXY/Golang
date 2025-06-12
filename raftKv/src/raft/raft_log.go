package raft

import (
	"course/labgob"
	"fmt"
)

type RaftLog struct {
	snapLastIdx  int
	snapLastTerm int

	snapshot []byte     //[1,snapLastIdx] 快照数据
	tailLog  []LogEntry //(snapLastIdx,snapLastIdx+len(tailLog)-1]，其中索引snapLastIdx为dummy log entry
}

func NewLog(snapLastIdx, snapLastTerm int, snapshot []byte, entries []LogEntry) *RaftLog {
	rl := &RaftLog{
		snapLastIdx:  snapLastIdx,
		snapLastTerm: snapLastTerm,
		snapshot:     snapshot,
	}

	rl.tailLog = append(rl.tailLog, LogEntry{
		Term: snapLastTerm,
	})
	rl.tailLog = append(rl.tailLog, entries...) //把实际的日志条目追加到dummy之后

	return rl
}

/*
	所有函数都要在加锁的情况下调用
*/

// readPersist 从稳定存储中读取持久化状态
func (rl *RaftLog) readPersist(d *labgob.LabDecoder) error {
	var lastIdx int
	if err := d.Decode(&lastIdx); err != nil {
		return fmt.Errorf("decode last include index failed")
	}
	rl.snapLastIdx = lastIdx

	var lastTerm int
	if err := d.Decode(&lastTerm); err != nil {
		return fmt.Errorf("decode last include term failed")
	}
	rl.snapLastTerm = lastTerm

	var log []LogEntry
	if err := d.Decode(&log); err != nil {
		return fmt.Errorf("decode tail log failed")
	}
	rl.tailLog = log

	return nil
}

// persist 将当前快照和日志编码到持久化存储
func (rl *RaftLog) persist(e *labgob.LabEncoder) {
	e.Encode(rl.snapLastIdx)
	e.Encode(rl.snapLastTerm)
	e.Encode(rl.tailLog)
}

// size 返回日志的长度
func (rl *RaftLog) size() int {
	return rl.snapLastIdx + len(rl.tailLog)
}

// idx 将整个日志的索引转换为tailLog中的索引
func (rl *RaftLog) idx(logicIdx int) int {
	if logicIdx < rl.snapLastIdx || logicIdx >= rl.size() {
		panic(fmt.Sprintf("%d is out of[%d,%d]", logicIdx, rl.snapLastIdx, rl.size()-1))
	}
	return logicIdx - rl.snapLastIdx
}

// at 返回指定索引的日志项
func (rl *RaftLog) at(logicIdx int) LogEntry {
	return rl.tailLog[rl.idx(logicIdx)]
}

// last 返回日志最后一个条目的索引和任期
func (rl *RaftLog) last() (index, term int) {
	i := len(rl.tailLog) - 1
	return rl.snapLastIdx + i, rl.tailLog[i].Term
}

// firstFor 查找日志中第一次出现某个term的日志索引
func (rl *RaftLog) firstFor(term int) int {
	for idx, entry := range rl.tailLog {
		if entry.Term == term {
			return idx + rl.snapLastIdx
		} else if entry.Term > term {
			break
		}
	}
	return InvalidIndex // 0 代表没找到
}

// tail 返回从指定索引开始的日志条目
func (rl *RaftLog) tail(startIdx int) []LogEntry {
	if startIdx >= rl.size() {
		return nil
	}
	return rl.tailLog[rl.idx(startIdx):]
}

// append 向日志末尾追加一个新日志条目
func (rl *RaftLog) append(e LogEntry) {
	rl.tailLog = append(rl.tailLog, e)
}

// appendFrom 删除logicPrevIndex+1之后的旧日志，追加leader发来的新日志
func (rl *RaftLog) appendFrom(logicPrevIndex int, entries []LogEntry) {
	rl.tailLog = append(rl.tailLog[:rl.idx(logicPrevIndex)+1], entries...)
}

/*
eg：Index: 0  1  2  3  4  5
	Term:  1  1  2  2  2  3
	-->  [0,1]T1[2,4]T2[5,5]T3  表示：日志索引0~1是任期1；2~4是任期2；5是任期3
*/
// String 将日志按任期划分，压缩成字符串
func (rl *RaftLog) String() string {
	var terms string
	prevTerm := rl.snapLastTerm //正在处理的日志的任期
	prevStart := rl.snapLastIdx //这段任期开始的索引
	for i := 0; i < len(rl.tailLog); i++ {
		if rl.tailLog[i].Term != prevTerm { //说明前一个任期结束了
			terms += fmt.Sprintf("[%d,%d]T%d", prevStart, rl.snapLastIdx+i-1, prevTerm)
			//更新prevTerm和prevStart为当前日志项
			prevTerm = rl.tailLog[i].Term
			prevStart = i
		}
	}
	terms += fmt.Sprintf("[%d,%d]T%d", prevStart, rl.snapLastIdx+len(rl.tailLog)-1, prevTerm) //最后一段日志
	return terms
}

// doSnapshot 在生成快照后，截断已持久化的旧日志
// App Layer -> Raft Layer
// index: 本次快照所包含的最后一个日志条目的逻辑索引
func (rl *RaftLog) doSnapshot(index int, snapshot []byte) {
	idx := rl.idx(index)

	//更新快照最后的索引和对应的任期号，表示[1, index]这段日志已快照
	rl.snapLastIdx = index
	rl.snapLastTerm = rl.tailLog[idx].Term
	rl.snapshot = snapshot

	newLog := make([]LogEntry, 0, rl.size()-rl.snapLastIdx)
	newLog = append(newLog, LogEntry{
		Term: rl.snapLastTerm, //添加dummy日志条目
	})
	newLog = append(newLog, rl.tailLog[idx+1:]...) //[index+1, last] 保留快照之后的日志
	rl.tailLog = newLog
}

// Raft Layer -> App Layer
// installSnapshot 更新快照相关状态
func (rl *RaftLog) installSnapshot(index, term int, snapshot []byte) {
	rl.snapLastIdx = index
	rl.snapLastTerm = term
	rl.snapshot = snapshot

	newLog := make([]LogEntry, 0, 1)
	newLog = append(newLog, LogEntry{
		Term: rl.snapLastTerm, //添加dummy日志条目
	})
	rl.tailLog = newLog
}
