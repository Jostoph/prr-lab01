// Package util implements the Configuration struct, protocol constants and useful functions
// used in both master and slave. It also contains function to simulate gaps and delays for testing.
package util

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"os"
)

// Configuration file structure
type Config struct {
	MulticastAddr string `json:"multicast_addr"`
	MulticastPort string `json:"multicast_port"`
	ServerAddr    string `json:"srv_addr"`
	ServerPort    int    `json:"srv_port"`
	SyncDelay     int    `json:"sync_delay"`
}

// Copy src input into des input
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

// Convert a int64 values in an array of bytes
func Int64ToByteArray(array *[]byte, i int64) {
	binary.LittleEndian.PutUint64(*array, uint64(i))
}

// Enum for protocol headers
const (
	Sync          byte = 11
	FollowUp      byte = 12
	DelayRequest  byte = 21
	DelayResponse byte = 22
)
