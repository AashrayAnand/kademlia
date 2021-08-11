package main

import (
	"math/big"
	"testing"
)

// Test the result of getting the highest allowable index
// for varying different node IDs.
func TestGetNodeDistances(t *testing.T) {
	node1 := NewNode()

	// Distance to self should be 0 (bytewise xor of byte slice with self)
	if distToSelf := node1.distance(node1); distToSelf.Cmp(big.NewInt(int64(0))) != 0 {
		t.Errorf("distanceToSelf: expected %d, got %d", 0, distToSelf)
	}

	// Create new node and overwrite ID with that of first node
	node2 := NewNode()
	node2.Id = node1.Id

	// Clear least significant byte of second node ID, now whatever
	// is in LSB for first Node is the only difference in ID
	node2.Id[keysize-1] = 0

	// Distance to self should be 0 (bytewise xor of byte slice with self)
	if dist1To2 := node1.distance(node2); dist1To2.Cmp(big.NewInt(int64(node1.Id[keysize-1]))) != 0 {
		t.Errorf("distanceToSelf: expected %d, got %d", int64(node1.Id[keysize-1]), dist1To2)
	}
}
