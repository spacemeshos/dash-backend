package history

import (
    "sync"

    sm "github.com/spacemeshos/go-spacemesh/common/types"
    "github.com/spacemeshos/dash-backend/client"
    "github.com/spacemeshos/dash-backend/types"
)

type Statistics struct {
    capacity		uint64	// Average tx/s rate over capacity considering all layers in the current epoch.
    decentral		uint64	// Distribution of storage between all active smeshers.
    smeshers		uint64	// Number of active smeshers in the current epoch.
    transactions	uint64	// Total number of transactions processed by the state transition function.
    accounts		uint64	// Total number of on-mesh accounts with a non-zero coin balance as of the current epoch.
    circulation		uint64	// Total number of Smesh coins in circulation. This is the total balances of all on-mesh accounts.
    rewards		uint64	// Total amount of Smesh minted as mining rewards as of the last known reward distribution event.
    security		uint64	// Total amount of storage committed to the network based on the ATXs in the previous epoch.
}
/*
Decentralization
Main display: degree_of_decentralization for the current layer.

    degree_of_decentralization is defined as: 0.5 * (min(n,1e4)^2/1e8) + 0.5 * (1 - gini_coeff(last_100_epochs))
    0.5 is an arbitrary weight assigned to each metric and can be adjusted
    n is the number of unique smeshers who contributed blocks over the last 100 epochs. (I’ve arbitrarily set a ceiling of 10k miners as “full decentralization”--this can be changed. The function is quadratic and should approach 1 as n approaches 10k.)
    gini_coeff is the GINI coefficient of the vector of total blocks (or block weights) contributed by those n smeshers over the last 100 layers. (We want one minus the GINI coefficient because higher is more unequal, i.e., less decentralized.)

def gini(array):
    """Calculate the Gini coefficient of a numpy array."""
    # based on bottom eq: http://www.statsdirect.com/help/content/image/stat0206_wmf.gif
    # from: http://www.statsdirect.com/help/default.htm#nonparametric_methods/gini.htm
    array = array.flatten() #all values are treated equally, arrays must be 1d
    if np.amin(array) < 0:
        array -= np.amin(array) #values cannot be negative
    array += 0.0000001 #values cannot be 0
    array = np.sort(array) #values must be sorted
    index = np.arange(1,array.shape[0]+1) #index per array element
    n = array.shape[0]#number of array elements
    return ((np.sum((2 * index - n  - 1) * array)) / (n * np.sum(array))) #Gini coefficient
*/
type Stats struct {
    current	Statistics
    cumulative	Statistics
}

type Epoch struct {
    history	*History
    prev	*Epoch
    number	uint64
    ended	bool
    smeshers	map[types.SmesherID]*types.Smesher

    layers 		map[sm.LayerID]*types.Layer
    lastLayer		*types.Layer
    lastApprovedLayer	*types.Layer
    lastConfirmedLayer	*types.Layer

    stats	Stats
}

type History struct {
    bus 	*client.Bus
//    storage	*storage

    network	types.NetworkInfo

    smeshers	map[types.SmesherID]*types.Smesher

    epoch 	*Epoch
    accounts	map[sm.Address]*types.Account
    epochs	map[uint64]*Epoch
    mux 	sync.Mutex

    total	Stats
}
