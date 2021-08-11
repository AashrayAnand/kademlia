package main

import "time"

type MessageType int

const (
	PingMsg      MessageType = 1
	StoreMsg     MessageType = 2
	FindNodeMsg  MessageType = 3
	FindValueMsg MessageType = 4
)

// Recipient of this message responds back with pong,
// which is same message format, except for setting
// the Pong flag to true.
type Message struct {
	Type                MessageType
	MsgId               []byte
	Sender              *Node
	Key                 []byte
	Data                []byte
	ReplicationInterval time.Duration
	Value               []byte
	ExpirationTime      time.Time
	Pong                bool
	KNearestNodes       []*Node
}
