package main

import (
    "log"
    "net"
    "os"
    "prr-lab01/common"
)

const multicastAddr = "224.0.0.1:6666"

func main() {
    conn, err := net.Dial("udp", multicastAddr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    util.MustCopy(conn, os.Stdin)
}
