package util

import (
    "encoding/json"
    "io"
    "log"
    "os"
)

type Config struct {
    MulticastAddr string `json:"multicast_addr"`
    Port string `json:"port"`
}

func MustCopy(dst io.Writer, src io.Reader) {
    if _, err := io.Copy(dst, src); err != nil {
        log.Fatal(err)
    }
}

func LoadConfiguration(filename string) (Config, error) {
    var config Config
    configFile, err := os.Open(filename)
    defer configFile.Close()

    if err != nil {
        return config, err
    }

    jsonParser := json.NewDecoder(configFile)
    err = jsonParser.Decode(&config)
    return config, err
}
