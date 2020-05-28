package types

import (
    "encoding/json"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    pb "github.com/spacemeshos/dash-backend/spacemesh"
)

const PointsCount = 50

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
    Smeshers		[PointsCount]Point `json:"smeshers"`
    Transactions	[PointsCount]Point `json:"transactions"`
    Accounts		[PointsCount]Point `json:"accounts"`
    Circulation		[PointsCount]Point `json:"circulation"`
    Rewards		[PointsCount]Point `json:"rewards"`
    Security		[PointsCount]Point `json:"security"`
}

type Network struct {
    NetId			uint64
    GenesisTime			uint64
    EpochNumLayers		uint64
    MaxTransactionsPerSecond	uint64
}

type Amount uint64

type SmesherID [32]byte

type TransactionFee struct {
    Gas_consumed	uint64
    Gas_price		uint64
    // tx_fee = gas_consumed * gas_price
}

type Account struct {
    Address	sm.Address	// account public address
    Counter	uint64		// aka nonce
    Balance	Amount		// known account balance
}

type Reward struct {
    Layer		sm.LayerID
    Total		Amount
    Layer_reward	Amount
    Layer_computed	sm.LayerID	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    Coinbase		sm.Address	// account awarded this reward
    Smesher		SmesherID	// it will be nice to always have this in reward events
}

type TransactionReceipt struct {
    Id			sm.TransactionID// the source transaction
    // The results of STF transaction processing
    Result		pb.TransactionReceipt_TransactionResult
    Gas_used		uint64		// gas units used by the transaction (gas price in tx)
    Fee			Amount		// transaction fee charged for the transaction
    Layer_number	sm.LayerID	// The layer in which the STF processed this transaction
}

type Transaction struct {
    TxType	pb.Transaction_TransactionType
    Id		sm.TransactionID
    Sender	sm.Address
    Fee		TransactionFee
    Timestamp	uint64		// shouldn't this be part of the event envelope?
    Receiver	sm.Address	// depending on tx type
    Amount	Amount		// amount of coin transfered in this tx by sender
    Counter	uint64		// tx counter aka nonce
    Data	[]byte		// binary payload - used for app, deploy, atx and spwan transactions
    SmesherId	SmesherID	// used in atx only
    Prev_atx	sm.TransactionID// previous ATX. used in atx.
    State	pb.TransactionState_TransactionStateType
    Result	pb.TransactionReceipt_TransactionResult
}

type Block struct {
    Id			sm.BlockID
    Transactions	[]*Transaction
}

type Layer struct {
    Index		sm.LayerID
    Status		pb.Layer_LayerStatus
    Hash		[]byte
    Blocks		[]*Block
    RootStateHash	[]byte
}

func (m *Message) ToJson() []byte {
    b, _ := json.Marshal(m)
    return b
}

func (t *TransactionReceipt) IsExecuted() bool {
    return t.Result == pb.TransactionReceipt_EXECUTED
}

func (t *Transaction) IsExecuted() bool {
    return t.Result == pb.TransactionReceipt_EXECUTED
}

func (t *Transaction) IsSimple() bool {
    return t.TxType == pb.Transaction_SIMPLE
}

func (t *Transaction) IsATX() bool {
    return t.TxType == pb.Transaction_ATX
}

func (t *Transaction) IsAPP() bool {
    return t.TxType == pb.Transaction_APP
}
