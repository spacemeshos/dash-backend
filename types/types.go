package types

import (
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

const (
    AddressLength = 20
    Hash32Length = 32
    hash20Length = 20
)

type Amount uint64
type LayerID uint32
type Address [AddressLength]byte
type Hash32 [Hash32Length]byte
type Hash20 [hash20Length]byte

type ActivationID  Hash32
type TransactionID Hash32
type BlockID Hash20
type SmesherID [32]byte

type Geo struct {
    Name	string `json:"name"`
    Coordinates	[2]float64 `json:"coordinates"`
}

type NetworkInfo struct {
    NetId			uint64
    GenesisTime			uint64
    EpochNumLayers		uint64
    MaxTransactionsPerSecond	uint64
    LayerDuration		uint64
}

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
    Address	Address	// account public address
    Counter	uint64	// aka nonce
    Balance	Amount	// known account balance
}

type Reward struct {
    Layer		LayerID
    Total		Amount
    Layer_reward	Amount
    Layer_computed	LayerID	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    Coinbase		Address	// account awarded this reward
    Smesher		SmesherID	// it will be nice to always have this in reward events
}

type GasOffered struct {
    Gas_provided	uint64
    Gas_price		uint64
}

type TransactionReceipt struct {
    Id			TransactionID	// the source transaction
    // The results of STF transaction processing
    Result		pb.TransactionReceipt_TransactionResult
    Gas_used		uint64		// gas units used by the transaction (gas price in tx)
    Fee			Amount		// transaction fee charged for the transaction
    Layer_number	LayerID		// The layer in which the STF processed this transaction
    Index		uint32		// the index of the tx in the ordered list of txs to be executed by stf in the layer
    SvmData		[]byte		// svm binary data. Decode with svm-codec
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
    Layer	LayerID	// the layer that this activation is part of
    Smesher_id	SmesherID	// id of smesher who created the ATX
    Coinbase	Address	// coinbase account id
    Prev_atx	ActivationID	// previous ATX pointed to
    Commitment_size	uint64	// commitment size in bytes
}

type Transaction interface {
    GetID() *TransactionID
    GetResult() pb.TransactionReceipt_TransactionResult
    SetResult(result pb.TransactionReceipt_TransactionResult)
    Print()
}

type TransactionBase struct {
    Id		TransactionID

    Sender	Address	// tx originator, should match signer inside Signature
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

    Receiver	Address
}

// Data specific to a smart contract transaction.
type SmartContractTransaction struct {
    TransactionBase

    Type	pb.SmartContractTransaction_TransactionType
    Data	[]byte		// packed binary arguments, including ABI selector
    Address	Address	// address of smart contract or template
}

type Block struct {
    Id			BlockID
    Transactions	[]Transaction
}

type Layer struct {
    Number		LayerID	// layer number - not hash - layer content may change
    Status		pb.Layer_LayerStatus
    Hash		[]byte		// computer layer hash - do we need this?
    Blocks		[]*Block	// layer's blocks
    Activations		[]*Activation	// list of layer's activations
    RootStateHash	[]byte		// when available - the root state hash of global state in this layer
    Transactions	map[TransactionID]Transaction
}

func (tx *TransactionBase) GetID() *TransactionID {
    return &tx.Id
}

func (tx *TransactionBase) GetResult() pb.TransactionReceipt_TransactionResult {
    return tx.Result
}

func (tx *TransactionBase) SetResult(result pb.TransactionReceipt_TransactionResult) {
    tx.Result = result
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
    var a Address
    a.SetBytes(b)
    return a
}

// Bytes gets the string representation of the underlying address.
func (a Address) Bytes() []byte { return a[:] }

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
    if len(b) > len(a) {
        b = b[len(b)-AddressLength:]
    }
    copy(a[AddressLength-len(b):], b)
}
