package mock

var config Config = Config {
    baseCommitmentSize: 4 * 1024 * 1024 * 1024,
    commitmentSizeRate: []Rate { {1, 14}, {2, 3}, {3, 2}, {4, 1} },
    accountsCountRate: []Rate { {1, 14}, {2, 3}, {3, 2}, {4, 1} },
    smeshersGrowsRate: []Rate { {0, 14}, {1, 3}, {2, 2}, {3, 1} },
    newAccountsRate: 10,
    layerReward: 50*1000*1000000000,
}

type Rate struct {
    Value int
    Rate int
}

type Config struct {
    baseCommitmentSize uint64
    commitmentSizeRate []Rate
    accountsCountRate []Rate
    smeshersGrowsRate []Rate
    newAccountsRate int
    layerReward uint64
}
