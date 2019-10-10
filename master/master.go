package main

import (
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

    // sync cycle
    var id int
    for {
        id++
        msg := "S," + strconv.Itoa(id)
        sendingTime := time.Now()

        r := strings.NewReader(msg)
        // send message
        util.MustCopy(conn, r)

        // TODO change time unit
        // TODO send bytes
        msg = "F," + strconv.FormatInt(sendingTime.UnixNano(), 10) + ","+ strconv.Itoa(id)
        r = strings.NewReader(msg)
        util.MustCopy(conn, r)

        // TODO k unit config ! (nano now)
        time.Sleep(time.Duration(config.SyncDelay))
    }
}
