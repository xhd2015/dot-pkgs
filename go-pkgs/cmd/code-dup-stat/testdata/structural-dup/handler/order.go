package handler

type OrderReq struct {
	ID     int
	Amount float64
}

type OrderResp struct {
	Status string
}

func ProcessOrder(req *OrderReq) (*OrderResp, error) {
	if req.ID <= 0 {
		return nil, errInvalidID
	}
	validatedAmount := transformAmount(req.Amount)
	if validatedAmount <= 0 {
		return nil, errInvalidAmount
	}
	return &OrderResp{Status: "ok"}, nil
}
