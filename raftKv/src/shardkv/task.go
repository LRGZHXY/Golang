package shardkv

import (
	"sync"
	"time"
)

// 处理apply任务
func (kv *ShardKV) applyTask() {
	for !kv.killed() {
		select {
		case message := <-kv.applyCh:
			if message.CommandValid {
				kv.mu.Lock()
				// 如果是已经处理过的消息则直接忽略
				if message.CommandIndex <= kv.lastApplied {
					kv.mu.Unlock()
					continue
				}
				kv.lastApplied = message.CommandIndex

				var opReply *OpReply
				raftCommand := message.Command.(RaftCommand)
				if raftCommand.CmdType == ClientOpeartion { //客户端请求Get/Put/Append
					// 取出用户的操作信息
					op := raftCommand.Data.(Op)
					opReply = kv.applyClientOperation(op)
				} else { //配置变更
					opReply = kv.handleConfigChangeMessage(raftCommand)
				}

				// 将结果发送回客户端
				if _, isLeader := kv.rf.GetState(); isLeader {
					notifyCh := kv.getNotifyChannel(message.CommandIndex)
					notifyCh <- opReply
				}

				// 判断是否需要snapshot
				if kv.maxraftstate != -1 && kv.rf.GetRaftStateSize() >= kv.maxraftstate {
					kv.makeSnapshot(message.CommandIndex)
				}

				kv.mu.Unlock()
			} else if message.SnapshotValid {
				kv.mu.Lock()
				kv.restoreFromSnapshot(message.Snapshot) //从快照中恢复状态
				kv.lastApplied = message.SnapshotIndex
				kv.mu.Unlock()
			}
		}
	}
}

// fetchConfigTask 获取当前配置
func (kv *ShardKV) fetchConfigTask() {
	for !kv.killed() {
		if _, isLeader := kv.rf.GetState(); isLeader {
			needFetch := true //需要拉取配置
			kv.mu.Lock()
			// 如果有shard的状态是非Normal的，则说明前一个配置变更的任务正在进行中
			for _, shard := range kv.shards {
				if shard.Status != Normal {
					needFetch = false
					break
				}
			}
			currentNum := kv.currentConfig.Num
			kv.mu.Unlock()

			if needFetch {
				newConfig := kv.mck.Query(currentNum + 1)
				// 传入raft模块进行同步
				if newConfig.Num == currentNum+1 {
					kv.ConfigCommand(RaftCommand{ConfigChange, newConfig}, &OpReply{})
				}
			}
		}
		time.Sleep(FetchConfigInterval)
	}
}

// shardMigrationTask 从其他Group拉取shard数据
func (kv *ShardKV) shardMigrationTask() {
	for !kv.killed() {
		if _, isLeader := kv.rf.GetState(); isLeader {
			kv.mu.Lock()
			// 找到需要迁移进来的shard
			gidToShards := kv.getShardByStatus(MoveIn)
			var wg sync.WaitGroup
			for gid, shardIds := range gidToShards {
				wg.Add(1)
				go func(servers []string, configNum int, shardIds []int) {
					defer wg.Done()
					// 遍历该Group中每一个节点，然后从Leader中读取到对应的shard数据
					getShardArgs := ShardOperationArgs{configNum, shardIds}
					for _, server := range servers {
						var getShardReply ShardOperationReply
						clientEnd := kv.make_end(server)
						ok := clientEnd.Call("ShardKV.GetShardsData", &getShardArgs, &getShardReply)
						// 获取到了shard的数据，执行shard迁移
						if ok && getShardReply.Err == OK {
							kv.ConfigCommand(RaftCommand{ShardMigration, getShardReply}, &OpReply{})
						}
					}
				}(kv.prevConfig.Groups[gid], kv.currentConfig.Num, shardIds)
			}

			kv.mu.Unlock()
			wg.Wait()
		}
		time.Sleep(ShardMigrationInterval)
	}
}

// shardGCTask 通知旧group删除已经迁出的shard数据
func (kv *ShardKV) shardGCTask() {
	for !kv.killed() {
		if _, isLeader := kv.rf.GetState(); isLeader {
			kv.mu.Lock()
			gidToShards := kv.getShardByStatus(GC) //查找状态为GC的shard
			var wg sync.WaitGroup
			for gid, shardIds := range gidToShards {
				wg.Add(1)
				go func(servers []string, configNum int, shardIds []int) {
					wg.Done()
					shardGCArgs := ShardOperationArgs{configNum, shardIds}
					for _, server := range servers {
						var shardGCReply ShardOperationReply
						clientEnd := kv.make_end(server)
						ok := clientEnd.Call("ShardKV.DeleteShardsData", &shardGCArgs, &shardGCReply)
						if ok && shardGCReply.Err == OK { //广播gc成功
							kv.ConfigCommand(RaftCommand{ShardGC, shardGCArgs}, &OpReply{})
						}
					}
				}(kv.prevConfig.Groups[gid], kv.currentConfig.Num, shardIds)
			}
			kv.mu.Unlock()
			wg.Wait()
		}

		time.Sleep(ShardGCInterval)
	}
}

// getShardByStatus 找出当前所有处于指定status状态的shard
func (kv *ShardKV) getShardByStatus(status ShardStatus) map[int][]int {
	gidToShards := make(map[int][]int)
	for i, shard := range kv.shards {
		if shard.Status == status {
			gid := kv.prevConfig.Shards[i] //找到原来所属的Group
			if gid != 0 {
				if _, ok := gidToShards[gid]; !ok {
					gidToShards[gid] = make([]int, 0)
				}
				gidToShards[gid] = append(gidToShards[gid], i)
			}
		}
	}
	return gidToShards
}

// GetShardsData 获取shard数据
func (kv *ShardKV) GetShardsData(args *ShardOperationArgs, reply *ShardOperationReply) {
	// 只需要从Leader获取数据
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	kv.mu.Lock()
	defer kv.mu.Unlock()

	// 当前Group的配置不是所需要的
	if kv.currentConfig.Num < args.ConfigNum {
		reply.Err = ErrNotReady
		return
	}

	// 拷贝shard数据
	reply.ShardData = make(map[int]map[string]string)
	for _, shardId := range args.ShardIds {
		reply.ShardData[shardId] = kv.shards[shardId].copyData()
	}

	// 拷贝去重表数据
	reply.DuplicateTable = make(map[int64]LastOperationInfo)
	for clientId, op := range kv.duplicateTable {
		reply.DuplicateTable[clientId] = op.copyData()
	}

	reply.ConfigNum, reply.Err = args.ConfigNum, OK
}

// DeleteShardsData 删除shard数据
func (kv *ShardKV) DeleteShardsData(args *ShardOperationArgs, reply *ShardOperationReply) {
	// 只需要从Leader获取数据
	if _, isLeader := kv.rf.GetState(); !isLeader {
		reply.Err = ErrWrongLeader
		return
	}

	kv.mu.Lock()
	if kv.currentConfig.Num > args.ConfigNum { //说明该数据已经是旧数据，可以直接删除
		reply.Err = OK
		kv.mu.Unlock()
		return
	}
	kv.mu.Unlock()

	var opReply OpReply
	kv.ConfigCommand(RaftCommand{ShardGC, *args}, &opReply) //向Raft提交命令，让状态机执行ShardGC操作

	reply.Err = opReply.Err
}

func (kv *ShardKV) applyClientOperation(op Op) *OpReply {
	// 判断请求key是否所属当前Group
	if kv.matchGroup(op.Key) {
		var opReply *OpReply
		if op.OpType != OpGet && kv.requestDuplicated(op.ClientId, op.SeqId) { //如果是重复请求，直接返回结果
			opReply = kv.duplicateTable[op.ClientId].Reply
		} else {
			// 将操作应用状态机中
			shardId := key2shard(op.Key)
			opReply = kv.applyToStateMachine(op, shardId)
			if op.OpType != OpGet { //如果是Put/Append操作，需要更新去重表
				kv.duplicateTable[op.ClientId] = LastOperationInfo{
					SeqId: op.SeqId,
					Reply: opReply,
				}
			}
		}
		return opReply
	}
	return &OpReply{Err: ErrWrongGroup}
}
