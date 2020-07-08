package history

import (
    "sort"
    "github.com/spacemeshos/dash-backend/types"
    "github.com/spacemeshos/go-spacemesh/log"
)

type CommitmentSizes []uint64

func (a CommitmentSizes) Len() int           { return len(a) }
func (a CommitmentSizes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a CommitmentSizes) Less(i, j int) bool { return a[i] < a[j] }
/*
        1          sum((n + 1 - i)*y[i])
    G = -(n + 1 - 2---------------------
        n                sum(y[i])
*/
func gini(smeshers map[types.SmesherID]*types.Smesher) float64 {
    var n int
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
    for i, y := range data {
        top += float64(uint64(n - i) * y)
    }
    c := (float64(n) + 1.0 - 2.0 * top / sum) / float64(n)
    log.Info("gini: top = %v, n = %v, sum = %v, coef = %v", top, n, sum, c)
    return c
}

