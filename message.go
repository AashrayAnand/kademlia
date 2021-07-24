package main

import "time"

type MessageType int

const (
	PingMsg      MessageType = iota
	StoreMsg     MessageType = iota
	FindNodeMsg  MessageType = iota
	FindValueMsg MessageType = iota
)

// Recipient of this message responds back with pong,
// which is same message format, except for setting
// the Pong flag to true.
type Ping struct {
	Type   MessageType
	MsgId  []byte
	Sender *Node
	Pong   bool
}

// Recipient of this message stores the given
// key/value pair in its key-value store.
type Store struct {
	Type                MessageType
	MsgId               []byte
	Sender              *Node
	Key                 []byte
	Data                []byte
	ReplicationInterval time.Duration
	ExpirationTime      time.Time
}

// Recipient of this message returns up to K
// nodes that are closest to the specified target
// node by the sender
type FindNode struct {
	Type   MessageType
	MsgId  []byte
	Sender *Node
	Target *Node
}

// Recipient of this message returns the value for
// the given key (if it has it stored), otherwise
// returns the K closes nodes it knows to the key
type FindValue struct {
	Type   MessageType
	MsgId  []byte
	Sender *Node
	Key    []byte
}
