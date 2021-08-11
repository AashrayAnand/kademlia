package main

import (
	"bytes"
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"time"
)

const (
	KEY_SIZE = 160
)

type Node struct {
	Id   []byte
	Addr net.IP
	Port int
}

// Use UDP dial to get preferred local IP address
// Does not require connection to be established
// to return the IP
func LookupLocalIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// Instantiate new node, generate random ID and get local IP
func NewNode() *Node {
	// Set random seed for ID generation
	rand.Seed(time.Now().UnixNano())
	node := &Node{
		Id:   make([]byte, keysize),
		Port: randomPort(),
	}

	if ip, err := LookupLocalIP(); err == nil {
		node.Addr = ip
	} else {
		node.Addr = net.ParseIP("127.0.0.1")
	}

	// Generate random id.
	rand.Read(node.Id)

	return node
}

func randomPort() int {
	return rand.Intn(1000) + 4000
}

func (n *Node) dummyId() {
	n.Id = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (n *Node) dummyIdWithNthByteSet(index int, filler byte) {
	n.Id = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	n.Id[index] = filler
}

// Get address string from node IP/port
func (n *Node) AddressString() string {
	return fmt.Sprintf("%s:%s", n.Addr.String(), strconv.Itoa(n.Port))
}

// Returns distance between two nodes as byte slice, xoring
// their respective IDs and returning the result
func (n *Node) distance(other *Node) *big.Int {
	return new(big.Int).Xor(new(big.Int).SetBytes(n.Id), new(big.Int).SetBytes(other.Id))
}

// Check node equality, given the identifying triple of ID/IP address/port
func (n *Node) equals(other *Node) bool {
	return bytes.Compare(n.Id, other.Id) == 0 && n.Addr.Equal(other.Addr) && n.Port == other.Port
}
