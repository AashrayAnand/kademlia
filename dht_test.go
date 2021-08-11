package main

import (
	"testing"
)

func TestAddNodesUpToMaxInBucket(t *testing.T) {
	table1 := NewDht()
	table1.Node.dummyId()

	for i := 0; i < maxNodesInBucket; i++ {
		nodei := NewNode()
		nodei.dummyId() // set same ID for all nodes to ensure placed in same bucket
		table1.addToKBucket(nodei)
	}

	if count := table1.nodeCount(); count != maxNodesInBucket {
		t.Errorf("Node count should be %d, but got %d", maxNodesInBucket, count)
	}

	// Try to add another node to the existing bucket, and
	// ensure we do not exceed themax nodes in the bucket
	nodePastMax := NewNode()
	nodePastMax.dummyId()
	table1.addToKBucket(nodePastMax)

	if count := table1.nodeCount(); count != maxNodesInBucket {
		t.Errorf("Node count should still be %d, but got %d", maxNodesInBucket, count)
	}
}

func TestAddNodesToKBucketBasic(t *testing.T) {
	table1 := NewDht()
	table1.Node.dummyId()

	node1 := NewNode()
	node1.dummyId()
	node2 := NewNode()
	node2.dummyId()
	node3 := NewNode()
	node3.dummyId()

	table1.addToKBucket(node1)
	table1.addToKBucket(node2)
	table1.addToKBucket(node3)

	if count := table1.nodeCount(); count != 3 {
		t.Errorf("Node count should be %d, but got %d", 3, count)
	}
}

// Test the result of getting the highest allowable index
// for varying different node IDs.
func TestGetHighestAllowableBucketIndex(t *testing.T) {
	table := NewDht()

	thisId := table.Node.Id

	// Test 1: Get index (159 - 159) = 0 for same key
	if result := table.getHighestAllowableBucketIndex(thisId); result != 0 {
		t.Error("Expected to get bucket index 0 for same key, actual index is: ", result)
	}

	// Test 2: Get index (159 - 0 - 1) = 158 for keys with differing MSB
	otherId := make([]byte, keysize)
	result := copy(otherId, thisId)
	if result != keysize {
		t.Error("Expected to copy ", keysize, " bytes, actual result is: ", result)
	}

	// Since the second most significant bit (out of 160 bits) is different,
	// the largest possible difference between the two IDs is 2 ^ 159 - 1. Thus
	// for N = 159, we have 2 ^ N < difference of IDs < 2 ^ (N + 1). Using 0-indexing
	// the resulting value we get is 158.
	for i := 0; i < 8; i++ {
		thisId[0] = 0b00000000
		otherId[0] = 1 << i
		expected := 159 - (7 - i)
		if result := table.getHighestAllowableBucketIndex(otherId); result != expected {
			t.Error("Expected to get bucket index ", expected, " for keys with differing MSB, actual index is: ", result)
		}
	}
}
