package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
)

// DHT struct wraps key value store and node
// to interact with other nodes in the network,
// extending this with buckets to create order of
// nodes in the network and properly partition keys
// across nodes, to provide a distributed hash table
// via the Kademlia protocol.
type Dht struct {
	mtx        sync.Mutex
	ConnClosed chan struct{}
	Done       chan struct{}
	Data       *kvstore
	Node       *Node
	Buckets    [][]*Node
	Listener   net.Listener
}

func (d *Dht) formPingMsg(pong bool) *Message {
	return &Message{
		Type:   PingMsg,
		MsgId:  GenerateMsgId(),
		Sender: d.Node,
		Pong:   pong,
	}
}

func (k *Dht) formFindKeyMsg(key string) *Message {
	return &Message{
		Type:   FindValueMsg,
		MsgId:  GenerateMsgId(),
		Sender: k.Node,
		Key:    []byte(key),
	}
}

func (k *Dht) formStoreMsg(value string) *Message {
	return &Message{
		Type:   StoreMsg,
		MsgId:  GenerateMsgId(),
		Sender: k.Node,
		Key:    Hash([]byte(value)),
		Data:   []byte(value),
	}
}

// Gets the index of the highest bucket which the dist fits into, where
// for bucket with index j, all entries in bucket j must have a distance
// from this node s.t. 2^(j) <= dist < 2^(j+1). We get this by finding
// the most significant differing bit between the two node IDs, because
// this tells us the cap on the distance between the two nodes as some
// value b = 2 ^ n. Then we can easily determine the bucket index by
// specifying an index i where 2^i <= 2^n < 2^(i+1).
// e.g. for nodes with IDs 01000... and 0010....., the most significant
// differing bit is bit 159, which means the max distance (xor product)
// between the two nodes is 2^159 - 1. The bucket index is thus 158, from
// which we subtract 1 to account for 0-indexing
func (d *Dht) getHighestAllowableBucketIndex(otherId []byte) int {
	sameBytes := 0
	for i := 0; i < keysize; i++ {
		if d.Node.Id[i] != otherId[i] {
			sameBits := 0
			// get the first differing bit
			for j := 7; j >= 0; j-- {
				if d.Node.Id[i]&(1<<j) != otherId[i]&(1<<j) {
					return numBuckets - 8*sameBytes - sameBits - 1
				} else {
					sameBits++
				}
			}
		} else {
			sameBytes++
		}
	}

	return 0
}

func (d *Dht) nodeCount() int {
	d.mtx.Lock()
	defer d.mtx.Unlock()
	count := 0
	for _, bucket := range d.Buckets {
		count += len(bucket)
	}

	return count
}

// Helper to place a node in one of the k-buckets, based on
// the distance from this node to the other. The k-buckets are
// sorted by the most to least recent communication with each
// node in the bucket.
func (d *Dht) addToKBucket(other *Node) {
	bucketIndex := d.getHighestAllowableBucketIndex(other.Id)

	writeLog("placing node in bucket %d\n", bucketIndex)

	bucket := d.Buckets[bucketIndex]

	// Move existing entry to the front of the list, if exists
	for i, entry := range bucket {
		if entry.equals(other) {
			d.Buckets[bucketIndex] = append([]*Node{other}, append(bucket[:i], bucket[i+1:]...)...)
		}
	}

	if len(bucket) == maxNodesInBucket {
		// If bucket is full, remove the last entry, replace
		// with the new Node
		d.Buckets[bucketIndex] = append([]*Node{other}, bucket[:len(bucket)-1]...)
	} else {
		d.Buckets[bucketIndex] = append([]*Node{other}, bucket...)
	}
}

func (d *Dht) FindValue(ReqMsg *Message) error {
	writeLog("Serving find value with ID %v\n", ReqMsg.MsgId)
	_ = &Message{
		Type:   FindValueMsg,
		MsgId:  ReqMsg.MsgId,
		Sender: d.Node,
		Value:  nil,
	}

	return nil
}

func (d *Dht) FindNode(ReqMsg *Message) error {
	writeLog("Serving find node with ID %v\n", ReqMsg.MsgId)
	_ = &Message{
		Type:   FindValueMsg,
		MsgId:  ReqMsg.MsgId,
		Sender: d.Node,
		Value:  nil,
	}

	return nil
}

func (d *Dht) Ping(ReqMsg *Message) error {
	writeLog("Serving ping with ID %v\n", ReqMsg.MsgId)

	d.addToKBucket(ReqMsg.Sender)

	if !ReqMsg.Pong {
		// This is ping from another node, so we need to send pong
		resp := &Message{
			Type:   PingMsg,
			MsgId:  ReqMsg.MsgId,
			Sender: d.Node,
			Pong:   true}
		d.sendMessageHost(resp, ReqMsg.Sender.Addr, ReqMsg.Sender.Port)
	} else {
		// This is response to our ping
		writeLog("Pong from %v\n", ReqMsg.Sender.Id)
	}

	return nil
}

func (d *Dht) Store(ReqMsg *Message) error {
	writeLog("Serving store with ID %v\n", ReqMsg.MsgId)
	d.Data.Set(ReqMsg.Key, ReqMsg.Data, ReqMsg.ExpirationTime, ReqMsg.ReplicationInterval)

	_ = &Message{
		Type:          StoreMsg,
		MsgId:         ReqMsg.MsgId,
		Sender:        d.Node,
		Value:         nil,
		KNearestNodes: d.getKNearestNodes(ReqMsg.Key),
	}

	return nil
}

// Return the k nearest nodes by ID to the given key
func (d *Dht) getKNearestNodes(key []byte) []*Node {
	return nil
}

// Sends message to host specified by IP/port
func (d *Dht) sendMessageHost(resp *Message, IP net.IP, port int) error {
	return d.sendMessage(resp, fmt.Sprintf("%s:%s", IP.String(), fmt.Sprint(port)))
}

// Sends message to node receiver
func (d *Dht) sendMessage(msg *Message, addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		writeLog("Error sending back response to message: %d, error: %v\n", msg.MsgId, err)
		return err
	}

	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(*msg)
	if err != nil {
		writeLog("Error encoding response to message: %d, error: %v\n", msg.MsgId, err)
		return err
	}

	conn.Close()
	return nil
}

// Join a Kademlia network, by pinging an existing node, and further
// acquiring a list of nodes in the network to seed the k buckets
func (d *Dht) join(IP net.IP, Port int) error {
	writeLog("Joining the kademlia network with buddy node %s:%d", IP, Port)
	pingMsg := &Message{
		Type:   PingMsg,
		MsgId:  GenerateMsgId(),
		Sender: d.Node,
		Pong:   false,
	}
	return d.sendMessageHost(pingMsg, IP, Port)
}

func (d *Dht) handleConn(conn net.Conn) {
	writeLog("handling connection %v", conn)
	defer func() {
		writeLog("Closing connection")
		conn.Close()
		d.ConnClosed <- struct{}{}
	}()

	// Decode the client message to well-defined message type
	decoder := gob.NewDecoder(conn)
	msg := Message{}
	err := decoder.Decode(&msg)
	if err != nil && err != io.EOF {
		writeLog("Error decoding message %s", err)
		return
	}

	// Route message to appropriate handler
	switch msg.Type {
	case PingMsg:
		d.Ping(&msg)
	case FindValueMsg:
		d.FindValue(&msg)
	case StoreMsg:
		d.Store(&msg)
	case FindNodeMsg:
		d.FindNode(&msg)
	default:
		writeErr("Unrecognized message type %d", msg.Type)
	}
}

func (d *Dht) entry() {
	for {
		conn, err := d.Listener.Accept()
		if err != nil {
			writeLog("Closing server connection %s", err)
			return
		}

		go d.handleConn(conn)
	}
}

func (d *Dht) initListener() {
	var err error
	d.Listener, err = net.Listen("tcp", fmt.Sprintf(":%d", d.Node.Port))
	if err != nil {
		writeLog("Error listening on port %d\n", d.Node.Port)
	}
}

func NewDht() *Dht {
	dht := &Dht{
		Done:       make(chan struct{}),
		ConnClosed: make(chan struct{}),
		Data:       NewKVStore(),
		Node:       NewNode(),
		Buckets:    make([][]*Node, numBuckets),
	}

	// Set up listener and proceed to entry, which is
	// serial loop waiting for connections, and dispatching
	// each to goroutine
	dht.initListener()

	return dht
}

func main() {
	// These flags are used for a server to be added to an existing kademlia
	// network. If they are provided, an initial ping will be sent to the
	// specified server, which will seed this new server with node information
	joinIP := flag.String("joinIP", "", "IP address of joining server")
	joinPort := flag.Int("joinPort", -1, "Port number of joining server")
	loggingEnabled := flag.Bool("loggingEnabled", false, "Enable logging")
	flag.Parse()

	configs.LoggingEnabled = *loggingEnabled
	dht := NewDht()

	go func() {
		defer func() {
			writeLog("Closing server. Goodbye")
			dht.Done <- struct{}{}
		}()

		dht.entry()
	}()

	writeLog("DHT server started at %s\n", dht.Listener.Addr().String())

	if *joinIP != "" && *joinPort != -1 {
		err := dht.join(net.ParseIP(*joinIP), *joinPort)

		if err != nil {
			writeLog("Fatal error attempting to join network %s", err)
		}
	}

	// TODO key clean up thread

	// TODO key republish-replicate thread

	// rpc receive queue thread

	// rpc send queue thread

	// Wait on done channel. This channel is signalled
	// only by user input to bring down the server
	<-dht.Done
}
