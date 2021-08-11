package main

import (
	"crypto/sha1"
	"log"
	"math/rand"
	"time"
)

// There will be one instance of this struct within the package, and
// configurations supplied by the user are stored in this structure
// and references in utilities
type Config struct {
	LoggingEnabled bool
}

var configs Config

const (
	alpha            = 3                                           // the degrees of parallelism in network requests
	numBuckets       = 160                                         // the number of k-buckets in the DHT
	maxNodesInBucket = 20                                          // the maximum keys in a single bucket
	keysize          = 20                                          // size in bytes of the keys used to ID nodes and values
	tExpire          = time.Duration(24 * time.Hour)               // time after which key value pair expires
	tRefresh         = time.Duration(time.Hour)                    // time after which bucket is refreshed
	tReplicate       = time.Duration(time.Hour)                    // time after which key value pair is replicated
	tRepublish       = time.Duration(24*time.Hour + 1*time.Minute) // time after which original publisher re-publishes key
)

// gets sha checksum for a data byte slice, resuts is a 160-bit key
func Hash(data []byte) []byte {
	hash := sha1.Sum(data)
	return hash[:]
}

// generate 64-bit random message ID
func GenerateMsgId() []byte {
	msgId := make([]byte, 16)
	_, err := rand.Read(msgId)
	if err != nil {
		log.Fatal(err)
	}

	return msgId
}

// Helper to write to log if logging flag is enabled
func writeLog(msg string, args ...interface{}) {
	if configs.LoggingEnabled {
		log.Printf(msg, args...)
	}
}
