package mock

import (
//    "math/big"
    "math/rand"
    "time"

    "github.com/spacemeshos/ed25519"
    sm "github.com/spacemeshos/go-spacemesh/common/types"
    "github.com/spacemeshos/go-spacemesh/log"
    pb "github.com/spacemeshos/api/release/go/spacemesh/v1"
)

func (s *Smesher) GetAccount() *Account {
    return &s.Accounts[rand.Intn(len(s.Accounts))]
} 

func NewSmesher() *Smesher {
    smesher := &Smesher {
        Activations: make([]*Activation, 0),
        Accounts: make([]Account, 0),
        CommitmentSize: config.baseCommitmentSize * uint64(getRate(config.commitmentSizeRate)),
    }
    newSmesherID(&smesher.Id)
    for i := 0; i < getRate(config.accountsCountRate[:]); i++ {
        publicKey, _, _ := ed25519.GenerateKey(nil)
        smesher.Accounts = append(smesher.Accounts, Account{Address: sm.BytesToAddress(publicKey)})
    }
    return smesher
}

func (n *Network) executeTransaction(layer *Layer, tx *Transaction, index uint32) {
    log.Info("Execute Tx %v", tx.Id)

    var gasUsed uint64
    result := pb.TransactionReceipt_TRANSACTION_RESULT_EXECUTED
    tx.State = pb.TransactionState_TRANSACTION_STATE_PROCESSED

    if tx.IsSimple {
        gasUsed = 21000;
        if gasUsed > tx.GasOffered.GasProvided {
            result = pb.TransactionReceipt_TRANSACTION_RESULT_INSUFFICIENT_GAS
            tx.State = pb.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS
            gasUsed = tx.GasOffered.GasProvided
        } else {
            from, ok := n.AccountsWithBalance[tx.Sender]
            if ok {
                if from.Balance >= tx.Amount {
                    to, ok := n.AccountsByAddress[tx.Receiver]
                    if !ok {
                        to = &Account{Address: tx.Receiver}
                        n.addAccount(to)
                    }
                    n.changeBalance(from, -tx.Amount)
                    n.changeBalance(to, tx.Amount)
                } else {
                    result = pb.TransactionReceipt_TRANSACTION_RESULT_INSUFFICIENT_FUNDS
                    tx.State = pb.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS
                }
            } else {
                result = pb.TransactionReceipt_TRANSACTION_RESULT_INSUFFICIENT_FUNDS
                tx.State = pb.TransactionState_TRANSACTION_STATE_INSUFFICIENT_FUNDS
            }
        }
    } else {
        tx.State = pb.TransactionState_TRANSACTION_STATE_REJECTED
        result = pb.TransactionReceipt_TRANSACTION_RESULT_RUNTIME_EXCEPTION
        gasUsed = 8000
    }
    tx.Receipt = &TransactionReceipt{
        Id: tx.Id,
        Result: result,
        GasUsed: gasUsed,
        Fee: Amount(gasUsed * tx.GasOffered.GasPrice),
        LayerNumber: layer.Number,
        Index: index,
    }
    log.Info("Tx execure result: %v", tx.Receipt.Result)
}

func NewCoinTransferTransaction(from *Account, to *Account) *Transaction {
    return nil
}

func NewBlock(smesher *Smesher, layer *Layer) (*Block, *Activation) {
    block := &Block {
        Transactions: make([]*Transaction, 0),
    }
    newBlockID(&block.Id)

    atx := &Activation {
        Layer: layer.Number,
        SmesherId: smesher.Id,
        Coinbase: smesher.GetAccount().Address,
        CommitmentSize: smesher.CommitmentSize,
    }
    if len(smesher.Activations) > 0 {
        atx.PrevAtx = smesher.Activations[len(smesher.Activations) - 1].Id
    }
    newActivationID(&atx.Id)
    smesher.Activations = append(smesher.Activations, atx)

    return block, atx
}

func (n *Network) playTransactions(layer *Layer) {
    var index uint32
    for _, tx := range layer.Transactions {
        n.executeTransaction(layer, tx, index)
        index++
    }
}

func (n *Network) approve(layer *Layer) {
    if layer.Status != pb.Layer_LAYER_STATUS_APPROVED {
        n.playTransactions(layer)
        layer.Status = pb.Layer_LAYER_STATUS_APPROVED
        n.Listener.OnLayerChanged(layer.Export())
    }
}

func (n *Network) addAccount(account *Account) {
    account.Active = true
    n.Accounts = append(n.Accounts, account)
    n.AccountsByAddress[account.Address] = account
    n.Listener.OnAccountChanged(account.Export())
    log.Info("New account %v", account)
}

func (n *Network) changeBalance(account *Account, amount Amount) {
    if account.Balance + amount < 0 {
        log.Info("Wrong change balance for %v, balance %v, amount %v", account.Address, account.Balance, amount)
    } else {
        if amount != 0 {
            before := account.Balance
            account.Balance += amount
            if account.Balance > 0 {
                n.AccountsWithBalance[account.Address] = account
            } else {
                delete(n.AccountsWithBalance, account.Address)
            }
            n.Listener.OnAccountChanged(account.Export())
            log.Info("Change account %v balance: %v -> %v", account.Address, before, account.Balance)
        }
    }
}

func (n *Network) applyReward(reward *Reward) {
    account, ok := n.AccountsByAddress[reward.Coinbase]
    if ok {
        n.changeBalance(account, reward.Total)
        n.Listener.OnReward(reward.Export())
    }
    log.Info("%v, %v, %v", reward.Coinbase, reward.Total, reward.LayerReward)
}

func (n *Network) confirm(layer *Layer) {
    if layer.Status == pb.Layer_LAYER_STATUS_CONFIRMED {
        return
    }
    layer.Status = pb.Layer_LAYER_STATUS_CONFIRMED
    n.Listener.OnLayerChanged(layer.Export())

    current := n.GetCurrentLayerNumber()

    if len(layer.Activations) > 0 {
        layerReward := config.layerReward
        txsFee := layer.GetTxsFee()
        totalReward := layerReward + txsFee

        var total uint64
        for _, atx := range layer.Activations {
            total += atx.CommitmentSize
        }

        units := total / config.baseCommitmentSize
        oneTotalReward := totalReward / units
        oneLayerReward := layerReward / units
        for _, atx := range layer.Activations {
            weight := atx.CommitmentSize / config.baseCommitmentSize
            reward := &Reward {
                Layer: layer.Number,
                Total: Amount(weight * oneTotalReward),
                LayerReward: Amount(weight * oneLayerReward),
                LayerComputed: sm.LayerID(current),
                Coinbase: atx.Coinbase,
                Smesher: atx.SmesherId,
            }
            layer.Rewards = append(layer.Rewards, reward)
            n.applyReward(reward)
        }
    }
}

func (layer *Layer) GetTxsFee() uint64 {
    var fee uint64
    for _, tx := range layer.Transactions {
        if tx.Receipt != nil {
            fee += uint64(tx.Receipt.Fee)
        }
    }
    return fee
}

func (n *Network) hare(layer *Layer) *Layer {
    n.approve(layer)
    return layer
}

func (n *Network) tortoise(layer *Layer, approved *Layer) bool {
    if layer.Status == pb.Layer_LAYER_STATUS_CONFIRMED {
        return true
    }
    if approved.Number - layer.Number >= 3 {
        n.confirm(layer)
        return true
    }
    return false
}

func (n *Network) newEpoch() *Epoch {
    number := uint64(len(n.Epochs))

    log.Info("New epoch %v", number)

    epoch := &Epoch {
        number: number,
        begin: sm.LayerID(number * n.EpochNumLayers),
        smeshers: make(map[SmesherID]*Smesher),
        accounts: make(map[sm.Address]*Account),
    }
    n.Epochs = append(n.Epochs, epoch)

    return epoch
}

func (n *Network) appendSmesher(smesher *Smesher) {
    n.Smeshers = append(n.Smeshers, smesher)
    n.SmeshersById[smesher.Id] = smesher
    log.Info("Append smesher with ID %v", smesher.Id)
    for _, account := range smesher.Accounts {
        n.addAccount(&account)
    }
}

func (n *Network) updateSmeshers() {
    if len(n.Smeshers) == 0 {
        n.appendSmesher(NewSmesher())
    } else {
        count := getRate(config.smeshersGrowsRate[:])
        for i := 0; i < count; i++ {
            n.appendSmesher(NewSmesher())
        }
    }
}

func getRandomSmesher(smeshers map[SmesherID]*Smesher) *Smesher {
    length := len(smeshers)
    if length > 0 {
        index := rand.Intn(length)
        for _, smesher := range smeshers {
            if index == 0 {
                return smesher
            }
            index--
        }
    }
    return nil
}

func (n *Network) getSmeshersForLayer() []*Smesher {
    smeshersPerLayer := len(n.Smeshers) / int(n.EpochNumLayers)
    if smeshersPerLayer < 50 {
        smeshersPerLayer = 1 + rand.Intn(49)
    }
    if smeshersPerLayer > len(n.Smeshers) {
        smeshersPerLayer = len(n.Smeshers)
    }
    if len(n.smeshersForEpoch) == 0 {
        for _, smesher := range n.Smeshers {
            n.smeshersForEpoch[smesher.Id] = smesher
        }
log.Info("smeshersForEpoch: %v", len(n.smeshersForEpoch))
    }
log.Info("smeshersPerLayer: %v", smeshersPerLayer)
    smeshers := make([]*Smesher, 0, smeshersPerLayer)
    for ; smeshersPerLayer > 0; {
        smesher := getRandomSmesher(n.smeshersForEpoch)
        if smesher != nil {
            smeshers = append(smeshers, smesher)
            delete(n.smeshersForEpoch, smesher.Id)
        }
        smeshersPerLayer--
    }
    return smeshers
}

func (n *Network) createMockBlocks(layer *Layer) {
    n.updateSmeshers()
    smeshers := n.getSmeshersForLayer()
    for _, smesher := range smeshers {
        block, atx := NewBlock(smesher, layer)
        layer.Blocks = append(layer.Blocks, block)
        layer.Activations = append(layer.Activations, atx)
    }
    log.Info("New blocks: %v", len(layer.Blocks))
}

func getRandomAccount(accounts map[sm.Address]*Account) *Account {
    length := len(accounts)
    if length > 0 {
        index := rand.Intn(length)
        for _, account := range accounts {
            if index == 0 {
                return account
            }
            index--
        }
    }
    return nil
}

func (n *Network) createMockTransactions(layer *Layer) {
    var transfers int
    var calls int
    if len(n.AccountsWithBalance) > 0 {
        transfers = rand.Intn(int(n.MaxTransactionsPerSecond))
    }
    for ; transfers > 0; {
        from := getRandomAccount(n.AccountsWithBalance)
        amount := from.Balance / Amount(1 + rand.Intn(10))
        var to sm.Address
        if rand.Intn(config.newAccountsRate) > 0 {
            to = n.Accounts[rand.Intn(len(n.Accounts))].Address
        } else {
            publicKey, _, _ := ed25519.GenerateKey(nil)
            to = sm.BytesToAddress(publicKey)
        }
        tx := &Transaction{
            Sender: from.Address,
            GasOffered: GasOffered{6000000, 5000000},
            Amount: amount,
            Counter: from.Counter,
            State: pb.TransactionState_TRANSACTION_STATE_MESH,
            IsSimple: true,
            Receiver: to,
        }
        newTransactionID(&tx.Id)
        from.Counter++
        block := layer.Blocks[rand.Intn(len(layer.Blocks))]
        block.Transactions = append(block.Transactions, tx)
        layer.Transactions[tx.Id] = tx
        transfers--

        log.Info("New transaction:")
        log.Info("id      %v", tx.Id)
        log.Info("from   %v", from.Address)
        log.Info("to     %v", to)
        log.Info("amount %v", amount)
    }
    if calls > 0 {
    }
}

func (n *Network) newLayer() *Layer {
    number := sm.LayerID(len(n.Layers))

    log.Info("New layer %v", number)

    layer := &Layer {
        Number: number,
        Status: pb.Layer_LAYER_STATUS_UNSPECIFIED,
        Hash: make([]byte, 0),
        Blocks: make([]*Block, 0),
        Activations: make([]*Activation, 0),
        RootStateHash: make([]byte, 0),
        Transactions: make(map[sm.TransactionID]*Transaction),
    }
    if number > 0 {
        approved := n.hare(n.Layers[number - 1])
        for i := int(number) - 2; i >= 0; i-- {
            if n.tortoise(n.Layers[i], approved) {
                break
            }
        }
    }
    n.Layers = append(n.Layers, layer)

    epochNumber := int(uint64(number) / n.EpochNumLayers)
    if len(n.Epochs) < (epochNumber + 1) {
        n.newEpoch()
    }

    n.createMockBlocks(layer);

    return layer
}

func (n *Network) getCurrentEpoch() *Epoch {
    return n.Epochs[len(n.Epochs) - 1]
}

func (n *Network) getCurrentLayer() *Layer {
    return n.Layers[len(n.Layers) - 1]
}

func (n *Network) GetCurrentEpochNumber() uint64 {
    return uint64(len(n.Epochs))
}

func (n *Network) GetCurrentLayerNumber() uint64 {
    return uint64(len(n.Layers))
}

func NewNetwork(listener Listener, netId int, epochLayers int, maxTxs int, layerDuration int) *Network {
    network := &Network {
        GenesisTime: uint64(time.Now().UTC().Unix()),
        Smeshers: make([]*Smesher, 0),
        SmeshersById: make(map[SmesherID]*Smesher),
        smeshersForEpoch: make(map[SmesherID]*Smesher),
        Accounts: make([]*Account, 0),
        AccountsByAddress: make(map[sm.Address]*Account),
        AccountsWithBalance: make(map[sm.Address]*Account),
        Epochs: make([]*Epoch, 0),
        Layers: make([]*Layer, 0),
        Listener: listener,
    }
    network.NetId = uint64(netId)
    network.EpochNumLayers = uint64(epochLayers)
    network.MaxTransactionsPerSecond = uint64(maxTxs)
    network.LayerDuration = uint64(layerDuration)

    log.Info("Created network ID %v, epoch layers %v, max TxS %v and layer duration %v s", netId, epochLayers, maxTxs, layerDuration)

    network.newLayer()

    return network
}

func (n *Network) Tick() {
    log.Info("Network tick")
    currentLayerNumber := sm.LayerID((uint64(time.Now().UTC().Unix()) - n.GenesisTime) / n.LayerDuration)
    layer := n.getCurrentLayer()
    if layer.Number < currentLayerNumber {
        layer = n.newLayer()
    }
    n.createMockTransactions(layer)
}
