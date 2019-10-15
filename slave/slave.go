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

		fmt.Println(buf[0])

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

		fmt.Println(strconv.FormatInt(int64(id), 10) + ") time master is : " +
			strconv.FormatInt(timeMaster, 10))

		fmt.Println(strconv.FormatInt(int64(id), 10) + ") time in slave is : " +
			strconv.FormatInt(timeSys, 10))

		fmt.Println(strconv.FormatInt(int64(id), 10) + ") time gap is :" + strconv.FormatInt(timeGap, 10))

		// If the first phase (sync-follow_up) is successful we can start the second phase (delay computation)
		if !step2ready {
			step2ready = true

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
	//minDelay := 4 * syncDelay
	//maxDelay := 10 * syncDelay
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
		// TODO change
		// k := rand.Intn(maxDelay - minDelay) + minDelay
		k := 6 * config.SyncDelay

		// Wait for next iteration
		time.Sleep(time.Duration(k) * time.Second)

		id++

		// Prepare DelayRequest
		msg := make([]byte, 2)
		msg[0] = util.DelayRequest // Header
		msg[1] = id                // Id

		// Send time (current system time)
		sendTime := time.Now().UnixNano()

		// Send request
		util.MustCopy(conn, bytes.NewReader(msg))

		fmt.Println("After mustCopy")

		buf := make([]byte, 1024)

		// Wait delay response
		err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Fatal(err)
		}

		// Read response from master server
		_, _, err = conn.ReadFrom(buf)
		if err != nil {
			fmt.Println("Didn't receive a response on time")
			continue
		}

		if buf[0] == util.DelayResponse {
			// Master system time of request reception
			mTime, valid := onDelayResponse(buf[:], id)

			if valid {
				// Computing delay (ping) between master and slave
				timeDelay := (mTime - sendTime) / 2
				timeSys := time.Now().UnixNano()

				fmt.Println("received delay resp from serv : delay is : " + strconv.FormatInt(timeDelay, 10))
				fmt.Println("Total decalage : " + strconv.FormatInt(timeDelay+timeGap, 10))
				fmt.Println("Local time synced : " + strconv.FormatInt(timeSys+timeDelay+timeGap, 10))
			}
		}
	}
}
