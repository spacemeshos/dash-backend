package main

import (
    "flag"
    "net/http"
    "math/rand"
    "time"

    "github.com/spacemeshos/go-spacemesh/log"

    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/collector"
    "github.com/spacemeshos/dash-backend/history"
)

var (
    version string
    commit  string
    branch  string

    nodeAddress = flag.String("node", "localhost:9990", "api node address")
    wsAddr = flag.String("ws", ":8080", "http service address")
    mockNetwork_NetId = flag.Int("mock-network-id", 1, "mock network ID")
    mockNetwork_EpochNumLayers = flag.Int("mock-network-epoch-layers", 288, "mock network Epoch Num Layers")
    mockNetwork_MaxTransactionsPerSecond = flag.Int("mock-network-txs", 10, "mock network Max Transactions Per Second")
    mockNetwork_LayerDuration = flag.Int("mock-network-layer-duration", 15, "mock network Layer Duration")
)

func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    flag.Parse()

    log.InitSpacemeshLoggingSystem("", "spacemesh-dashboard.log")

    bus := client.NewBus()
    go bus.Run()

    history := history.NewHistory(bus)
    go history.Run()

    collector := collector.NewCollector(*nodeAddress, history)
    go collector.Run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        client.ServeWs(bus, w, r)
    })

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        client.ServeWs(bus, w, r)
    })

    err := http.ListenAndServe(*wsAddr, nil)
    if err != nil {
        log.Error("ListenAndServe: ", err)
    }
}

