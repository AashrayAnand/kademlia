package main

const (
	alpha   = 3
	k       = 20
	keysize = 20
)

// DHT struct wraps key value store and node
// API to interact with other nodes in the network,
// extending this with buckets to create order of
// nodes in the network and properly partition keys
// across nodes, to provide a distributed hash table
type dht struct {
	data    *kvstore
	node    *Node
	buckets [][]*Node
}

func NewDHT() *dht {
	dht := dht{
		data:    NewKVStore(),
		node:    NewNode(),
		buckets: make([][]*Node, keysize),
	}

	return &dht
}

func main() {
	_ = NewDHT()

}
