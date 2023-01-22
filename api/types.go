package api

import (
	"encoding/json"
	"github.com/spacemeshos/dash-backend/types"
)

const PointsCount = 50

type Point struct {
	Uv    int    `json:"uv"`
	Epoch int32  `json:"epoch"`
	Age   uint32 `json:"age"`
	Amt   int64  `json:"amt"`
}

type Message struct {
	Network            string `json:"network"`
	Age                uint32 `json:"age"`
	Layer              uint32 `json:"layer"`
	Epoch              uint32 `json:"epoch"`
	GenesisTime        uint32 `json:"genesis"`
	EpochNumLayers     uint32 `json:"epochnumlayers"`
	LayerDuration      uint32 `json:"layerduration"`
	MaxCapacity        uint32 `json:"maxcapacity"`
	LastLayer          uint32 `json:"lastlayer"`
	LastLayerTimestamp uint32 `json:"lastlayerts"`
	LastApprovedLayer  uint32 `json:"lastapprovedlayer"`
	LastConfirmedLayer uint32 `json:"lastconfirmedlayer"`
	IsSynced           bool   `json:"issynced"`
	SyncedLayer        uint32 `json:"syncedlayer"`
	TopLayer           uint32 `json:"toplayer"`
	VerifiedLayer      uint32 `json:"verifiedlayer"`

	Capacity     int64              `json:"capacity"`
	Decentral    int64              `json:"decentral"`
	SmeshersGeo  []types.Geo        `json:"smeshersGeo"`
	Smeshers     [PointsCount]Point `json:"smeshers"`
	Transactions [PointsCount]Point `json:"transactions"`
	Accounts     [PointsCount]Point `json:"accounts"`
	Circulation  [PointsCount]Point `json:"circulation"`
	Rewards      [PointsCount]Point `json:"rewards"`
	Security     [PointsCount]Point `json:"security"`
}

func (m *Message) ToJson() []byte {
	b, _ := json.Marshal(m)
	return b
}
