package main

import (
    "math/rand"
    "encoding/json"
    "sync"
    "time"
//    "github.com/spacemeshos/go-spacemesh/log"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
//    pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
)

type State struct {
    layer *Layer

    capacity		uint64
    decentral		uint64
    smeshers		uint64
    transactions	uint64
    accounts		uint64
    circulation		uint64
    rewards		uint64
    security		uint64
}

type History struct {
    bus 	*Bus

    smeshersGeo	[]Geo
    genesis	time.Time
    epoch	uint64
    capacity	uint64
    decentral	uint64

    current 	*State

    layers 	map[sm.LayerID]*State
    mux 	sync.Mutex
}

func newState(layer *Layer) *State {
    state := &State{}
    state.update(layer)
    return state
}

func (s *State) update(layer *Layer) {
    s.layer = layer
    s.transactions = 0
    for _, block := range layer.blocks {
        s.transactions += uint64(len(block.transactions))
    }
}

func (h *History) onLayer(layer *Layer) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[layer.index]
    if ok {
        state.update(layer)
    } else {
        state = newState(layer)
        h.layers[layer.index] = state
    }
    if h.current != nil {
        if h.current.layer.index < state.layer.index  {
            h.current = state
        }
    } else {
        h.current = state
    }
    message := h.createMessage()
    h.push(message)
}

func (h *History) createMessage() *Message {
    var i int
    var index sm.LayerID = 1
    statesCount := len(h.layers)
    message := &Message{}
    message.Network = "TESTNET 0.1"
    message.Age = uint64(time.Now().Second() - h.genesis.Second())
    message.Epoch = h.epoch
    if h.current != nil {
        message.Layer = uint64(h.current.layer.index)
        message.Capacity = h.current.capacity
        message.Decentral = h.current.decentral
    }
//    message.SmeshersGeo		[]Geo
    for i = 0; i < cPointsCount; i++ {
        message.Smeshers[i].Uv     = uint64(i)
        message.Transactions[i].Uv = uint64(i)
        message.Accounts[i].Uv     = uint64(i)
        message.Circulation[i].Uv  = uint64(i)
        message.Rewards[i].Uv      = uint64(i)
        message.Security[i].Uv     = uint64(i)
    }
    if statesCount < cPointsCount {
        i = cPointsCount - statesCount;
    } else {
        i = 0
    }
    for ; i < cPointsCount; i++ {
        state, _ := h.layers[index]
        index++
        message.Smeshers[i].Amt     = state.smeshers
        message.Transactions[i].Amt = state.transactions
        message.Accounts[i].Amt     = state.accounts
        message.Circulation[i].Amt  = state.circulation
        message.Rewards[i].Amt      = state.rewards
        message.Security[i].Amt     = state.security
    }
    return message
}

func (h *History) createMockState() {
    state := &State{}
    if h.current != nil {
        state.layer = &Layer{index: h.current.layer.index + 1}
        state.initMock(h.current)
    } else {
        state.layer = &Layer{index: 1}
        state.initEmptyMock()
    }
    h.layers[state.layer.index] = state
    h.current = state;
    message := h.createMessage()
    h.push(message)
}

func randUint(min int, max int) int {
    return min + rand.Intn(max-min)
}

func (s *State) initEmptyMock() {
    s.capacity     = uint64(randUint(0, 100))
    s.decentral    = uint64(randUint(0, 100))
    s.smeshers     = uint64(randUint(0, 100))
    s.transactions = uint64(randUint(0, 100))
    s.accounts     = uint64(randUint(0, 100))
    s.circulation  = uint64(randUint(0, 100))
    s.rewards      = uint64(randUint(0, 100))
    s.security     = uint64(randUint(0, 100))
}

func (s *State) initMock(prev *State) {
    s.capacity     = uint64(randUint(0, 100))
    s.decentral    = uint64(randUint(0, 100))
    s.smeshers     = prev.smeshers + uint64(randUint(0, 100))
    s.transactions = prev.transactions + uint64(randUint(0, 100))
    s.accounts     = prev.accounts + uint64(randUint(0, 100))
    s.circulation  = prev.circulation + uint64(randUint(0, 100))
    s.rewards      = prev.rewards + uint64(randUint(0, 100))
    s.security     = prev.security + uint64(randUint(0, 100))
}

func (m *Message) toJson() []byte {
    b, _ := json.Marshal(m)
    return b
}

func (h *History) push(m *Message) {
    h.bus.notify <- m.toJson()
}

func NewHistory(bus *Bus) *History {
    return &History{bus: bus, current: nil, genesis: time.Now(), layers: make(map[sm.LayerID]*State)}
}

func (h *History) Run() {
}

func (h *History) RunMock() {
    h.epoch = 1
    h.genesis = time.Now()
    time.Sleep(5 * time.Second)
    for {
        h.createMockState()
        time.Sleep(15 * time.Second)
    }
}
