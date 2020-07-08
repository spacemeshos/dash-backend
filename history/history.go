package history

import (
//    "context"
    "errors"
//    "reflect"
    "time"

    "github.com/spacemeshos/go-spacemesh/log"
    sm "github.com/spacemeshos/go-spacemesh/common/types"

//    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/types"

//    "go.mongodb.org/mongo-driver/bson"
)

func (h *History) SetNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64) {
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
    h.network.LayerDuration = layerDuration
}

func (h *History) GetStatistics(epochNumber uint64, stats *Stats) {
    stats.capacity = 0
    stats.decentral = 0
    stats.smeshers = 0
    stats.accounts = 0
    stats.transactions = 0
    stats.circulation = 0
    stats.rewards = 0
    stats.security = 0

    epoch, ok := h.epochs[epochNumber]
    if ok {
        epoch.GetStatistics(stats)
    }
}

func (h *History) AddLayer(layer *types.Layer) {
    h.mux.Lock()
    defer h.mux.Unlock()

    log.Info("History: add layer %v with status %v", layer.Number, layer.Status)

    epochNumber := uint64(layer.Number) / h.network.EpochNumLayers
    epoch, ok := h.epochs[epochNumber]
    if !ok {
        if epochNumber == 0 {
            epoch = newEpoch(h, epochNumber, nil)
        } else {
            prev, _ := h.epochs[epochNumber - 1]
            epoch = newEpoch(h, epochNumber, prev)
        }
        h.epochs[epochNumber] = epoch
    }

    if h.epoch == nil {
        h.epoch = epoch
    } else {
        if epoch.number > h.epoch.number {
            h.epoch = epoch
        }
    }

    epoch.addLayer(layer)
}

func (h *History) AddAccount(account *types.Account) {
    h.mux.Lock()
    defer h.mux.Unlock()

    log.Info("History: add account with balance %v", account.Balance)
    acc, ok := h.accounts[account.Address]
    if !ok {
        h.accounts[account.Address] = account
    } else {
        acc.Balance = account.Balance
    }
}

func (h *History) AddReward(reward *types.Reward) {
    h.mux.Lock()
    defer h.mux.Unlock()

    log.Info("History: add reward %v", reward.Total)
    epochNumber := uint64(reward.Layer) / h.network.EpochNumLayers
    epoch, ok := h.epochs[epochNumber]
    if ok {
        epoch.addReward(uint64(reward.Total))
    }
}

func (h *History) AddTransactionReceipt(txReceipt *types.TransactionReceipt) {
    h.mux.Lock()
    defer h.mux.Unlock()

    log.Info("History: add transaction receipt")
    epochNumber := uint64(txReceipt.Layer_number) / h.network.EpochNumLayers
    epoch, ok := h.epochs[epochNumber]
    if ok {
        epoch.addTransactionReceipt(txReceipt)
    }
}

func (h *History) pushStatistics() {
    var i uint64
    var layerNumber int = -1
    var epochNumber int = -1

    if h.epoch != nil {
        epochNumber = int(h.epoch.number)
        if h.epoch.lastLayer != nil {
            layerNumber = int(h.epoch.lastLayer.Number)
        }
    }

    log.Info("History: pushStatistics %v/%v", layerNumber, epochNumber)

    message := &types.Message{}
    message.Network = "TESTNET 0.1"
    message.Age = uint64(time.Now().Unix()) - h.network.GenesisTime
    message.SmeshersGeo = h.smeshersGeo

    for i = 0; i < types.PointsCount; i++ {
        message.Smeshers[i].Uv     = uint64(i)
        message.Transactions[i].Uv = uint64(i)
        message.Accounts[i].Uv     = uint64(i)
        message.Circulation[i].Uv  = uint64(i)
        message.Rewards[i].Uv      = uint64(i)
        message.Security[i].Uv     = uint64(i)
    }

    if h.epoch != nil && h.epoch.lastLayer != nil {
        var stats Stats
        var epochCount uint64
        var epochNumber uint64

        epochNumber = h.epoch.number
        message.Epoch = epochNumber
        message.Layer = uint64(h.epoch.lastLayer.Number)

        i = types.PointsCount - 1
        epochCount = h.epoch.number + 1
        if epochCount > types.PointsCount {
            epochCount = types.PointsCount
        }

        h.GetStatistics(epochNumber, &stats)
        message.Capacity = stats.capacity
        if epochCount > 1 {
            h.GetStatistics(epochNumber - 1, &stats)
            message.Decentral = stats.decentral
        }

        for ; epochCount > 0;  {
            h.GetStatistics(epochNumber, &stats)
            log.Info("History: stats for epoch %v: %v", epochNumber, stats)
            message.Smeshers[i].Amt     = stats.smeshers
            message.Transactions[i].Amt = stats.transactions
            message.Accounts[i].Amt     = stats.accounts
            message.Circulation[i].Amt  = stats.circulation
            message.Rewards[i].Amt      = stats.rewards
            message.Security[i].Amt     = stats.security
            i--
            epochCount--
            epochNumber--
        }
    }

    h.bus.Notify <- message.ToJson()
}

func (h *History) store(epoch *Epoch) error {
//    if h.storage.db != nil {
//        return h.storage.putEpoch(epoch)
//    }
    return errors.New("No Database")
}

func (h *History) push(m *types.Message) {
    h.bus.Notify <- m.ToJson()
}

func NewHistory(bus *client.Bus) *History {
    return &History{
        bus: bus,
        accounts: make(map[sm.Address]*types.Account),
        epochs: make(map[uint64]*Epoch),
    }
}

func (h *History) Run() {
//    err := h.storage.open()
//    if err != nil {
//        panic("Error open MongoDB")
//    }
//    defer h.storage.close()
    h.smeshersGeo = append(h.smeshersGeo,
        types.Geo{Name: "Tel Aviv", Coordinates: [2]float64{34.78057, 32.08088}},
        types.Geo{Name: "New York", Coordinates: [2]float64{-74.00597, 40.71427}},
        types.Geo{Name: "Chernihiv", Coordinates: [2]float64{31.28487, 51.50551}},
        types.Geo{Name: "Montreal", Coordinates: [2]float64{-73.58781, 45.50884}},
        types.Geo{Name: "Kyiv", Coordinates: [2]float64{30.5238, 50.45466}},
    )
    for {
        h.pushStatistics()
        time.Sleep(15 * time.Second)
    }
}
