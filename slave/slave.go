package main

import (
    "bufio"
    "bytes"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
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
var timeDelay int64

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
    // TODO REMOVE PRINT
    fmt.Println("sync from serv : " + strconv.Itoa(id))
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
        fmt.Println("follow up from serv with gap : " + strconv.FormatInt(timeGap, 10))

        if !step2ready {
            step2ready = true
            // TODO REMOVE PRINT
            fmt.Println("step 2 ready")
            go delayCorrection()
        }
    }
}

func onDelayResponse(msg []string, id int) (mTime int64, valid bool){
    resId, err := strconv.Atoi(msg[2])
    if err != nil {
        log.Fatal(err)
    }
    var receivedTime int64
    receivedTime, err = strconv.ParseInt(msg[1], 10, 64)
    if err != nil {
        log.Fatal(err)
    }

    return receivedTime, id == resId
}

func delayCorrection() {
    //minDelay := 4 * syncDelay
    //maxDelay := 10 * syncDelay
    id := 0

    conn, err := net.Dial("udp", srvAddress + ":" + srvPort)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    for {
        // k := rand.Intn(maxDelay - minDelay) + minDelay
        k := 15 * syncDelay

        // wait for next iteration
        time.Sleep(time.Duration(k))

        // TODO check setreaddeadline

        // send delay request
        id++
        sendTime := time.Now().UnixNano()
        _ , err := fmt.Fprintf(conn, "R," + strconv.Itoa(id))
        if err != nil {
            log.Fatal(err)
        }

        // TODO REMOVE PRINT
        fmt.Println("send delay request")

        // wait delay response

        buf := make([]byte, 1024)

        // TODO set goo deadline time (5 sec now)
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))

        n, err := conn.Read(buf)
        if err != nil {
            log.Fatal(err)
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
        for s.Scan() {
            // Q -> response delay
            msg := strings.Split(s.Text(), ",")
            switch msg[0] {
            case "Q" :
                // TODO REMOVE PRINT
                mTime, valid := onDelayResponse(msg, id)
                if valid {
                    timeDelay = (mTime - sendTime) / 2
                    fmt.Println("received delay resp from serv : delay is : " + strconv.FormatInt(timeDelay, 10))
                }
            }
            fmt.Println("Total decalage : " + strconv.FormatInt(timeDelay + timeGap, 10))
        }
    }
}