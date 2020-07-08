package history

import (
    "time"

//    "github.com/spacemeshos/dash-backend/types"
    "github.com/spacemeshos/dash-backend/mock"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

type proxy struct {
    history	*History
}

func (h *History) RunMock(netId int, epochLayers int, maxTxs int, layerDuration int) {
    network := mock.NewNetwork(&proxy{h}, netId, epochLayers, maxTxs, layerDuration)

//    h.smeshersGeo = append(h.smeshersGeo,
//        types.Geo{Name: "Tel Aviv", Coordinates: [2]float64{34.78057, 32.08088}},
//        types.Geo{Name: "New York", Coordinates: [2]float64{-74.00597, 40.71427}},
//        types.Geo{Name: "Chernihiv", Coordinates: [2]float64{31.28487, 51.50551}},
//        types.Geo{Name: "Montreal", Coordinates: [2]float64{-73.58781, 45.50884}},
//        types.Geo{Name: "Kyiv", Coordinates: [2]float64{30.5238, 50.45466}},
//    )
    for {
        network.Tick()
        time.Sleep(1 * time.Second)
    }
}

func (proxy *proxy) SetNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64) {
    proxy.history.SetNetworkInfo(
        netId,
        genesisTime,
        epochNumLayers,
        maxTransactionsPerSecond,
        layerDuration,
    )
}

func (proxy *proxy) OnLayerChanged(layer *pb.Layer) {
}

func (proxy *proxy) OnAccountChanged(account *pb.Account) {
}

func (proxy *proxy) OnReward(reward *pb.Reward) {
}

func (proxy *proxy) OnTransactionReceipt(receipt *pb.TransactionReceipt) {
}
