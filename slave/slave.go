// Package main (slave) implements a simple client to be synchronized by a master server using
// broadcasting and point-to-point UDP connections following the Precise Time Protocol (PTP).
// Times are defined by Unix nanoseconds timestamps (int64).
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"golang.org/x/net/ipv4"
	"log"
	"math/rand"
	"net"
	"prr-lab01/common"
	"runtime"
	"strconv"
	"time"
)

var config util.Config

var syncId byte     // Current sync id
var step2ready bool // Status of the second PTP phase (set to true when ready)
var timeGap int64   // time gap between the master and the slave's time

// Main function of the Slave programme. It connects to a broadcast channel and waits for
// the synchronization messages from the master.
func main() {
	// Load configuration file
	config = util.LoadConfiguration("common/config.json")

	// Info log
	fmt.Println("\nStarting Slave...")
	fmt.Println("Joining group : " + config.MulticastAddr + ":" + config.MulticastPort + "\n")

	multicastAddress := config.MulticastAddr + ":" + config.MulticastPort

	// Open connection
	conn, err := net.ListenPacket("udp", multicastAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	p := ipv4.NewPacketConn(conn)
	addr, err := net.ResolveUDPAddr("udp", multicastAddress)
	if err != nil {
		log.Fatal(err)
	}

	// Handle network interface issues with darwin GOOS
	var interf *net.Interface
	if runtime.GOOS == "darwin" {
		interf, _ = net.InterfaceByName("en0")
	}

	// Join broadcast group
	if err = p.JoinGroup(interf, addr); err != nil {
		log.Fatal(err)
	}

	buf := make([]byte, 1024)
	var timeSys int64
	for {
		// Read master packets
		_, _, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}

		switch buf[0] {
		case util.Sync:
			timeSys = onSync(buf[:])
		case util.FollowUp:
			onFollowUp(buf[:], timeSys)
		}
	}
}

// Handle Sync request returning the current system time.
func onSync(msg []byte) int64 {
	syncId = msg[1]
	return time.Now().UnixNano()
}

// Handle FollowUp requests and start the second phase of the protocol if the operation was
// successful.
func onFollowUp(msg []byte, timeSys int64) {
	id := msg[9]

	// If the id of the last SYNC message matches the FOLLOW_UP id we can proceed
	if id == syncId {
		timeMaster := int64(binary.LittleEndian.Uint64(msg[1:9]))

		// Computing time gap between master and slave
		timeGap = timeMaster - timeSys

		// Info log
		fmt.Println("Time Gap Master-Slave : " + time.Duration(timeGap).String())

		// If the first phase (sync-follow_up) is successful we can start the second phase (delay computation)
		if !step2ready {
			step2ready = true

			// Info log
			fmt.Println("Starting second phase, computing delay correction...")

			// Start delay correction routine
			go delayCorrection()
		}
	}
}

// Handle DelayResponse requests and returns the time at which the master
// received the delay request with a boolean set to true if the id of the
// delay request matches the response id.
func onDelayResponse(msg []byte, id byte) (int64, bool) {
	resId := msg[9]
	receivedTime := binary.LittleEndian.Uint64(msg[1:9])

	return int64(receivedTime), id == resId
}

// Check the delay and correct it with the help of a DelayRequest
func delayCorrection() {
	minDelay := 4 * config.SyncDelay
	maxDelay := 10 * config.SyncDelay
	var id byte

	port := strconv.Itoa(config.ServerPort)
	addr, err := net.ResolveUDPAddr("udp", config.ServerAddr+":"+port)
	if err != nil {
		log.Fatal(err)
	}

	// Open connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// For loop that executes the second phase of the PTP protocol. It sends and receives delay requests
	// and responses from the master in a random-timed cycle defined by K (SyncDelay, in the configuration file)
	for {
		// Computing random waiting time
		k := rand.Intn(maxDelay-minDelay) + minDelay

		// Wait for next iteration
		time.Sleep(time.Duration(k) * time.Second)

		id++

		// Prepare DelayRequest
		msg := make([]byte, 2)
		msg[0] = util.DelayRequest // Header
		msg[1] = id                // Id

		// Send time (current system time)
		sendTime := time.Now().UnixNano()

		// Info log
		fmt.Println("\nSending Delay Request with id : " + strconv.Itoa(int(id)))

		// Send request
		util.MustCopy(conn, bytes.NewReader(msg))

		buf := make([]byte, 1024)

		// Wait delay response with time out of 5 + syncDelay seconds
		err := conn.SetReadDeadline(time.Now().Add(time.Duration(5000 + config.SimulationDelay) * time.Millisecond))
		if err != nil {
			log.Fatal(err)
		}

		// Read response from master server
		_, _, err = conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Timeout : Didn't receive a response from the server. Retrying....")
			continue
		}

		if buf[0] == util.DelayResponse {
			// Master system time of request reception
			mTime, valid := onDelayResponse(buf[:], id)

			if valid {
				// Computing delay between master and slave (one way)
				timeDelay := (mTime - (sendTime + timeGap)) / 2

				// Info log
				fmt.Println("Received Delay Response with id : " + strconv.Itoa(int(id)))
				fmt.Println("Delay Master->Slave is : " + time.Duration(timeDelay).String())
				fmt.Println("Total time difference Master-Slave (delay + gap) : " +
					time.Duration(timeDelay+timeGap).String())
			}
		}
	}
}
