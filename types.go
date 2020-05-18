package main

import (
//    "encoding/json"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    pb "github.com/spacemeshos/dash-backend/api/proto/spacemesh"
)

const cPointsCount = 50

type Point struct {
    Uv	uint64 `json:"uv"`
    Amt	uint64 `json:"amt"`
}

type Geo struct {
    Name	string `json:"name"`
    Coordinates	[2]float64 `json:"coordinates"`
}

type Message struct {
    Network	string `json:"network"`
    Age		uint64 `json:"age"`
    Layer	uint64 `json:"layer"`
    Epoch	uint64 `json:"epoch"`
    Capacity	uint64 `json:"capacity"`
    Decentral	uint64 `json:"decentral"`
    SmeshersGeo		[]Geo   `json:"smeshersGeo"`
    Smeshers		[cPointsCount]Point `json:"smeshers"`
    Transactions	[cPointsCount]Point `json:"transactions"`
    Accounts		[cPointsCount]Point `json:"accounts"`
    Circulation		[cPointsCount]Point `json:"circulation"`
    Rewards		[cPointsCount]Point `json:"rewards"`
    Security		[cPointsCount]Point `json:"security"`
}

type Amount uint64

type TransactionFee struct {
    gas_consumed	uint64
    gas_price		uint64
    // tx_fee = gas_consumed * gas_price
}

type Account struct {
    address	sm.Address	// account public address
    counter	uint64		// aka nonce
    balance	Amount		// known account balance
}

type Reward struct {
    layer		sm.LayerID
    total		Amount
    layer_reward	Amount
    layer_computed	sm.LayerID	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    coinbase		sm.Address	// account awarded this reward
    smesher		sm.Address	// it will be nice to always have this in reward events
}

type TransactionReceipt struct {
    id			sm.TransactionID// the source transaction
    // The results of STF transaction processing
    result		pb.TransactionReceipt_TransactionResult
    gas_used		uint64		// gas units used by the transaction (gas price in tx)
    fee			Amount		// transaction fee charged for the transaction
    layer_number	sm.LayerID	// The layer in which the STF processed this transaction
}

type Transaction struct {
    txType	pb.Transaction_TransactionType
    id		sm.TransactionID
    sender	sm.Address
    fee		TransactionFee
    timestamp	uint64		// shouldn't this be part of the event envelope?
    receiver	sm.Address	// depending on tx type
    amount	Amount		// amount of coin transfered in this tx by sender
    counter	uint64		// tx counter aka nonce
    data	[]byte		// binary payload - used for app, deploy, atx and spwan transactions
    smesher_id	sm.Address	// used in atx only
    prev_atx	sm.TransactionID// previous ATX. used in atx.
}

type Block struct {
    id			sm.BlockID
    transactions	[]*Transaction
}

type Layer struct {
    index	sm.LayerID
    blocks	[]*Block
}
