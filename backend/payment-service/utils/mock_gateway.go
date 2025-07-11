package utils

import (
	"fmt"
	"math/rand"
	"time"
)

type SimulatePaymentGatewayRequest struct {
	Amount        float64
	CardNumber    string
	ExpiryDate    string
	CVV           string
	PaymentMethod string
}

type SimulatePaymentGatewayResponse struct {
	TransactionID string
	Status        string
	Message       string
	GatewayFee    float64
}

func SimulatePaymentGateway(req SimulatePaymentGatewayRequest) SimulatePaymentGatewayResponse {
	LogInfo("Simulating payment gateway request for amount: %.2f via %s", req.Amount, req.PaymentMethod)

	time.Sleep(1 * time.Second)

	if req.Amount <= 0 {
		return SimulatePaymentGatewayResponse{
			Status:  "failed",
			Message: "Invalid payment amount.",
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	transactionID := fmt.Sprintf("TXN-%d-%d", time.Now().Unix(), r.Intn(100000))
	gatewayFee := req.Amount * 0.02

	return SimulatePaymentGatewayResponse{
		TransactionID: transactionID,
		Status:        "success",
		Message:       "Payment processed successfully.",
		GatewayFee:    gatewayFee,
	}
}
