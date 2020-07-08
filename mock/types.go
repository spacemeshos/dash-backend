package mock

import (
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

type Listener interface {
    SetNetworkInfo(netId uint64, genesisTime uint64, epochNumLayers uint64, maxTransactionsPerSecond uint64, layerDuration uint64)
    OnLayerChanged(layer *pb.Layer)
    OnAccountChanged(account *pb.Account)
    OnReward(reward *pb.Reward)
    OnTransactionReceipt(receipt *pb.TransactionReceipt)
}

type Amount uint64

type ActivationID sm.ATXID

type SmesherID [32]byte

type Account struct {
    Address	sm.Address	// account public address
    Counter	uint64		// aka nonce
    Balance	Amount		// known account balance
    Active	bool
}

type Geo struct {
    Name	string
    Coordinates	[2]float64
}

type Smesher struct {
    Id			SmesherID
    Geo			Geo
    Activations		[]*Activation
    CommitmentSize	uint64	// commitment size in bytes
    Accounts		[]Account
    Active		bool
}

type TransactionFee struct {
    GasConsumed	uint64
    GasPrice		uint64
    // tx_fee = gas_consumed * gas_price
}

type Reward struct {
    Layer		sm.LayerID
    Total		Amount
    LayerReward		Amount
    LayerComputed	sm.LayerID	// layer number of the layer when reward was computed
    // tx_fee = total - layer_reward
    Coinbase		sm.Address	// account awarded this reward
    Smesher		SmesherID	// it will be nice to always have this in reward events
}

type Activation struct {
    Id		ActivationID
    Layer	sm.LayerID	// the layer that this activation is part of
    SmesherId	SmesherID	// id of smesher who created the ATX
    Coinbase	sm.Address	// coinbase account id
    PrevAtx	ActivationID	// previous ATX pointed to
    CommitmentSize	uint64	// commitment size in bytes
}

type TransactionReceipt struct {
    Id		sm.TransactionID// the source transaction
    // The results of STF transaction processing
    Result	pb.TransactionReceipt_TransactionResult
    GasUsed	uint64		// gas units used by the transaction (gas price in tx)
    Fee		Amount		// transaction fee charged for the transaction
    LayerNumber	sm.LayerID	// The layer in which the STF processed this transaction
    Index	uint32		// the index of the tx in the ordered list of txs to be executed by stf in the layer
    AppAddress	sm.Address	// deployed app address or code template address
}

type GasOffered struct {
    GasProvided		uint64
    GasPrice		uint64
}

type Signature struct {
    Scheme	pb.Signature_Scheme	// the signature's scheme
    Signature	[]byte			// the signature itself
    PublicKey	[]byte			// included in schemes which require signer to provide a public key
}

type Transaction struct {
    Id		sm.TransactionID

    Sender	sm.Address	// tx originator, should match signer inside Signature
    GasOffered	GasOffered	// gas price and max gas offered
    Amount	Amount		// amount of coin transfered in this tx by sender
    Counter	uint64		// tx counter aka nonce
    Signature	Signature	// sender signature on transaction

    State	pb.TransactionState_TransactionState
    Receipt	*TransactionReceipt

    IsSimple	bool

    // CoinTransferTransaction
    Receiver	sm.Address

    // SmartContractTransaction
    Type	pb.SmartContractTransaction_TransactionType
    Data	[]byte		// packed binary arguments, including ABI selector
    Address	sm.Address	// address of smart contract or template
}

type Block struct {
    Id			sm.BlockID
    Transactions	[]*Transaction
}

type Layer struct {
    Number		sm.LayerID
    Status		pb.Layer_LayerStatus
    Hash		[]byte		// computer layer hash - do we need this?
    Blocks		[]*Block	// layer's blocks
    Activations		[]*Activation	// list of layer's activations
    RootStateHash	[]byte		// when available - the root state hash of global state in this layer
    Transactions	map[sm.TransactionID]*Transaction
    Rewards		[]*Reward
}

type Epoch struct {
    number	uint64
    begin	sm.LayerID
    confirmed	bool
    smeshers	map[SmesherID]*Smesher
    accounts	map[sm.Address]*Account
}

type Network struct {
    NetId			uint64
    GenesisTime			uint64
    EpochNumLayers		uint64
    MaxTransactionsPerSecond	uint64
    LayerDuration		uint64

    Smeshers		[]*Smesher
    SmeshersById	map[SmesherID]*Smesher
    smeshersForEpoch	map[SmesherID]*Smesher

    Accounts		[]*Account
    AccountsByAddress	map[sm.Address]*Account
    AccountsWithBalance	map[sm.Address]*Account

    Epochs	[]*Epoch
    Layers 	[]*Layer

    Listener	Listener
}
