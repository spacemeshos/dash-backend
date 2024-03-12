package utils

import (
	"math"
	"strconv"
)

type CoinUnits struct {
	SMH    string
	Smidge string
}

var coinUnits = CoinUnits{
	SMH:    "SMH",
	Smidge: "Smidge",
}

type Coin struct {
	Value string
	Unit  string
}

func packValueAndUnit(value float64, unit string) Coin {
	return Coin{
		Value: strconv.FormatFloat(value, 'f', 3, 64),
		Unit:  unit,
	}
}

func ParseSmidge(amount float64) Coin {
	if amount <= 0 || math.IsNaN(amount) {
		return packValueAndUnit(0, coinUnits.SMH)
	}
	if amount >= math.Pow10(6) {
		return packValueAndUnit(toSMH(amount), coinUnits.SMH)
	}
	return packValueAndUnit(amount, coinUnits.Smidge)
}

func toSMH(smidge float64) float64 {
	return smidge / math.Pow10(9)
}
