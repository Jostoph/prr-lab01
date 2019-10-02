package main

import (
    "bufio"
    "bytes"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
    "net"
    "runtime"
)

const multicastAddr = "224.0.0.1:6666"

func main() {
    clientReader()
}

func clientReader() {
    conn, err := net.ListenPacket("udp", multicastAddr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    p := ipv4.NewPacketConn(conn)
    addr, err := net.ResolveUDPAddr("udp", multicastAddr)
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