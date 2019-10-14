package main

import (
    "bufio"
    "bytes"
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "prr-lab01/common"
    "strconv"
    "strings"
    "time"
)

var address string

func main() {

    // load address from config file
    config, err := util.LoadConfiguration("common/config.json")
    if err != nil {
        log.Fatal(err)
    }
    address = config.MulticastAddr + ":" + config.MulticastPort

    // open connection
    conn, err := net.Dial("udp", address)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // start receptionist
    go receptionist(config)

    // sync cycle
    var id uint32
    for {
        id++

        idBytes := make([]byte, 4)
        binary.LittleEndian.PutUint32(idBytes, id)

        // add header SYNC
        msg := make([]byte, 1)
        msg[0] = util.Sync // = append(msg, util.Sync)

        // add id
        msg = append(msg, idBytes...)
        //r := strings.NewReader(msg)

        // send message
        r := bytes.NewReader(msg)
        util.MustCopy(conn, r)
        // TODO remove
        fmt.Println("before sync")

        // add header FOLLOW_UP
        msg = make([]byte, 1)
        msg[0] = util.FollowUp//append(msg, util.FollowUp)

        // add time
        timeBytes := make([]byte, 4)
        util.UintToBytes(&timeBytes, util.GetMilliTimeStamp())
        msg = append(msg, timeBytes...)

        // add id
        msg = append(msg, idBytes...)

        // send message
        r = bytes.NewReader(msg)
        util.MustCopy(conn, r)

        fmt.Println("after sync")
        // TODO k unit config ! (nano now)
        // sleep until next cycle
        time.Sleep(time.Duration(config.SyncDelay))
    }
}

func receptionist(config util.Config) {
    conn, err := net.ListenPacket("udp", config.ServerAddr + ":" + config.ServerPort)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    buf := make([]byte, 1024)

    for {
        n, cliAddr, err := conn.ReadFrom(buf)
        if err != nil {
            log.Fatal(err)
        }

        go worker(&conn, cliAddr, buf, n, time.Now().UnixNano())
    }
}

func worker(conn *net.PacketConn, cliAddr net.Addr, buf []byte, n int, receiveTime int64) {
    s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
    for s.Scan() {

        msg := strings.Split(s.Text(), ",")
        switch msg[0] {
        case "R" :
            res := "Q," + strconv.FormatInt(receiveTime, 10) + "," + msg[1]

            // TODO remove debug
            fmt.Println("send delay response to : " + cliAddr.String())

            if _, err := (*conn).WriteTo([]byte(res), cliAddr); err != nil {
                log.Fatal(err)
            }
        }
    }
}
