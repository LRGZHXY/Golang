package raft

// applicationTicker 将已经提交的日志应用到状态机 只有在commitIndex增大后才唤醒
func (rf *Raft) applicationTicker() {
	for !rf.killed() {
		rf.mu.Lock()
		rf.applyCond.Wait() //一直阻塞，直到其他线程调用applyCond.Signal()
		entries := make([]LogEntry, 0)
		snapPendingApply := rf.snapPending //是否有待处理的快照

		if !snapPendingApply {
			///
			if rf.lastApplied < rf.log.snapLastIdx {
				rf.lastApplied = rf.log.snapLastIdx
			}

			// make sure that the rf.log have all the entries
			start := rf.lastApplied + 1
			end := rf.commitIndex
			if end >= rf.log.size() {
				end = rf.log.size() - 1
			}
			for i := start; i <= end; i++ {
				entries = append(entries, rf.log.at(i))
			}
			///
			/*for i := rf.lastApplied + 1; i <= rf.commitIndex; i++ {
				entries = append(entries, rf.log.at(i))
			}*/
		}
		rf.mu.Unlock()

		//在给applyCh发送ApplyMsg时，不要在加锁的情况下进行。因为我们并不知道这个操作会耗时多久（即应用层多久会取走数据）
		if !snapPendingApply {
			for i, entry := range entries {
				rf.applyCh <- ApplyMsg{
					CommandValid: entry.CommandValid,     //日志是否有效
					Command:      entry.Command,          //提交的命令
					CommandIndex: rf.lastApplied + 1 + i, //这条日志在原始日志数组中的索引
				}
			}
		} else { //发送快照
			rf.applyCh <- ApplyMsg{
				SnapshotValid: true,
				Snapshot:      rf.log.snapshot,
				SnapshotIndex: rf.log.snapLastIdx,
				SnapshotTerm:  rf.log.snapLastTerm,
			}
		}

		rf.mu.Lock()
		if !snapPendingApply {
			LOG(rf.me, rf.currentTerm, DApply, "Apply log for [%d,%d]", rf.lastApplied+1, rf.lastApplied+len(entries))
			rf.lastApplied += len(entries)
		} else {
			LOG(rf.me, rf.currentTerm, DApply, "Apply snapshot for [0,%d]", 0, rf.log.snapLastIdx)
			rf.lastApplied = rf.log.snapLastIdx
			if rf.commitIndex < rf.lastApplied {
				rf.commitIndex = rf.lastApplied
			}
			rf.snapPending = false
		}
		rf.mu.Unlock()
	}
}
