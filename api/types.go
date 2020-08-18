package api

import (
    "encoding/json"
    "github.com/spacemeshos/dash-backend/types"
)

const PointsCount = 50

type Point struct {
    Uv	uint64 `json:"uv"`
    Amt	uint64 `json:"amt"`
}

type Message struct {
    Network	string `json:"network"`
    Age		uint64 `json:"age"`
    Layer	uint64 `json:"layer"`
    Epoch	uint64 `json:"epoch"`
    Capacity	uint64 `json:"capacity"`
    Decentral	uint64 `json:"decentral"`
    SmeshersGeo		[]types.Geo        `json:"smeshersGeo"`
    Smeshers		[PointsCount]Point `json:"smeshers"`
    Transactions	[PointsCount]Point `json:"transactions"`
    Accounts		[PointsCount]Point `json:"accounts"`
    Circulation		[PointsCount]Point `json:"circulation"`
    Rewards		[PointsCount]Point `json:"rewards"`
    Security		[PointsCount]Point `json:"security"`
}

func (m *Message) ToJson() []byte {
    b, _ := json.Marshal(m)
    return b
}
