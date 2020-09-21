package history

import (
    "context"
//    "errors"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/spacemeshos/go-spacemesh/log"

//    "github.com/spacemeshos/explorer-backend/model"
    "github.com/spacemeshos/explorer-backend/storage"

    "github.com/spacemeshos/dash-backend/api"
    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/types"
)

type History struct {
    ctx      context.Context
    cancel   context.CancelFunc

    bus      *client.Bus
    storage  *storage.Storage
}

func NewHistory(ctx context.Context, bus *client.Bus, dbUrl string, dbName string) (*History, error) {
    var err error

    log.Info("Create new History service")

    history := &History{
        bus: bus,
    }

    if ctx == nil {
        history.ctx, history.cancel = context.WithCancel(context.Background())
    } else {
        history.ctx, history.cancel = context.WithCancel(ctx)
    }

    if history.storage, err = storage.New(history.ctx, dbUrl, dbName); err != nil {
        return nil, err
    }

    for history.storage.NetworkInfo.GenesisTime == 0 {
        info, err :=  history.storage.GetNetworkInfo(history.ctx)
        if err == nil {
            log.Info("Readed network info: %+v", info)
            history.storage.NetworkInfo = *info
            log.Info("Storage network info: %+v", history.storage.NetworkInfo)
            break
        }
        log.Info("No network info found in database. Wait for collector.")
        time.Sleep(1 * time.Second)
    }

    log.Info("New History service is created")

    return history, nil
}

func (h *History) Run() {
    for {
        if h.storage.NetworkInfo.LayerDuration > 0 {
            time.Sleep(time.Duration(h.storage.NetworkInfo.LayerDuration) * time.Second / 2)
        } else {
            time.Sleep(15 * time.Second)
        }
        h.pushStatistics()
    }
}

func getObject(d *bson.D, name string) *bson.E {
    for _, obj := range *d {
        if obj.Key == name {
            return &obj
        }
    }
    return nil
}

func (h *History) pushStatistics() {
    var i int

    message := &api.Message{}
    message.Network = ""
    message.Age = uint32(time.Now().Unix()) - h.storage.NetworkInfo.GenesisTime
    message.SmeshersGeo = make([]types.Geo, 0)

    for i = 0; i < api.PointsCount; i++ {
        message.Smeshers[i].Uv     = i
        message.Transactions[i].Uv = i
        message.Accounts[i].Uv     = i
        message.Circulation[i].Uv  = i
        message.Rewards[i].Uv      = i
        message.Security[i].Uv     = i
    }

    message.Layer = h.storage.GetLastLayer(h.ctx)
    message.Epoch = message.Layer / h.storage.NetworkInfo.EpochNumLayers

    epochs, err := h.storage.GetEpochsData(h.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(api.PointsCount).SetProjection(bson.D{{"_id", 0}}))

    if err == nil {
        i = api.PointsCount - 1
        for _, epoch := range epochs {
            log.Info("History: stats for epoch %v: %v", epoch.Number, epoch.Stats)
            message.Smeshers[i].Amt     = epoch.Stats.Cumulative.Smeshers
            message.Transactions[i].Amt = epoch.Stats.Cumulative.Transactions
            message.Accounts[i].Amt     = epoch.Stats.Cumulative.Accounts
            message.Circulation[i].Amt  = epoch.Stats.Cumulative.Circulation
            message.Rewards[i].Amt      = epoch.Stats.Cumulative.Rewards
            message.Security[i].Amt     = epoch.Stats.Cumulative.Security
            i--
        }
    }

    fmt.Print(message, "\n")

    h.bus.Notify <- message.ToJson()
}

func (h *History) push(m *api.Message) {
    h.bus.Notify <- m.ToJson()
}
