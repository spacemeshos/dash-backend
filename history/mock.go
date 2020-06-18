package history

import (
    "math/rand"
    "time"

    "github.com/spacemeshos/dash-backend/types"
)

func (h *History) createMockState() {
/*
    state := &State{}
    if h.current != nil {
        state.layer = &types.Layer{Number: h.current.layer.Number + 1}
        state.stats.initMock(&h.current.stats)
    } else {
        state.layer = &types.Layer{Number: 0}
        state.stats.initEmptyMock()
    }
    h.layers[state.layer.Number] = state
    h.current = state;
    message := h.createMessage()
    h.push(message)
*/
}

func randUint(min int, max int) int {
    return min + rand.Intn(max-min)
}

func (s *Stats) initEmptyMock() {
    s.capacity     = uint64(randUint(0, 100))
    s.decentral    = uint64(randUint(0, 100))
    s.smeshers     = uint64(randUint(0, 100))
    s.transactions = uint64(randUint(0, 100))
    s.accounts     = uint64(randUint(0, 100))
    s.circulation  = uint64(randUint(0, 100))
    s.rewards      = uint64(randUint(0, 100))
    s.security     = uint64(randUint(0, 100))
}

func (s *Stats) initMock(prev *Stats) {
    s.capacity     = uint64(randUint(0, 100))
    s.decentral    = uint64(randUint(0, 100))
    s.smeshers     = prev.smeshers + uint64(randUint(0, 100))
    s.transactions = prev.transactions + uint64(randUint(0, 100))
    s.accounts     = prev.accounts + uint64(randUint(0, 100))
    s.circulation  = prev.circulation + uint64(randUint(0, 100))
    s.rewards      = prev.rewards + uint64(randUint(0, 100))
    s.security     = prev.security + uint64(randUint(0, 100))
}

func (h *History) RunMock() {
    h.network.GenesisTime = uint64(time.Now().Unix())
    h.network.EpochNumLayers = 100
    h.network.MaxTransactionsPerSecond = 100

    time.Sleep(5 * time.Second)
    h.smeshersGeo = append(h.smeshersGeo,
        types.Geo{Name: "Tel Aviv", Coordinates: [2]float64{34.78057, 32.08088}},
        types.Geo{Name: "New York", Coordinates: [2]float64{-74.00597, 40.71427}},
        types.Geo{Name: "Chernihiv", Coordinates: [2]float64{31.28487, 51.50551}},
        types.Geo{Name: "Montreal", Coordinates: [2]float64{-73.58781, 45.50884}},
        types.Geo{Name: "Kyiv", Coordinates: [2]float64{30.5238, 50.45466}},
    )
    for {
        h.createMockState()
        time.Sleep(15 * time.Second)
    }
}
