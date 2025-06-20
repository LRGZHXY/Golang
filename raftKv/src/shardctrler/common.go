package shardctrler

import (
	"log"
	"time"
)

//
// Shard controler: assigns shards to replication groups.
//
// RPC interface:
// Join(servers) -- add a set of groups (gid -> server-list mapping).
// Leave(gids) -- delete a set of groups.
// Move(shard, gid) -- hand off one shard from current owner to gid.
// Query(num) -> fetch Config # num, or latest config if num==-1.
//
// A Config (configuration) describes a set of replica groups, and the
// replica group responsible for each shard. Configs are numbered. Config
// #0 is the initial configuration, with no groups and all shards
// assigned to group 0 (the invalid group).
//
// You will need to add fields to the RPC argument structs.
//

// NShards The number of shards.
const NShards = 10

// Config A configuration -- an assignment of shards to groups.
// Please don't change this.

/*
	Shards = [2,2,1,1,1,3,3,3,3,3]
              0 1 2 3 4 5 6 7 8 9
	Groups = {
    1: ["server1:port", "server2:port"], // group 1
    2: ["server3:port", "server4:port"], // group 2
    3: ["server5:port", "server6:port"], // group 3
}
*/
type Config struct {
	Num    int              // config number
	Shards [NShards]int     // shard -> gid,表示第i个shard属于哪个group
	Groups map[int][]string // gid -> servers[],表示这个group是哪些服务器组成的
}

func DefaultConfig() Config {
	return Config{
		Groups: make(map[int][]string),
	}
}

const (
	OK             = "OK"
	ErrNoKey       = "ErrNoKey"
	ErrWrongLeader = "ErrWrongLeader"
	ErrTimeout     = "ErrTimeout"
)

type Err string

type JoinArgs struct {
	Servers  map[int][]string // new GID -> servers mappings
	ClientId int64
	SeqId    int64
}

type JoinReply struct {
	WrongLeader bool
	Err         Err
}

type LeaveArgs struct {
	GIDs     []int
	ClientId int64
	SeqId    int64
}

type LeaveReply struct {
	WrongLeader bool
	Err         Err
}

type MoveArgs struct {
	Shard    int
	GID      int
	ClientId int64
	SeqId    int64
}

type MoveReply struct {
	WrongLeader bool
	Err         Err
}

type QueryArgs struct {
	Num int // desired config number
}

type QueryReply struct {
	WrongLeader bool
	Err         Err
	Config      Config
}

const ClientRequestTimeout = 500 * time.Millisecond

const Debug = false

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	// Your definitions here.
	// Field names must start with capital letters,
	// otherwise RPC will break.
	Servers  map[int][]string // Join
	GIDs     []int            // Leave 要移除的group ID列表
	Shard    int              // Move 指定要移动的shard编号
	GID      int              // Move 目标group ID，把Shard移动到这个group
	Num      int              // Query
	OpType   OperationType
	ClientId int64
	SeqId    int64
}

type OpReply struct {
	ControllerConfig Config
	Err              Err
}

type OperationType uint8

const (
	OpJoin OperationType = iota
	OpLeave
	OpMove
	OpQuery
)

type LastOperationInfo struct {
	SeqId int64
	Reply *OpReply
}
