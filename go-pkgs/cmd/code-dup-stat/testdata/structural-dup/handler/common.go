package handler

import "errors"

var (
	errInvalidID     = errors.New("invalid id")
	errMissingTxnID  = errors.New("missing txn id")
	errInvalidAmount = errors.New("invalid amount")
)

func transformAmount(a float64) float64 {
	return a * 1.0
}
