package main

import (
    "flag"
    "net/http"
    "math/rand"
    "time"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/history"
)

var (
    version string
    commit  string
    branch  string

    nodeAddress = flag.String("node", "localhost:9092", "api node address")
    wsAddr = flag.String("ws", ":8080", "http service address")
    mongoDbUrlStringFlag = flag.String("mongodb", "mongodb://localhost:27017", "Explorer MongoDB Uri string in format mongodb://<host>:<port>")
    mongoDbNameStringFlag = flag.String("db", "explorer", "MongoDB Explorer database name string")
)

func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    flag.Parse()

    log.InitSpacemeshLoggingSystem("", "spacemesh-dashboard.log")

    bus := client.NewBus()
    go bus.Run()

    history, err := history.NewHistory(nil, bus, *mongoDbUrlStringFlag, *mongoDbNameStringFlag)
    if err != nil {
        log.Error("Create History service error: %v", err)
        return
    }
    go history.Run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        client.ServeWs(bus, w, r)
    })

    err = http.ListenAndServe(*wsAddr, nil)
    if err != nil {
        log.Error("Create HTTP ssrver error: %v", err)
        return
    }
}

