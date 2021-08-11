package main

import (
	"fmt"
	"net"
	"testing"
)

// For this basic test, we should spin up a single dht, and
// confirm that we can then connect to this dht from a client
// and close the dht listener.
func TestConnectToDhtBasic(t *testing.T) {
	dht := NewDht()

	go func() {
		defer func() { dht.Done <- struct{}{} }()

		dht.entry()
	}()

	client, err := net.Dial("tcp", dht.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Error connecting to dht at %s, error: %s", dht.Listener.Addr().String(), err)
	}

	client.Close()
	<-dht.ConnClosed
	dht.Listener.Close()
	<-dht.Done
}

// For this basic test, we should spin up a single dht, and
// confirm that we can then connect to this dht from a client
// and close the dht listener.
func TestPing(t *testing.T) {
	dht := NewDht()
	dht2 := NewDht()

	go func() {
		defer func() { dht.Done <- struct{}{} }()
		dht.entry()
	}()

	go func() {
		defer func() { dht2.Done <- struct{}{} }()
		dht2.entry()
	}()

	fmt.Println(dht.Listener.Addr().String())
	fmt.Println(dht2.Listener.Addr().String())

	_, err := net.Dial("tcp", dht.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Error connecting to dht at %s, error: %s", dht.Listener.Addr().String(), err)
	}

	_, err = net.Dial("tcp", dht2.Listener.Addr().String())
	if err != nil {
		t.Fatalf("Error connecting to dht at %s, error: %s", dht.Listener.Addr().String(), err)
	}

	// Sending ping should initiate 2 connections, each should
	// send a conn closed signal to the channel of the respective
	// dht, we wait on both of these signals to proceed
	dht2.sendMessage(dht2.formPingMsg(false), dht.Listener.Addr().String())
	<-dht.ConnClosed
	<-dht2.ConnClosed

	// After ping, we should add dht2 node to the k buckets for dht
	if res := dht.nodeCount(); res != 1 {
		t.Fatalf("Node count for dht should be 1, was %d", res)
	}

	// Likewise after pong, we add dht2 node to the k buckets for dht
	if res := dht2.nodeCount(); res != 1 {
		t.Fatalf("Node count for dht2 should be 1, was %d", res)
	}

	dht.Listener.Close()
	<-dht.Done
	dht2.Listener.Close()
	<-dht2.Done
}
