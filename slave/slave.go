package main

import (
    "bufio"
    "bytes"
    "encoding/binary"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
    "net"
    "prr-lab01/common"
    "runtime"
    "strconv"
    "time"
)

var config util.Config

var syncId uint32
var step2ready bool
var timeGap int64

func main() {
    // Load address from config file
    config = util.LoadConfiguration("common/config.json")

    // Read incoming messages
    clientReader()
}

func clientReader() {
    multicastAddress := config.MulticastAddr + ":" + config.MulticastPort

    conn, err := net.ListenPacket("udp", multicastAddress)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    p := ipv4.NewPacketConn(conn)
    addr, err := net.ResolveUDPAddr("udp", multicastAddress)
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
    var timeSys int64
    for {
        n, _, err := conn.ReadFrom(buf)
        if err != nil {
            log.Fatal(err)
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))

        for s.Scan() {
            msg := s.Bytes()
            fmt.Println(buf[0])

            switch msg[0] {
            case util.Sync :
                timeSys = onSync(msg[:])
            case util.FollowUp :
                onFollowUp(msg[:], timeSys)
            }
        }
    }
}

// Handle Sync request
func onSync(msg []byte) int64 {
    syncId = binary.LittleEndian.Uint32(msg[1:5])
    return time.Now().UnixNano()
}

// Handle FollowUp request
func onFollowUp(msg []byte, timeSys int64) {
    id := binary.LittleEndian.Uint32(msg[9:13])

    if id == syncId {
       timeMaster := int64(binary.LittleEndian.Uint64(msg[1:9]))
       timeGap = timeMaster - timeSys

       fmt.Println(strconv.FormatInt(int64(id), 10) + ") time master is : " +
                   strconv.FormatInt(timeMaster, 10))

       fmt.Println(strconv.FormatInt(int64(id), 10) + ") time in slave is : " +
                   strconv.FormatInt(timeSys, 10))

       fmt.Println(strconv.FormatInt(int64(id), 10) + ") time gap is :" + strconv.FormatInt(timeGap, 10))

       if !step2ready {
           step2ready = true
           go delayCorrection()
       }
    }
}

// Handle DelayResponse resquest
func onDelayResponse(msg []byte, id uint32) (int64, bool){
    resId := binary.LittleEndian.Uint32(msg[9:13])
    receivedTime := binary.LittleEndian.Uint64(msg[1:9])

    return int64(receivedTime), id == resId
}

// Check the delay and correct it with the help of a DelayRequest
func delayCorrection() {
    //minDelay := 4 * syncDelay
    //maxDelay := 10 * syncDelay
    var id uint32 = 0

    addr, err := net.ResolveUDPAddr("udp", config.ServerAddr + ":" + config.ServerPort)
    if err != nil {
      log.Fatal(err)
    }

    conn, err := net.DialUDP("udp", nil, addr)
    if err != nil {
      log.Fatal(err)
    }

    defer conn.Close()

    for {
        // TODO change
        // k := rand.Intn(maxDelay - minDelay) + minDelay
        k := 15 * config.SyncDelay

        // Wait for next iteration
        time.Sleep(time.Duration(k))

        id++
        idBytes := make([]byte, 4)
        binary.LittleEndian.PutUint32(idBytes, id)

        // Prepare DelayRequest
        msg := make([]byte, 1)
        msg[0] = util.DelayRequest      // Header
        msg = append(msg, idBytes...)   // Id

        sendTime := time.Now().UnixNano()

        util.MustCopy(conn, bytes.NewReader(msg))

        fmt.Println("After mustCopy")

        buf := make([]byte, 1024)

        // TODO Change timeout value
        // Wait delay response
        err := conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        if err != nil {
            fmt.Println("Didn't receive a response on time")
            continue
        }

        n, _, err := conn.ReadFrom(buf)
        if err != nil {
            log.Fatal(err)
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
        for s.Scan() {
            msg := s.Bytes()

            switch msg[0] {
            case util.DelayResponse :
                mTime, valid := onDelayResponse(msg[:], id)
                if valid {
                    timeDelay := (mTime - sendTime) / 2
                    timeSys := time.Now().UnixNano()

                    fmt.Println("received delay resp from serv : delay is : " + strconv.FormatInt(timeDelay, 10))
                    fmt.Println("Total decalage : " + strconv.FormatInt(timeDelay + timeGap, 10))
                    fmt.Println("Local time synced : " + strconv.FormatInt(timeSys + timeDelay + timeGap, 10))
                }
            }
        }
    }
}