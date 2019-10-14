package main

import (
    "bufio"
    "bytes"
    "fmt"
    "log"
    "net"
    "os"
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

    // redirect stdin to connection
    go util.MustCopy(conn, os.Stdin)

    //

    go receptionist(config)

    // sync cycle
    var id int
    for {
        id++
        msg := "S," + strconv.Itoa(id)
        sendingTime := time.Now()

        // TODO REMOVE PRINT
        fmt.Println("sync : " + msg)
        r := strings.NewReader(msg)
        // send message
        util.MustCopy(conn, r)

        // TODO change time unit
        // TODO send bytes
        msg = "F," + strconv.FormatInt(sendingTime.UnixNano(), 10) + ","+ strconv.Itoa(id)
        // TODO REMOVE println
        fmt.Println("followup : " + "F," + strconv.FormatInt(sendingTime.UnixNano(), 10) + ","+ strconv.Itoa(id))
        r = strings.NewReader(msg)
        util.MustCopy(conn, r)

        // TODO k unit config ! (nano now)
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
