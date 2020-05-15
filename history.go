package main

import (
    "encoding/json"
//    "github.com/spacemeshos/go-spacemesh/log"
    "github.com/spacemeshos/go-spacemesh/common/types"
)

type Point struct {
    Uv	uint64
    Amt	uint64
}

type Geo struct {
    Name	string
    Coordinates	[2]float64
}

type Message struct {
    Network	string
    Age		uint64
    Layer	uint64
    Epoch	uint64
    Capacity	uint64
    Decentral	uint64
    SmeshersGeo		[]Geo
    Smeshers		[]Point
    Transactions	[]Point
    Accounts		[]Point
    Circulation		[]Point
    Rewards		[]Point
    Security		[]Point
}

type State struct {
    layer types.Layer
}

type History struct {
    bus *Bus

    current *State

    layers map[types.LayerID]State
}

func (m *Message) toJson() []byte {
    b, _ := json.Marshal(m)
    return b
}

func (h *History) push(m *Message) {
    h.bus.notify <- m.toJson()
}

func NewHistory(bus *Bus) *History {
    return &History{bus: bus, current: nil, layers: make(map[types.LayerID]State)}
}

func (c *History) Run() {
}
