package history

import (
    "sort"
    "github.com/spacemeshos/dash-backend/types"
)

type CommitmentSizes []uint64

func (a CommitmentSizes) Len() int           { return len(a) }
func (a CommitmentSizes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitmentSizes) Less(i, j int) bool { return a[i] < a[j] }

func gini(smeshers map[types.SmesherID]*types.Smesher) float64 {
    var n uint64
    var sum float64
    data := make(CommitmentSizes, len(smeshers))
    for _, smesher := range smeshers {
        data[n] = smesher.Commitment_size
        if data[n] == 0 {
            data[n] = 1
        }
        sum += float64(data[n])
        n++
    }
    sort.Sort(data)
    var top float64
    for i, x := range data {
        top += float64((2 * (uint64(i) + 1) - n - 1) * x)
    }
    return top / (float64(n) * sum)
}

