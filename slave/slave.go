package main

import (
    "bufio"
    "bytes"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
    "math/rand"
    "net"
    "prr-lab01/common"
    "runtime"
    "strconv"
    "strings"
    "time"
)

var address string
var srvAddress string
var srvPort string

var syncDelay int

var syncId int

var step2ready bool

var timeSys int64
var timeGap int64

func main() {
    // load address from config file
    config, err := util.LoadConfiguration("common/config.json")
    if err != nil {
        log.Fatal(err)
    }
    address = config.MulticastAddr + ":" + config.MulticastPort
    srvAddress = config.ServerAddr
    srvPort = config.ServerPort

    // k
    syncDelay = config.SyncDelay

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

            msg := strings.Split(s.Text(), ",")
            switch msg[0] {
            case "S" :
                onSync(msg[0:])
            case "F" :
                onFollowUp(msg[0:])
            }
        }
    }
}

func onSync(msg []string) {
    id, err := strconv.Atoi(msg[1])
    if err != nil {
        log.Fatal(err)
    }

    syncId = id
    timeSys = time.Now().UnixNano()
}

func onFollowUp(msg []string) {
    id, err := strconv.Atoi(msg[2])
    if err != nil {
        log.Fatal(err)
    }

    if id == syncId {
        timeMaster, err := strconv.ParseInt(msg[1], 10, 64)
        if err != nil {
            log.Fatal(err)
        }

        timeGap = timeMaster - timeSys
        fmt.Println(timeGap)

        if !step2ready {
            step2ready = true
            go delayCorrection()
        }
    }
}

func delayCorrection() {
    minDelay := 4 * syncDelay
    maxDelay := 60 * syncDelay

    conn, err := net.Dial("udp", srvAddress + ":" + srvPort)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    for {
        k := rand.Intn(maxDelay - minDelay) + minDelay

        // wait for next iteration
        time.Sleep(time.Duration(k))

        // TODO check setreaddeadline
    }
}