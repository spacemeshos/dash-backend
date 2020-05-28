package history

import (
    "math/rand"
    "sync"
    "time"
    "reflect"

//    "github.com/spacemeshos/go-spacemesh/log"
    sm "github.com/spacemeshos/go-spacemesh/common/types"

    pb "github.com/spacemeshos/dash-backend/spacemesh"
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
    epoch uint32
    layer *types.Layer
    transactions	map[sm.TransactionID]*types.Transaction

    stats	Stats
}

type History struct {
    bus 	*client.Bus

    network	types.Network

    smeshersGeo	[]types.Geo
    decentral	uint64

    current 	*State

    layers 		map[sm.LayerID]*State
    accounts		map[sm.Address]*types.Account
    smeshers		map[types.SmesherID]sm.LayerID
    mux 		sync.Mutex
}

func newState(layer *types.Layer, h *History) *State {
    state := &State{
        transactions: make(map[sm.TransactionID]*types.Transaction),
    }
    state.update(layer, h)
    return state
}

func (s *State) update(layer *types.Layer, h *History) {
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
    accounts, totalBalance := getAccountsCountAndTotalBalance(h.accounts)
    s.stats.accounts = accounts
    s.stats.circulation = totalBalance
}

func (h *History) SetNetwork(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64) {
    h.network.NetId = netId
    h.network.GenesisTime = genesisTime
    if epochNumLayers == 0 {
        epochNumLayers = 100
    }
    h.network.EpochNumLayers = epochNumLayers
    if maxTransactionsPerSecond == 0 {
        maxTransactionsPerSecond = 100
    }
    h.network.MaxTransactionsPerSecond = maxTransactionsPerSecond
}

func (h *History) AddLayer(layer *types.Layer) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[layer.Index]
    if ok {
        if reflect.DeepEqual(layer.Hash, state.layer.Hash) {
            state.layer.Status = layer.Status
        } else {
            state.update(layer, h)
        }
    } else {
        state = newState(layer, h)
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
    acc, ok := h.accounts[account.Address]
    if !ok {
        h.accounts[account.Address] = account
    } else {
        acc.Balance = account.Balance
    }
    if h.current != nil {
        h.current.stats.accounts = uint64(len(h.accounts))
    }
}

func (h *History) AddReward(reward *types.Reward) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[reward.Layer]
    if ok {
        state.stats.rewards += uint64(reward.Total)
    }
}

func (h *History) AddTransactionReceipt(txReceipt *types.TransactionReceipt) {
    h.mux.Lock()
    defer h.mux.Unlock()
    state, ok := h.layers[txReceipt.Layer_number]
    if ok {
        tx, ok := state.transactions[txReceipt.Id]
        if ok {
            tx.Result = txReceipt.Result
            if tx.Result == pb.TransactionReceipt_EXECUTED && tx.IsATX() {
                h.smeshers[tx.SmesherId] = txReceipt.Layer_number
            }
        }
    }
}

func (h *History) AddTransactionState(txId *sm.TransactionID, txState pb.TransactionState_TransactionStateType) {
    h.mux.Lock()
    defer h.mux.Unlock()
    if h.current != nil {
        tx, ok := h.current.transactions[*txId]
        if ok {
            tx.State = txState
        }
    }
}

func (h *History) createMessage() *types.Message {
    var i int
    var index sm.LayerID = 0
    statesCount := len(h.layers)
    message := &types.Message{}
    message.Network = "TESTNET 0.1"
    message.Age = uint64(time.Now().Unix()) - h.network.GenesisTime
    if h.current != nil {
        message.Epoch = uint64(h.current.layer.Index) / h.network.EpochNumLayers
        message.Layer = uint64(h.current.layer.Index)
        message.Capacity = h.network.MaxTransactionsPerSecond
        message.Decentral = h.current.stats.decentral
    }
    message.SmeshersGeo = h.smeshersGeo
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
        state, ok := h.layers[index]
        index++
        if !ok {
            panic("layer not found!")
        }
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

func getAccountsCountAndTotalBalance(accounts map[sm.Address]*types.Account) (uint64, uint64) {
    var count uint64
    var total uint64
    for _, account := range accounts {
        if account.Balance > 0 {
            count++
            total += uint64(account.Balance)
        }
    }
    return count, total
}

func (h *History) createMockState() {
    state := &State{}
    if h.current != nil {
        state.layer = &types.Layer{Index: h.current.layer.Index + 1}
        state.stats.initMock(&h.current.stats)
    } else {
        state.layer = &types.Layer{Index: 0}
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
        layers: make(map[sm.LayerID]*State),
        accounts: make(map[sm.Address]*types.Account),
        smeshers: make(map[types.SmesherID]sm.LayerID),
    }
}

func (h *History) Run() {
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
