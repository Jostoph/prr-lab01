package main

import (
    "bufio"
    "bytes"
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "prr-lab01/common"
    "time"
)

var config util.Config

func main() {
    // load address from config file
    config := util.LoadConfiguration("common/config.json")

    address := config.MulticastAddr + ":" + config.MulticastPort

    // open connection
    conn, err := net.Dial("udp", address)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // start receptionist
    go receptionist()

    // sync cycle
    var id uint32
    for {
        id++
        idBytes := make([]byte, 4)
        binary.LittleEndian.PutUint32(idBytes, id)

        // Prepare Sync request
        msg := make([]byte, 1)
        msg[0] = util.Sync              // Header
        msg = append(msg, idBytes...)   // Id

        // Send sync request
        util.MustCopy(conn, bytes.NewReader(msg))

        // Convert time in byte array
        timeBytes := make([]byte, 4)
        util.UintToBytes(&timeBytes, util.GetMilliTimeStamp())

        // Prepare FollowUp request
        msg = make([]byte, 1)
        msg[0] = util.FollowUp          // Header
        msg = append(msg, timeBytes...) // Time
        msg = append(msg, idBytes...)   // Id

        // Send FollowUp request
        util.MustCopy(conn, bytes.NewReader(msg))

        fmt.Println(util.GetMilliTimeStamp()) // TODO remove

        // TODO k unit config ! (nano now)
        // sleep until next cycle
        time.Sleep(time.Duration(config.SyncDelay))
    }
}

func receptionist() {
    // conn, err := net.ListenPacket("udp", config.ServerAddr + ":" + config.ServerPort)
    conn, err := net.ListenUDP("udp", &net.UDPAddr{IP:[]byte{0,0,0,0},Port:8173,Zone:""})
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    buf := make([]byte, 1024)

    for {
        fmt.Println("Before ReadFROM") // TODO remove
        n, cliAddr, err := conn.ReadFrom(buf)
        fmt.Println("After ReadFROM") // TODO remove
        if err != nil {
            log.Fatal(err)
        }

        // Handle communication with a slave
        go worker(conn, cliAddr, buf, n, util.GetMilliTimeStamp())
    }
}

func worker(conn *net.UDPConn, cliAddr net.Addr, buf []byte, n int, receiveTime uint32) {
    s := bufio.NewScanner(bytes.NewReader(buf[0:n]))

    for s.Scan() {
        msg := s.Bytes()

        switch msg[0] {
        case util.DelayRequest :
            // Convert time in byte array
            timeBytes := make([]byte, 4)
            util.UintToBytes(&timeBytes, receiveTime)

            // Prepare DelayResponse request
            res := make([]byte, 1)
            res[0] = util.DelayResponse     // Header
            res = append(res, timeBytes...) // Time
            res = append(res, msg[1:5]...)  // Id

            // Send DelayResponse request
            if _, err := (*conn).WriteTo(res, cliAddr); err != nil {
                log.Fatal(err)
            }
        }
    }
}
