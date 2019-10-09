package main

import (
    "bufio"
    "bytes"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
    "net"
    util "prr-lab01/common"
    "runtime"
)

var address string

func main() {
    // load address from config file
    config, err := util.LoadConfiguration("common/config.json")
    if err != nil {
        log.Fatal(err)
    }
    address = config.MulticastAddr + ":" + config.Port

    // read incoming messages
    clientReader()
}

func clientReader() {
    conn, err := net.ListenPacket("udp", address)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    p := ipv4.NewPacketConn(conn)
    addr, err := net.ResolveUDPAddr("udp", address)
    if err != nil {
        log.Fatal(err)
    }

    var interf *net.Interface
    if runtime.GOOS == "darwin" {
        interf, _ = net.InterfaceByName("en0")
    }

    if err = p.JoinGroup(interf, addr); err != nil {
        log.Fatal(err)
    }

    buf := make([]byte, 1024)
    for {
        n, addr, err := conn.ReadFrom(buf)
        if err != nil {
            log.Fatal(err)
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
        for s.Scan() {
            fmt.Printf("%s from %v\n", s.Text(), addr)
        }
    }
}