// Package main (master) implements a simple server to synchronize multiple Slave clients
// using an implementation of the Precise Time Protocol (PTP). The server uses multicast
// networking and UDP point-to-point connections to communicate with the Slaves.
// Times are defined by Unix nanoseconds timestamps (int64).
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"prr-lab01/common"
	"strconv"
	"time"
)

var config util.Config

// Main function of the master's programme. It starts the synchronization broadcasting and launches
// the receptionist routine that will serve the connecting Slaves.
func main() {
	// Load configuration file
	config = util.LoadConfiguration("common/config.json")

	// Info log
	fmt.Println("\nStarting Master Server...")
	fmt.Println("\nSimulated Gap : " + strconv.Itoa(config.SimulationGap) + "ms")
	fmt.Println("Simulated Master->Slave Delay : " + strconv.Itoa(config.SimulationDelay) + "ms\n")

	address := config.MulticastAddr + ":" + config.MulticastPort

	// Info log
	fmt.Println("Broadcasting on : " + address + "\n")

	// Open broadcasting connection
	conn, err := net.Dial("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Start receptionist
	go receptionist()

	// Verification id
	var id byte

	// For loop that sends the two synchronization message used int the master-slave time gap correction
	// every K seconds (SyncDelay, in the configuration file).
	for {
		id++

		// Prepare Sync request
		msg := make([]byte, 2)
		msg[0] = util.Sync // Header
		msg[1] = id        // Id

		// Send sync request
		util.MustCopy(conn, bytes.NewReader(msg))

		// Convert time in byte array
		timeBytes := make([]byte, 8)

		// Artificial gap that can be added to the timestamp (milliseconds)
		simGap := int64(config.SimulationGap * int(time.Millisecond))

		util.Int64ToByteArray(&timeBytes, time.Now().UnixNano()+simGap)

		// Prepare FollowUp request
		msg = make([]byte, 1)
		msg[0] = util.FollowUp          // Header
		msg = append(msg, timeBytes...) // Time
		msg = append(msg, id)           // Id

		// Send FollowUp request
		util.MustCopy(conn, bytes.NewReader(msg))

		// Sleep until next cycle
		time.Sleep(time.Duration(config.SyncDelay) * time.Second)
	}
}

// Point-to-point UDP server that receives arriving clients and delegates them to worker routines.
func receptionist() {

	// Resolve server address
	addr, err := net.ResolveUDPAddr("udp", config.ServerAddr+":"+strconv.Itoa(config.ServerPort))
	if err != nil {
		log.Fatal(err)
	}

	// Open connection
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		// Read incoming Slave packets
		_, cliAddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}

		// Delegate Slave handling to a worker routine
		go worker(conn, cliAddr, buf)
	}
}

// Handle Slave message and send according response
func worker(conn *net.UDPConn, cliAddr net.Addr, buf []byte) {

	// Info log
	fmt.Println("Delay request from Slave : " + cliAddr.String())

	// Add artificial reception delay
	util.SimulateDelay(config.SimulationDelay)

	// Get timestamp of reception
	receiveTime := time.Now().UnixNano()

	if buf[0] == util.DelayRequest {

		// Convert time in byte array
		timeBytes := make([]byte, 8)
		util.Int64ToByteArray(&timeBytes, receiveTime)

		// Prepare DelayResponse request
		res := make([]byte, 1)
		res[0] = util.DelayResponse     // Header
		res = append(res, timeBytes...) // Time
		res = append(res, buf[1])       // Id

		// Send DelayResponse request
		if _, err := (conn).WriteTo(res, cliAddr); err != nil {
			log.Fatal(err)
		}
	}
}
