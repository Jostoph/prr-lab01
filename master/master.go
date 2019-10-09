package main

import (
    "log"
    "net"
    "os"
    "prr-lab01/common"
)

var address string

func main() {

    // load address from config file
    config, err := util.LoadConfiguration("common/config.json")
    if err != nil {
        log.Fatal(err)
    }
    address = config.MulticastAddr + ":" + config.Port

    // open connection
    conn, err := net.Dial("udp", address)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    util.MustCopy(conn, os.Stdin)
}
