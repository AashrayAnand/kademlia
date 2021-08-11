package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
)

// DHT struct wraps key value store and node
// to interact with other nodes in the network,
// extending this with buckets to create order of
// nodes in the network and properly partition keys
// across nodes, to provide a distributed hash table
// via the Kademlia protocol.
type Dht struct {
	mtx            sync.Mutex
	Data           *kvstore
	Node           *Node
	Buckets        [][]*Node
	LoggingEnabled bool
}

func (d *Dht) formPingMsg(pong bool) *Ping {
	return &Ping{
		Type:   PingMsg,
		MsgId:  GenerateMsgId(),
		Sender: d.Node,
		Pong:   pong,
	}
}

func (k *Dht) formFindKeyMsg(key string) *FindValue {
	return &FindValue{
		Type:   FindValueMsg,
		MsgId:  GenerateMsgId(),
		Sender: k.Node,
		Key:    []byte(key),
	}
}

func (k *Dht) formStoreMsg(value string) *Store {
	return &Store{
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

// FindValue RPC
func (d *Dht) FindValue(ReqMsg *FindValue, Resp *FindValueResult) error {
	writeLog("Serving find RPC with ID %v\n", ReqMsg.MsgId)
	Resp = &FindValueResult{
		Type:   FindValueMsg,
		MsgId:  ReqMsg.MsgId,
		Sender: d.Node,
		Value:  nil,
	}

	// Get value from kvstore if cached locally, otherwise
	// initiate process of finding value via DHT
	if v, err := d.Data.Get(ReqMsg.Key); err != nil {
		Resp.Value = v
	} else {
		return err
	}

	return nil
}

// Ping RPC
func (d *Dht) Ping(ReqMsg *Ping, Resp *Ping) error {
	writeLog("Serving ping RPC with ID %v\n", ReqMsg.MsgId)
	Resp = &Ping{
		Type:   PingMsg,
		MsgId:  ReqMsg.MsgId,
		Sender: d.Node,
		Pong:   true}

	d.addToKBucket(ReqMsg.Sender)

	return nil
}

// Store RPC
func (d *Dht) Store(ReqMsg *Store, resp *StoreResp) error {
	writeLog("Serving store RPC with ID %v\n", ReqMsg.MsgId)
	d.Data.Set(ReqMsg.Key, ReqMsg.Data, ReqMsg.ExpirationTime, ReqMsg.ReplicationInterval)

	resp = &StoreResp{
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

// Join a Kademlia network, by pinging an existing node, and further
// acquiring a list of nodes in the network to seed the k buckets
func (d *Dht) join(IP net.IP, Port int) error {
	client, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%s", IP.String(), fmt.Sprint(Port)))

	if err != nil {
		return err
	}

	pingResp := new(Ping)
	err = client.Call("Dht.Ping", d.formPingMsg(false), pingResp)

	if err != nil {
		return err
	}

	writeLog("Received ping response on join %v", pingResp)

	d.addToKBucket(pingResp.Sender)

	// TODO: seed buckets with more nodes

	return nil
}

func (d *Dht) init() {
	d.Data = NewKVStore()
	d.Node = NewNode()
	d.Buckets = make([][]*Node, numBuckets)
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
	dht := new(Dht)
	dht.init()
	rpc.Register(&dht)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", fmt.Sprint(randomPort())))
	if err != nil {
		log.Fatal("dialing:", err)
	}
	rpc.Accept(l)

	if *joinIP != "" && *joinPort != -1 {
		err := dht.join(net.ParseIP(*joinIP), *joinPort)

		if err != nil {
			log.Fatal("Fatal error attempting to join network ", err)
		}
	}

	// key clean up thread

	// key republish-replicate thread

	// rpc receive queue thread

	// rpc send queue thread

	fmt.Printf("Starting up svr at %s\n", l.Addr().String())
	http.Serve(l, nil)
}
