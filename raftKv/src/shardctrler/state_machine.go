package shardctrler

import "sort"

type CtrlerStateMachine struct {
	Configs []Config
}

func NewCtrlerStateMachine() *CtrlerStateMachine {
	cf := &CtrlerStateMachine{Configs: make([]Config, 1)}
	cf.Configs[0] = DefaultConfig()
	return cf
}

// Query 查询指定编号的配置
func (csm *CtrlerStateMachine) Query(num int) (Config, Err) {
	if num < 0 || num >= len(csm.Configs) { //超出范围，返回最新的配置
		return csm.Configs[len(csm.Configs)-1], OK
	}
	return csm.Configs[num], OK
}

// Join 加入新的Group到集群中（需要处理加入之后的负载均衡问题）
func (csm *CtrlerStateMachine) Join(groups map[int][]string) Err {
	num := len(csm.Configs)
	lastConfig := csm.Configs[num-1]
	// 构建新的配置
	newConfig := Config{
		Num:    num,
		Shards: lastConfig.Shards,
		Groups: copyGroups(lastConfig.Groups),
	}

	// 将新的Group加入到Groups中
	for gid, servers := range groups {
		if _, ok := newConfig.Groups[gid]; !ok { //不存在表示这是一个新加入的组
			newServers := make([]string, len(servers))
			copy(newServers, servers)
			newConfig.Groups[gid] = newServers
		}
	}

	// 构造 gid -> shardid 映射关系
	// shard   gid
	//	 0      1
	//	 1      1
	//	 2      2
	//	 3      2
	//	 4      1
	// 转换后：
	//  gid       shard
	//   1       [0, 1, 4]
	//   2       [2, 3]
	gidToShards := make(map[int][]int)
	for gid := range newConfig.Groups {
		gidToShards[gid] = make([]int, 0) //初始化
	}
	for shard, gid := range newConfig.Shards {
		gidToShards[gid] = append(gidToShards[gid], shard)
	}

	// 进行 shard 迁移
	//   gid         shard
	//    1	     [0, 1, 4, 6]
	//    2	     [2, 3]
	//    3	     [5, 7]
	//    4	     []
	// ------------ 第1次移动 --------------
	//    1	     [1, 4, 6]
	//    2	     [2, 3]
	//    3	     [5, 7]
	//    4	     [0]
	// ------------ 第2次移动 --------------
	//    1	     [4, 6]
	//    2	     [2, 3]
	//    3	     [5, 7]
	//    4	     [0, 1]
	for {
		maxGid, minGid := gidWithMaxShards(gidToShards), gidWithMinShards(gidToShards)
		if maxGid != 0 && len(gidToShards[maxGid])-len(gidToShards[minGid]) <= 1 { //已经负载均衡
			break
		}

		// 最少shard的gid增加一个shard：maxGid中第一个shard移动给minGid
		gidToShards[minGid] = append(gidToShards[minGid], gidToShards[maxGid][0])
		// 最多shard的gid减少一个shard
		gidToShards[maxGid] = gidToShards[maxGid][1:]
	}

	/*
			gidToShards = {1: [0, 3, 6],2: [1, 4, 7],3: [2, 5, 8, 9]}

			newShards = [1, 2, 3, 1, 2, 3, 1, 2, 3, 3]
		                 ↑  ↑  ↑  ↑  ↑  ↑  ↑  ↑  ↑  ↑
		       shard id: 0  1  2  3  4  5  6  7  8  9
	*/
	// 得到新的gid -> shard 信息之后，存储到shards数组中
	var newShards [NShards]int
	for gid, shards := range gidToShards {
		for _, shard := range shards {
			newShards[shard] = gid
		}
	}
	newConfig.Shards = newShards
	csm.Configs = append(csm.Configs, newConfig)

	return OK
}

// Leave 移除Group（需要处理移除之后的负载均衡问题）
func (csm *CtrlerStateMachine) Leave(gids []int) Err {
	num := len(csm.Configs)
	lastConfig := csm.Configs[num-1]
	// 构建新的配置
	newConfig := Config{
		Num:    num,
		Shards: lastConfig.Shards,
		Groups: copyGroups(lastConfig.Groups),
	}

	// 构造 gid -> shard 的映射关系
	gidToShards := make(map[int][]int)
	for gid := range newConfig.Groups {
		gidToShards[gid] = make([]int, 0)
	}
	for shard, gid := range newConfig.Shards {
		gidToShards[gid] = append(gidToShards[gid], shard)
	}

	// 删除对应的gid，并且将对应的shard暂存起来
	/*
		gidToShards = {1: [0, 1],2: [2, 3, 4],3: [5, 6, 7, 8, 9]}
	*/
	var unassignedShards []int
	for _, gid := range gids {
		//如果gid在Group中，则删除掉
		if _, ok := newConfig.Groups[gid]; ok {
			delete(newConfig.Groups, gid)
		}
		// 取出对应的shard
		if shards, ok := gidToShards[gid]; ok {
			unassignedShards = append(unassignedShards, shards...) // [5 6 7 8 9]
			delete(gidToShards, gid)
		}
	}

	var newShards [NShards]int
	// 重新分配被删除的gid对应的shard
	if len(newConfig.Groups) != 0 {
		for _, shard := range unassignedShards {
			minGid := gidWithMinShards(gidToShards)
			gidToShards[minGid] = append(gidToShards[minGid], shard)
			/*
			   gidToShards = {1: [0, 1, 5, 6, 8],2: [2, 3, 4, 7, 9]}
			*/
		}

		// 重新存储 shards 数组
		for gid, shards := range gidToShards {
			for _, shard := range shards {
				newShards[shard] = gid
			}
		}
	}

	// 将配置保存
	newConfig.Shards = newShards
	csm.Configs = append(csm.Configs, newConfig)
	return OK
}

// Move 将shard移动到指定的group中
func (csm *CtrlerStateMachine) Move(shardid, gid int) Err {
	num := len(csm.Configs)
	lastConfig := csm.Configs[num-1]
	// 构建新的配置
	newConfig := Config{
		Num:    num,
		Shards: lastConfig.Shards,
		Groups: copyGroups(lastConfig.Groups),
	}

	newConfig.Shards[shardid] = gid // eg: newConfig.Shards[3] = 5（第3号分片的归属变成group 5）
	csm.Configs = append(csm.Configs, newConfig)
	return OK
}

// copyGroups 深拷贝
func copyGroups(groups map[int][]string) map[int][]string {
	newGroup := make(map[int][]string, len(groups))
	for gid, servers := range groups {
		newServers := make([]string, len(servers))
		copy(newServers, servers) //复制原来切片的内容
		newGroup[gid] = newServers
	}
	return newGroup
}

// gitsWithMaxShards 找到拥有最多shard的gid
func gidWithMaxShards(gidToShars map[int][]int) int {
	/*
		在系统初始化时，所有shard默认归属于gid == 0
		 gidToShars := map[int][]int{0: {0, 1, 2, 3, 4, 5, 6, 7, 8, 9}}
	*/
	if shard, ok := gidToShars[0]; ok && len(shard) > 0 {
		return 0
	}

	/*
	   排序的作用：
	   gidToShards := map[int][]int{ 3: {0, 1}, 1: {2, 3, 4}, 2: {5, 6, 7} }
	   遍历顺序随机，可能是3 1 2，也可能是3 2 1，得到的最大值gid=1/gid=2
	   排序后 gids = [1, 2, 3]，遍历顺序是固定的，得到最大值gid=1
	*/
	// 为了让每个节点在调用的时候获取到的配置是一样的
	// 这里将 gid 进行排序，确保遍历的顺序是确定的
	var gids []int
	for gid := range gidToShars {
		gids = append(gids, gid)
	}
	sort.Ints(gids) //排序

	maxGid, maxShards := -1, -1
	for _, gid := range gids {
		if len(gidToShars[gid]) > maxShards {
			maxGid, maxShards = gid, len(gidToShars[gid])
		}
	}
	return maxGid
}

// gidWithMinShards 找到拥有最少shard的gid
func gidWithMinShards(gidToShars map[int][]int) int {
	var gids []int
	for gid := range gidToShars {
		gids = append(gids, gid)
	}
	sort.Ints(gids)

	minGid, minShards := -1, NShards+1
	for _, gid := range gids {
		if gid != 0 && len(gidToShars[gid]) < minShards { //gid == 0通常表示未分配或无效组
			minGid, minShards = gid, len(gidToShars[gid])
		}
	}
	return minGid
}
