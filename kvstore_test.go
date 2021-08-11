package main

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"
)

// Test hash set + get
func TestHashSetAndGet(t *testing.T) {
	kvstore := NewKVStore()

	// Create random data
	data := make([]byte, 64)
	rand.Read(data)

	// Hash data to get key
	key := Hash(data)
	kvstore.Set(key, data, time.Now().Add(time.Minute), time.Hour)

	value, err := kvstore.Get(key)
	if err != nil {
		t.Error("Error getting value:", err)
	}

	if !bytes.Equal(value, data) {
		t.Error("Error getting value:", value)
	}
}
