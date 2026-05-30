package handler

type PaymentReq struct {
	TxnID  string
	Amount float64
}

type PaymentResp struct {
	State string
}

func ProcessPayment(req *PaymentReq) (*PaymentResp, error) {
	if req.TxnID == "" {
		return nil, errMissingTxnID
	}
	validatedAmount := transformAmount(req.Amount)
	if validatedAmount <= 0 {
		return nil, errInvalidAmount
	}
	return &PaymentResp{State: "done"}, nil
}
