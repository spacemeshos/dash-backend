package types

import (
    "encoding/json"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
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

type NetworkInfo struct {
    NetId			uint64
    GenesisTime			uint64
    EpochNumLayers		uint64
    MaxTransactionsPerSecond	uint64
    LayerDuration		uint64
}

type Amount uint64

type ActivationID sm.ATXID

type SmesherID [32]byte

type Smesher struct {
    Id			SmesherID
    Geo			Geo
    Activations		[]*Activation
    Commitment_size	uint64	// commitment size in bytes
}

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

type GasOffered struct {
    Gas_provided	uint64
    Gas_price		uint64
}

type TransactionReceipt struct {
    Id			sm.TransactionID// the source transaction
    // The results of STF transaction processing
    Result		pb.TransactionReceipt_TransactionResult
    Gas_used		uint64		// gas units used by the transaction (gas price in tx)
    Fee			Amount		// transaction fee charged for the transaction
    Layer_number	sm.LayerID	// The layer in which the STF processed this transaction
    Index		uint32		// the index of the tx in the ordered list of txs to be executed by stf in the layer
    App_address		sm.Address	// deployed app address or code template address
}

// A simple signature data
type Signature struct {
    Scheme	pb.Signature_Scheme	// the signature's scheme
    Signature	[]byte			// the signature itself
    Public_key	[]byte			// included in schemes which require signer to provide a public key
}

// An Activation "transaction" (ATX)
type Activation struct {
    Id		ActivationID
    Layer	sm.LayerID	// the layer that this activation is part of
    Smesher_id	SmesherID	// id of smesher who created the ATX
    Coinbase	sm.Address	// coinbase account id
    Prev_atx	ActivationID	// previous ATX pointed to
    Commitment_size	uint64	// commitment size in bytes
}

type Transaction interface {
    GetID() *sm.TransactionID
    GetResult() pb.TransactionReceipt_TransactionResult
    SetResult(result pb.TransactionReceipt_TransactionResult)
    Print()
}

type TransactionBase struct {
    Id		sm.TransactionID

    Sender	sm.Address	// tx originator, should match signer inside Signature
    Gas_offered	GasOffered	// gas price and max gas offered
    Amount	Amount		// amount of coin transfered in this tx by sender
    Counter	uint64		// tx counter aka nonce
    Signature	Signature	// sender signature on transaction

//    State	pb.TransactionState_TransactionState
    Result	pb.TransactionReceipt_TransactionResult
}

// Data specific to a simple coin transaction.
type CoinTransferTransaction struct {
    TransactionBase

    Receiver	sm.Address
}

// Data specific to a smart contract transaction.
type SmartContractTransaction struct {
    TransactionBase

    Type	pb.SmartContractTransaction_TransactionType
    Data	[]byte		// packed binary arguments, including ABI selector
    Address	sm.Address	// address of smart contract or template
}

type Block struct {
    Id			sm.BlockID
    Transactions	[]Transaction
}

type Layer struct {
    Number		sm.LayerID	// layer number - not hash - layer content may change
    Status		pb.Layer_LayerStatus
    Hash		[]byte		// computer layer hash - do we need this?
    Blocks		[]*Block	// layer's blocks
    Activations		[]*Activation	// list of layer's activations
    RootStateHash	[]byte		// when available - the root state hash of global state in this layer
    Transactions	map[sm.TransactionID]Transaction
}

func (m *Message) ToJson() []byte {
    b, _ := json.Marshal(m)
    return b
}

func (tx *TransactionBase) GetID() *sm.TransactionID {
    return &tx.Id
}

func (tx *TransactionBase) GetResult() pb.TransactionReceipt_TransactionResult {
    return tx.Result
}

func (tx *TransactionBase) SetResult(result pb.TransactionReceipt_TransactionResult) {
    tx.Result = result
}
