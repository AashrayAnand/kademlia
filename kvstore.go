package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"sync"
	"time"
)

// The actual value for the given key is simply data, but
// to handle emptying the table and replicating the data
// we define Value to wrap the data with the corresponding timeouts
type Value struct {
	Value               []byte
	ExpirationTime      time.Time
	LastTimeReplicated  time.Time
	ReplicationInterval time.Duration
}

type kvstore struct {
	mtx sync.RWMutex
	// The hash table of key/value pairs.
	table map[string]Value
}

// Instantiate key-value store
func NewKVStore() *kvstore {
	k := &kvstore{table: make(map[string]Value), mtx: sync.RWMutex{}}
	return k
}

// Iterate through the key/value pairs in the hash table
// and delete any pairs from the table for which the
// expiration time has passed, should be done periodically
// to avoid congesting table with data
func (k *kvstore) FlushExpiredPairs() {
	for key, value := range k.table {
		if value.ExpirationTime.Before(time.Now()) {
			delete(k.table, key)
		}
	}
}

// Iterate through the key/value pairs in the hash table
// and fetch all keys which have surpassed their replication interval
func (k *kvstore) GetKeysForReplicaion() [][]byte {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	var keys [][]byte
	for key, value := range k.table {
		if value.LastTimeReplicated.Add(value.ReplicationInterval).Before(time.Now()) {
			keys = append(keys, []byte(key))
		}
	}

	return keys
}

// gets sha checksum for a data byte slice, resuts is a 160-bit key
func (k *kvstore) Hash(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

// Get the value for the given key, if found
func (k *kvstore) Get(key []byte) (value string, error error) {
	k.mtx.RLock()
	defer k.mtx.RUnlock()
	if v, ok := k.table[string(key)]; ok {
		return string(v.Value), nil
	}
	return "", errors.New(fmt.Sprintf("Key %s not found", key))
}

// Set the value for the given key. We enforce a timeout on the
// pair to avoid congesting the hash table with too much stale
// data, as well as a replication timer, which enforces how
// often the pair should be replicated to other nodes
func (k *kvstore) Set(key []byte, value []byte, expirationTime time.Time, replicationInterval time.Duration) {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	k.table[string(key)] = Value{Value: value, ExpirationTime: expirationTime, LastTimeReplicated: time.Now(), ReplicationInterval: replicationInterval}
}

// Delete the value for the given key, if found.
func (k *kvstore) Delete(key []byte) error {
	k.mtx.Lock()
	defer k.mtx.Unlock()
	if _, ok := k.table[string(key)]; ok {
		delete(k.table, string(key))
		return nil
	}

	return errors.New(fmt.Sprintf("Key %s not found", key))
}
