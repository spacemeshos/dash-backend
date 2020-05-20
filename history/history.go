package history

import (
    "math/rand"
    "sync"
    "time"
    "reflect"

//    "github.com/spacemeshos/go-spacemesh/log"
    sm "github.com/spacemeshos/go-spacemesh/common/types"

    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/types"
)

type Stats struct {
    capacity		uint64
    decentral		uint64
    smeshers		uint64
    accounts		uint64
    transactions	uint64
    circulation		uint64
    rewards		uint64
    security		uint64
}

type State struct {
    layer *types.Layer
    transactions	map[sm.TransactionID]*types.Transaction

    stats	Stats
}

type History struct {
    bus 	*client.Bus

    smeshersGeo	[]types.Geo
    genesis	time.Time
    epoch	uint64
    capacity	uint64
    decentral	uint64

    current 	*State

    layers 		map[sm.LayerID]*State
    accounts		map[sm.Address]*types.Account
    smeshers		map[sm.Address]*types.Account
    mux 		sync.Mutex
}

func newState(layer *types.Layer) *State {
    state := &State{
        transactions: make(map[sm.TransactionID]*types.Transaction),
    }
    state.update(layer)
    return state
}

func (s *State) update(layer *types.Layer) {
    s.layer = layer
    for _, block := range layer.Blocks {
        for _, tx := range block.Transactions {
            _, ok := s.transactions[tx.Id]
            if !ok {
                s.transactions[tx.Id] = tx
            }
        }
    }
    s.stats.transactions = getTransactionsCount(s.transactions)
}

func (h *History) AddLayer(layer *types.Layer) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[layer.Index]
    if ok {
        if reflect.DeepEqual(layer.Hash, state.layer.Hash) {
            state.layer.Status = layer.Status
        } else {
            state.update(layer)
        }
    } else {
        state = newState(layer)
        h.layers[layer.Index] = state
    }
    if h.current != nil {
        if h.current.layer.Index < state.layer.Index  {
            h.current = state
        }
    } else {
        h.current = state
    }
    message := h.createMessage()
    h.push(message)
}

func (h *History) AddAccount(account *types.Account) {
    h.mux.Lock()
    defer h.mux.Unlock()
    _, ok := h.accounts[account.Address]
    if !ok {
        h.accounts[account.Address] = account
    }
    if h.current != nil {
        h.current.stats.accounts = uint64(len(h.accounts))
    }
}

func (h *History) AddReward(reward *types.Reward) {
    h.mux.Lock()
    defer h.mux.Unlock()
}

func (h *History) AddTransactionReceipt(txReceipt *types.TransactionReceipt) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[txReceipt.Layer_number]
    if ok {
        tx, ok := state.transactions[txReceipt.Id]
        if ok {
            tx.Result = txReceipt.Result
            if tx.IsATX() {
                _, ok := h.smeshers[tx.Smesher_id]
                h.smeshers[tx.Smesher_id] = tx.Smesher_id
                if h.current != nil {
                    h.current.stats.smeshers = uint64(len(h.smeshers))
                }
            }
        }
    }
}

func (h *History) createMessage() *types.Message {
    var i int
    var index sm.LayerID = 1
    statesCount := len(h.layers)
    message := &types.Message{}
    message.Network = "TESTNET 0.1"
    message.Age = uint64(time.Now().Second() - h.genesis.Second())
    message.Epoch = h.epoch
    if h.current != nil {
        message.Layer = uint64(h.current.layer.Index)
        message.Capacity = h.current.stats.capacity
        message.Decentral = h.current.stats.decentral
    }
//    message.SmeshersGeo		[]Geo
    for i = 0; i < types.PointsCount; i++ {
        message.Smeshers[i].Uv     = uint64(i)
        message.Transactions[i].Uv = uint64(i)
        message.Accounts[i].Uv     = uint64(i)
        message.Circulation[i].Uv  = uint64(i)
        message.Rewards[i].Uv      = uint64(i)
        message.Security[i].Uv     = uint64(i)
    }
    if statesCount < types.PointsCount {
        i = types.PointsCount - statesCount;
    } else {
        i = 0
    }
    for ; i < types.PointsCount; i++ {
        state, _ := h.layers[index]
        index++
        message.Smeshers[i].Amt     = state.stats.smeshers
        message.Transactions[i].Amt = state.stats.transactions
        message.Accounts[i].Amt     = state.stats.accounts
        message.Circulation[i].Amt  = state.stats.circulation
        message.Rewards[i].Amt      = state.stats.rewards
        message.Security[i].Amt     = state.stats.security
    }
    return message
}

func getTransactionsCount(txs map[sm.TransactionID]*types.Transaction) uint64 {
    var count uint64
    for _, tx := range txs {
        if tx.IsExecuted() {
            if tx.IsSimple() || tx.IsAPP() {
                count++
            }
        }
    }
    return count
}

func (h *History) createMockState() {
    state := &State{}
    if h.current != nil {
        state.layer = &types.Layer{Index: h.current.layer.Index + 1}
        state.stats.initMock(&h.current.stats)
    } else {
        state.layer = &types.Layer{Index: 1}
        state.stats.initEmptyMock()
    }
    h.layers[state.layer.Index] = state
    h.current = state;
    message := h.createMessage()
    h.push(message)
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

func (h *History) push(m *types.Message) {
    h.bus.Notify <- m.ToJson()
}

func NewHistory(bus *client.Bus) *History {
    return &History{
        bus: bus,
        current: nil,
        genesis: time.Now(),
        layers: make(map[sm.LayerID]*State),
        accounts: make(map[sm.Address]*types.Account),
        smeshers: make(map[sm.Address]*types.Account),
    }
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
