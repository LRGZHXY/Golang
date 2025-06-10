package raft

// applicationTicker 将已经提交的日志应用到状态机 只有在commitIndex增大后才唤醒
func (rf *Raft) applicationTicker() {
	for !rf.killed() {
		rf.mu.Lock()
		rf.applyCond.Wait() //一直阻塞，直到其他线程调用applyCond.Signal()
		entries := make([]LogEntry, 0)
		for i := rf.lastApplied + 1; i <= rf.commitIndex; i++ {
			entries = append(entries, rf.log[i])
		}
		rf.mu.Unlock()

		//在给applyCh发送ApplyMsg时，不要在加锁的情况下进行。因为我们并不知道这个操作会耗时多久（即应用层多久会取走数据）
		for i, entry := range entries {
			rf.applyCh <- ApplyMsg{
				CommandValid: entry.CommandValid,     //日志是否有效
				Command:      entry.Command,          //提交的命令
				CommandIndex: rf.lastApplied + 1 + i, //这条日志在原始日志数组中的索引
			}
		}

		rf.mu.Lock()
		LOG(rf.me, rf.currentTerm, DApply, "Apply log for [%d,%d]", rf.lastApplied+1, rf.lastApplied+len(entries))
		rf.lastApplied += len(entries)
		rf.mu.Unlock()
	}
}
