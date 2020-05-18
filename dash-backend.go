package main

import (
    "flag"
    "net/http"
    "math/rand"
    "time"

    "github.com/spacemeshos/go-spacemesh/log"

//    "github.com/spacemeshos/dash-backend/api"
)

var (
    version string
    commit  string
    branch  string

    nodeAddress = flag.String("node", "localhost:9990", "api node address")
    wsAddr = flag.String("ws", ":8080", "http service address")
    mock = flag.Bool("mock", false, "mock mode")
)

func main() {
    rand.Seed(time.Now().UTC().UnixNano())

    flag.Parse()

    log.InitSpacemeshLoggingSystem("", "spacemesh-dashboard.log")

    bus := newBus()
    go bus.run()

    history := NewHistory(bus)
    if *mock {
        go history.RunMock()
    } else {
        go history.Run()
    }

    if !*mock {
        collector := NewCollector(*nodeAddress, history)
        go collector.Run()
    }

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(bus, w, r)
    })
    err := http.ListenAndServe(*wsAddr, nil)
    if err != nil {
        log.Error("ListenAndServe: ", err)
    }
}
