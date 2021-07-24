package main

import (
	"bytes"
	"math/rand"
	"net"
)

const (
	KEY_SIZE     = 160
	DEFAULT_PORT = 8888
)

type Node struct {
	Id   []byte
	Addr net.IP
	Port int
}

// Create short lived UDP conn to determine
// preferred outbound local IP addr.
func LookupLocalIP() net.IP {
	conn, err := net.Dial("tcp", "8.8.8.8:80")
	if err != nil {
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

// Instantiate new node, generate random ID and get local IP
func NewNode() *Node {
	node := &Node{
		Id:   make([]byte, keysize),
		Addr: LookupLocalIP(),
		Port: DEFAULT_PORT,
	}

	// Generate random id.
	rand.Read(node.Id)

	return node
}

// Returns distance between two nodes as byte slice, xoring
// each byte of their respective IDs
func (n *Node) distance(other *Node) []byte {
	distance := make([]byte, keysize)
	for i := 0; i < keysize; i++ {
		distance[i] = n.Id[i] ^ other.Id[i]
	}

	return distance
}

func (n *Node) equals(other *Node) bool {
	return bytes.Compare(n.Id, other.Id) == 0 && n.Addr.Equal(other.Addr) && n.Port == other.Port
}
