package main

import (
    "flag"
    "net/http"

    "github.com/spacemeshos/go-spacemesh/log"

//    "github.com/spacemeshos/dash-backend/api"
)

var (
    version string
    commit  string
    branch  string

    nodeAddress = flag.String("node", "localhost:9990", "api node address")
    wsAddr = flag.String("ws", ":8080", "http service address")
)

func main() {
    flag.Parse()

    log.InitSpacemeshLoggingSystem("", "spacemesh-dashboard.log")

    bus := newBus()
    go bus.run()

    history := NewHistory(bus)

    m := Message{
        "TESTNET 0.1", 1, 2, 3, 4, 5,
        []Geo{{"1",[2]float64{1,1}}, {"2",[2]float64{2,2}}},
        []Point{{1,1}, {2,1}, {3,1}},
        []Point{{1,2}, {2,2}, {3,2}},
        []Point{{1,3}, {2,3}, {3,3}},
        []Point{{1,4}, {2,4}, {3,4}},
        []Point{{1,5}, {2,5}, {3,5}},
        []Point{{1,6}, {2,6}, {3,6}},
    }
    history.push(&m)

    collector := NewCollector(*nodeAddress, history)
    go collector.Run()

    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        serveWs(bus, w, r)
    })
    err := http.ListenAndServe(*wsAddr, nil)
    if err != nil {
        log.Error("ListenAndServe: ", err)
    }
}
