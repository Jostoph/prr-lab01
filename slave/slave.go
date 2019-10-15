package main

import (
    "../common"
    "bufio"
    "bytes"
    "encoding/binary"
    "fmt"
    "golang.org/x/net/ipv4"
    "log"
    "net"
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
    for {
        n, _, err := conn.ReadFrom(buf)
        if err != nil {
            log.Fatal(err)
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
        var timeSys uint32

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

func onSync(msg []byte) uint32 {
    syncId = binary.LittleEndian.Uint32(msg[1:5])
    return util.GetMilliTimeStamp()
}

func onFollowUp(msg []byte, timeSys uint32) {
    id := binary.LittleEndian.Uint32(msg[5:9])

    if id == syncId {
       timeMaster := binary.LittleEndian.Uint32(msg[1:5])
       timeGap = int64(timeMaster) - int64(timeSys)

       if !step2ready {
           step2ready = true
           go delayCorrection()
       }
    }
}

func onDelayResponse(msg []byte, id uint32) (mTime uint32, valid bool){
    resId := binary.LittleEndian.Uint32(msg[5:9])
    receivedTime := binary.LittleEndian.Uint32(msg[1:5])

    return receivedTime, id == resId
}

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

    //conn, err := net.Dial("udp", config.ServerAddr + ":" + config.ServerPort)
    //if err != nil {
    //   log.Fatal(err)
    //}
    //defer conn.Close()

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

        sendTime := util.GetMilliTimeStamp()

        util.MustCopy(conn, bytes.NewReader(msg))

        fmt.Println("After mustCopy")

        buf := make([]byte, 1024)

        // TODO Change timeout value
        // Wait delay response
        conn.SetReadDeadline(time.Now().Add(5 * time.Second))

        n, err := conn.Read(buf)
        if err != nil {
            fmt.Println("Didn't receive a response on time")
            continue
        }

        s := bufio.NewScanner(bytes.NewReader(buf[0:n]))
        for s.Scan() {
            msg := s.Bytes()

            switch msg[0] {
            case util.DelayResponse :
                mTime, valid := onDelayResponse(msg[:], id)
                if valid {
                    timeDelay := (int64(mTime) - int64(sendTime)) / 2
                    timeSys := util.GetMilliTimeStamp()

                    fmt.Println("received delay resp from serv : delay is : " + strconv.FormatInt(timeDelay, 10))
                    fmt.Println("Total decalage : " + strconv.FormatInt(timeDelay + timeGap, 10))
                    fmt.Println("Local time synced : " + strconv.FormatInt(int64(timeSys) + timeDelay + timeGap, 10))
                }
            }
        }
    }
}