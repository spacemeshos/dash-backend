package history

import (
    "math"
    "reflect"

    sm "github.com/spacemeshos/go-spacemesh/common/types"
    "github.com/spacemeshos/go-spacemesh/log"

    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
    "github.com/spacemeshos/dash-backend/types"
)

func newEpoch(h *History, number uint64, prev *Epoch) *Epoch {
    return &Epoch{
        history: h,
        prev: prev,
        number: number,
        smeshers: make(map[types.SmesherID]*types.Smesher),
        layers: make(map[sm.LayerID]*types.Layer),
    }
}

func (epoch *Epoch) end() {
    log.Info("Epoch: end")
    epoch.confirmed = (epoch.prev == nil) || (epoch.prev != nil && epoch.prev.confirmed)
    if epoch.confirmed {
        epoch.computeStatistics(&epoch.stats.current)
        if epoch.prev != nil {
            epoch.stats.cumulative.capacity     = epoch.stats.current.capacity
            epoch.stats.cumulative.decentral    = epoch.stats.current.decentral
            epoch.stats.cumulative.smeshers     = epoch.prev.stats.cumulative.smeshers + epoch.stats.current.smeshers
            epoch.stats.cumulative.transactions = epoch.prev.stats.cumulative.transactions + epoch.stats.current.transactions
            epoch.stats.cumulative.accounts     = epoch.stats.current.accounts
            epoch.stats.cumulative.circulation  = epoch.stats.current.circulation
            epoch.stats.cumulative.rewards      = epoch.prev.stats.cumulative.rewards + epoch.stats.current.rewards
            epoch.stats.cumulative.security     = epoch.prev.stats.current.security
        } else {
            epoch.stats.cumulative = epoch.stats.current
        }
        epoch.history.store(epoch)
    }
}

func allLayersConfirmed(layers map[sm.LayerID]*types.Layer) bool {
    for _, layer := range layers {
        if layer.Status != pb.Layer_LAYER_STATUS_CONFIRMED {
            return false
        }
    }
    return true
}

func (epoch *Epoch) addLayer(l *types.Layer) {
    log.Info("Epoch: add layer %v with status %v", l.Number, l.Status)
    layer, ok := epoch.layers[l.Number]
    if !ok {
        layer = l
        epoch.layers[l.Number] = l
    } else {
        if reflect.DeepEqual(layer.Hash, l.Hash) {
            layer.Status = l.Status
        } else {
            layer = l
            epoch.layers[l.Number] = l
        }
    }
    if epoch.lastLayer == nil || epoch.lastLayer.Number < l.Number {
        epoch.lastLayer = layer
    }
    if layer.Status == pb.Layer_LAYER_STATUS_APPROVED {
        if epoch.lastApprovedLayer == nil || epoch.lastApprovedLayer.Number < l.Number {
            epoch.lastApprovedLayer = layer
        }
    } else if layer.Status == pb.Layer_LAYER_STATUS_CONFIRMED {
        if epoch.lastConfirmedLayer == nil || epoch.lastConfirmedLayer.Number < l.Number {
            epoch.lastConfirmedLayer = layer
        }
        layer.Print()
        if uint64(len(epoch.layers)) == epoch.history.network.EpochNumLayers && allLayersConfirmed(epoch.layers) {
            epoch.end()
        }
    }
}

func (epoch *Epoch) addReward(reward uint64) {
    epoch.stats.current.rewards += reward
}

func (epoch *Epoch) addTransactionReceipt(txReceipt *types.TransactionReceipt) {
    layer, ok := epoch.layers[txReceipt.Layer_number]
    if ok {
        tx, ok := layer.Transactions[txReceipt.Id]
        if ok {
            tx.SetResult(txReceipt.Result)
        }
    }
}

func getTransactionsCount(layer *types.Layer) uint64 {
    return uint64(len(layer.Transactions))
}

func (epoch *Epoch) computeStatistics(stats *Statistics) {
    // log.Info("Epoch: computeStatistics")
    duration := float64(epoch.history.network.LayerDuration) * float64(len(epoch.layers))
    if duration > 0 && epoch.history.network.MaxTransactionsPerSecond > 0 {
        stats.capacity = uint64(math.Round(((float64(stats.transactions) / duration) / float64(epoch.history.network.MaxTransactionsPerSecond)) * 100.0))
    }
    if len(epoch.smeshers) > 0 {
        epoch.smeshers = make(map[types.SmesherID]*types.Smesher)
    }
    for _, layer := range epoch.layers {
        stats.transactions += getTransactionsCount(layer)
        for _, atx := range layer.Activations {
            smesher, ok := epoch.smeshers[atx.Smesher_id]
            if ok {
                stats.security += atx.Commitment_size
                smesher.Commitment_size = atx.Commitment_size
            } else {
                epoch.smeshers[atx.Smesher_id] = &types.Smesher{Id: atx.Smesher_id, Commitment_size: atx.Commitment_size}
            }
        }
    }
    stats.smeshers = uint64(len(epoch.smeshers))
    // degree_of_decentralization is defined as: 0.5 * (min(n,1e4)^2/1e8) + 0.5 * (1 - gini_coeff(last_100_epochs))
    a := math.Min(float64(len(epoch.smeshers)), 1e4)
    stats.decentral = uint64(100.0 * (0.5 * (a * a) /1e8  + 0.5 * (1.0 - gini(epoch.smeshers))))
    for _, account := range epoch.history.accounts {
        if account.Balance > 0 {
            stats.accounts++
            stats.circulation += uint64(account.Balance)
        }
    }
    stats.rewards = epoch.stats.current.rewards
    // log.Info("Epoch Statistics:")
    // log.Info("    capacity: %v", stats.capacity)
    // log.Info("    decentral: %v", stats.decentral)
    // log.Info("    smeshers: %v", stats.smeshers)
    // log.Info("    transactions: %v", stats.transactions)
    // log.Info("    accounts: %v", stats.accounts)
    // log.Info("    circulation: %v", stats.circulation)
    // log.Info("    rewards: %v", stats.rewards)
    // log.Info("    security: %v", stats.security)
}

func (epoch *Epoch) GetStatistics(stats *Statistics) {
    log.Info("Epoch: GetStatistics for epoch %v", epoch.number)
    if epoch.confirmed {
        *stats = epoch.stats.cumulative
    } else {
        var current Statistics
        epoch.computeStatistics(&current)
        if epoch.prev != nil {
            stats.capacity     = current.capacity
            stats.decentral    = current.decentral
            stats.smeshers     = epoch.prev.stats.cumulative.smeshers + current.smeshers
            stats.transactions = epoch.prev.stats.cumulative.transactions + current.transactions
            stats.accounts     = current.accounts
            stats.circulation  = current.circulation
            stats.rewards      = epoch.prev.stats.cumulative.rewards + current.rewards
            stats.security     = epoch.prev.stats.current.security
        } else {
            *stats = current
        }
    }
}

/*
func (s *State) update(layer *types.Layer) {
    s.layer = layer

    s.stats.capacity = 0
    s.stats.decentral = 0
    s.stats.smeshers = 0
    s.stats.accounts = 0
    s.stats.transactions = 0
    s.stats.circulation = 0
    s.stats.rewards = 0
    s.stats.security = 0

    for _, block := range layer.Blocks {
        for _, tx := range block.Transactions {
            _, ok := s.transactions[*tx.GetID()]
            if !ok {
                s.transactions[*tx.GetID()] = tx
            }
        }
    }

    stats.transactions = uint64(len(txs))

    for _, atx := range layer.Activations {
        history.smeshers[atx.Smesher_id] = atx.Layer
        atx.Commitment_size
    }

    for _, account := range accounts {
        if account.Balance > 0 {
            s.stats.accounts++
            s.stats.circulation += uint64(account.Balance)
        }
    }
}
*/
