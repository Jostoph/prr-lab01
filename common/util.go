package util

import (
    "encoding/binary"
    "encoding/json"
    "io"
    "log"
    "os"
)

// Get the configuration from a JSON file
type Config struct {
    MulticastAddr string `json:"multicast_addr"`
    MulticastPort string `json:"multicast_port"`
    ServerAddr string `json:"srv_addr"`
    ServerPort string `json:"srv_port"`
    SyncDelay int `json:"sync_delay"`
}

func MustCopy(dst io.Writer, src io.Reader) {
    if _, err := io.Copy(dst, src); err != nil {
        log.Fatal(err)
    }
}

// Load the configuration file
func LoadConfiguration(filename string) Config {
    var config Config
    configFile, err := os.Open(filename)
    defer configFile.Close()

    if err != nil {
        log.Fatal(err)
    }

    jsonParser := json.NewDecoder(configFile)
    err = jsonParser.Decode(&config)
    return config
}

// Convert a uint32 value in an array of byte
func Int64ToByteArray(array *[]byte, i int64) {
    binary.LittleEndian.PutUint64(*array, uint64(i))
}

// Enum for protocol header
const (
    Sync byte = 11
    FollowUp byte = 12
    DelayRequest byte = 21
    DelayResponse byte = 22
)
