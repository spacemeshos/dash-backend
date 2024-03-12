package history

import (
	"context"

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
	ctx    context.Context
	cancel context.CancelFunc

	bus     *client.Bus
	storage *storage.Storage
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
		info, err := history.storage.GetNetworkInfo(history.ctx)
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
	period := time.Duration(h.storage.NetworkInfo.LayerDuration) * time.Second / 2
	if period > time.Minute {
		period = time.Minute
	}
	for {
		if h.storage.NetworkInfo.LayerDuration > 0 {
			time.Sleep(period)
		} else {
			time.Sleep(15 * time.Second)
		}
		h.pushStatistics()
	}
}

func (h *History) GetStorage() *storage.Storage {
	return h.storage
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
	message := h.GetStats()
	h.bus.Notify <- message.ToJson()
}

func (h *History) push(m *api.Message) {
	h.bus.Notify <- m.ToJson()
}

func (h *History) GetStats() *api.Message {
	var i int

	now := uint32(time.Now().Unix())
	message := &api.Message{}

	networkInfo, err := h.storage.GetNetworkInfo(h.ctx)
	if err == nil {
		message.Network = networkInfo.GenesisId
		message.Age = now - networkInfo.GenesisTime
		message.MaxCapacity = networkInfo.MaxTransactionsPerSecond
		message.GenesisTime = networkInfo.GenesisTime
		message.EpochNumLayers = networkInfo.EpochNumLayers
		message.LayerDuration = networkInfo.LayerDuration
		message.LastLayer = networkInfo.LastLayer
		message.LastLayerTimestamp = networkInfo.LastLayerTimestamp
		message.LastApprovedLayer = networkInfo.LastApprovedLayer
		message.LastConfirmedLayer = networkInfo.LastConfirmedLayer
		message.IsSynced = networkInfo.IsSynced
		message.SyncedLayer = networkInfo.SyncedLayer
		message.TopLayer = networkInfo.TopLayer
		message.VerifiedLayer = networkInfo.VerifiedLayer
	} else {
		message.Network = h.storage.NetworkInfo.GenesisId
		message.Age = now - h.storage.NetworkInfo.GenesisTime
		message.MaxCapacity = h.storage.NetworkInfo.MaxTransactionsPerSecond
		message.GenesisTime = h.storage.NetworkInfo.GenesisTime
		message.EpochNumLayers = h.storage.NetworkInfo.EpochNumLayers
		message.LayerDuration = h.storage.NetworkInfo.LayerDuration
	}

	message.SmeshersGeo = make([]types.Geo, 0)

	for i = 0; i < api.PointsCount; i++ {
		message.Smeshers[i].Uv = i
		message.Transactions[i].Uv = i
		message.Accounts[i].Uv = i
		message.Circulation[i].Uv = i
		message.Rewards[i].Uv = i
		message.Security[i].Uv = i
	}

	message.Layer = h.storage.GetLastLayer(h.ctx)
	message.Epoch = message.Layer / h.storage.NetworkInfo.EpochNumLayers

	epochs, err := h.storage.GetEpochsData(h.ctx, &bson.D{}, options.Find().SetSort(bson.D{{"number", -1}}).SetSkip(1).SetLimit(api.PointsCount).SetProjection(bson.D{{"_id", 0}}))

	if err == nil && len(epochs) > 0 {
		message.Capacity = epochs[0].Stats.Current.Capacity
		message.Decentral = epochs[0].Stats.Current.Decentral
		i = api.PointsCount - 1
		for _, epoch := range epochs {
			log.Info("History: stats for epoch %v: %v", epoch.Number, epoch.Stats)
			age := now - epoch.Start
			message.Smeshers[i].Amt = epoch.Stats.Current.Smeshers
			message.Smeshers[i].Epoch = epoch.Number
			message.Smeshers[i].Age = age
			message.Transactions[i].Amt = epoch.Stats.Current.Transactions
			message.Transactions[i].Epoch = epoch.Number
			message.Transactions[i].Age = age
			message.Accounts[i].Amt = epoch.Stats.Current.Accounts
			message.Accounts[i].Epoch = epoch.Number
			message.Accounts[i].Age = age
			message.Circulation[i].Amt = epoch.Stats.Current.Rewards
			message.Circulation[i].Epoch = epoch.Number
			message.Circulation[i].Age = age
			message.Rewards[i].Amt = epoch.Stats.Current.Rewards
			message.Rewards[i].Epoch = epoch.Number
			message.Rewards[i].Age = age
			message.Security[i].Amt = epoch.Stats.Current.Security
			message.Security[i].Epoch = epoch.Number
			message.Security[i].Age = age
			i--
		}
	}

	return message
}
