package main

import (
	"bytes"
	"log"
	"net"
	"prr-lab01/common"
	"strconv"
	"time"
)

var config util.Config

func main() {
	// load address from config file
	config = util.LoadConfiguration("common/config.json")

	address := config.MulticastAddr + ":" + config.MulticastPort

	// Open connection
	conn, err := net.Dial("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// start receptionist
	go receptionist()

	// sync cycle
	var id byte
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
		util.Int64ToByteArray(&timeBytes, time.Now().UnixNano())

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

func receptionist() {

	addr, err := net.ResolveUDPAddr("udp", config.ServerAddr+":"+strconv.Itoa(config.ServerPort))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)

	for {
		// Wait a communication from a slave
		n, cliAddr, err := conn.ReadFrom(buf)
		if err != nil {
			log.Fatal(err)
		}

		// Handle communication with a slave
		go worker(conn, cliAddr, buf, n, time.Now().UnixNano())
	}
}

func worker(conn *net.UDPConn, cliAddr net.Addr, buf []byte, n int, receiveTime int64) {

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
