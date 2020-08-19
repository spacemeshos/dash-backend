package history

import (
//    "context"
    "errors"
    "fmt"
//    "reflect"
    "time"

    "github.com/spacemeshos/go-spacemesh/log"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/types"
    "github.com/spacemeshos/dash-backend/api"

//    "go.mongodb.org/mongo-driver/bson"
)

func (h *History) OnNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64) {
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

func (h *History) GetStatistics(epochNumber uint64, stats *Statistics) {
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

func (h *History) OnLayer(pbLayer *pb.Layer) {
    h.mux.Lock()
    defer h.mux.Unlock()

    layer := types.NewLayer(pbLayer)

//    log.Info("History: add layer %v with status %v", layer.Number, layer.Status)

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

func (h *History) OnAccount(pbAccount *pb.Account) {
    h.mux.Lock()
    defer h.mux.Unlock()

    account := types.NewAccount(pbAccount)

    log.Info("History: add account with balance %v", account.Balance)
    acc, ok := h.accounts[account.Address]
    if !ok {
        h.accounts[account.Address] = account
    } else {
        acc.Balance = account.Balance
    }
}

func (h *History) addAccountAmount(address *types.Address, amount types.Amount) {
    acc, ok := h.accounts[*address]
    if !ok {
        h.accounts[*address] = &types.Account{*address, 0, amount}
    } else {
        acc.Balance += amount
    }
}

func (h *History) OnReward(pbReward *pb.Reward) {
    h.mux.Lock()
    defer h.mux.Unlock()

    reward := types.NewReward(pbReward)

//    log.Info("History: add reward %v", reward.Total)
    epochNumber := uint64(reward.Layer) / h.network.EpochNumLayers
    h.addAccountAmount(&reward.Coinbase, reward.Total)
    epoch, ok := h.epochs[epochNumber]
    if ok {
        epoch.addReward(uint64(reward.Total))
    }
}

func (h *History) OnTransactionReceipt(pbTxReceipt *pb.TransactionReceipt) {
    h.mux.Lock()
    defer h.mux.Unlock()

    txReceipt := types.NewTransactionReceipt(pbTxReceipt)

//    log.Info("History: add transaction receipt")
    epochNumber := uint64(txReceipt.Layer_number) / h.network.EpochNumLayers
    epoch, ok := h.epochs[epochNumber]
    if ok {
        epoch.addTransactionReceipt(txReceipt)
    }
}

func (h *History) getSmesher(id *types.SmesherID) *types.Smesher {
    smesher, ok := h.smeshers[*id]
    if ok {
        return smesher
    } else {
        return nil
    }
}

func (h *History) addSmesher(id *types.SmesherID, commitmentSize uint64) *types.Smesher {
    smesher, ok := h.smeshers[*id]
    if ok {
        return smesher
    }
    smesher = &types.Smesher{Id: *id, Commitment_size: commitmentSize, Geo: getRandomGeo()}
//    log.Info("Add smesher from %v", smesher.Geo)
    h.smeshers[*id] = smesher
    return smesher
}

func (h *History) pushStatistics() {
    h.mux.Lock()
    defer h.mux.Unlock()

    var i uint64
/*
    var layerNumber int = -1
    var epochNumber int = -1

    if h.epoch != nil {
        epochNumber = int(h.epoch.number)
        if h.epoch.lastLayer != nil {
            layerNumber = int(h.epoch.lastLayer.Number)
        }
    }

    log.Info("History: pushStatistics %v/%v", layerNumber, epochNumber)
*/
    message := &api.Message{}
    message.Network = "TESTNET 0.1"
    message.Age = uint64(time.Now().Unix()) - h.network.GenesisTime
    message.SmeshersGeo = make([]types.Geo, 0)

    for i = 0; i < api.PointsCount; i++ {
        message.Smeshers[i].Uv     = uint64(i)
        message.Transactions[i].Uv = uint64(i)
        message.Accounts[i].Uv     = uint64(i)
        message.Circulation[i].Uv  = uint64(i)
        message.Rewards[i].Uv      = uint64(i)
        message.Security[i].Uv     = uint64(i)
    }

    if h.epoch != nil && h.epoch.lastLayer != nil {
        var stats Statistics
        var epochCount uint64
        var epochNumber uint64

        epochNumber = h.epoch.number
        message.Epoch = epochNumber
        message.Layer = uint64(h.epoch.lastLayer.Number)

        i = api.PointsCount - 1
        epochCount = h.epoch.number + 1
        if epochCount > api.PointsCount {
            epochCount = api.PointsCount
        }

        h.GetStatistics(epochNumber, &stats)
        message.Capacity = stats.capacity
        if epochCount > 1 {
            h.GetStatistics(epochNumber - 1, &stats)
            message.Decentral = stats.decentral
        }

        if h.epoch.prev != nil && len(h.epoch.prev.smeshers) > 0 {
            var i int
            message.SmeshersGeo = make([]types.Geo, len(h.epoch.prev.smeshers))
            for _, smesher := range h.epoch.prev.smeshers {
                message.SmeshersGeo[i] = smesher.Geo
                i++
            }
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

    fmt.Print(message, "\n")

    h.bus.Notify <- message.ToJson()
}

func (h *History) store(epoch *Epoch) error {
//    if h.storage.db != nil {
//        return h.storage.putEpoch(epoch)
//    }
    return errors.New("No Database")
}

func (h *History) push(m *api.Message) {
    h.bus.Notify <- m.ToJson()
}

func NewHistory(bus *client.Bus) *History {
    return &History{
        bus: bus,
        smeshers: make(map[types.SmesherID]*types.Smesher),
        accounts: make(map[types.Address]*types.Account),
        epochs: make(map[uint64]*Epoch),
    }
}

func (h *History) Run() {
//    err := h.storage.open()
//    if err != nil {
//        panic("Error open MongoDB")
//    }
//    defer h.storage.close()
    for {
        if h.network.LayerDuration > 0 {
            time.Sleep(time.Duration(h.network.LayerDuration) * time.Second / 2)
        } else {
            time.Sleep(15 * time.Second)
        }
        h.pushStatistics()
    }
}
