package common

import "context"

type CacheServiceProvider interface {
	GetOrderInfo(ctx context.Context, orderID, executorID string) (map[string]any, error)
}

type DBProvider interface {
	CreateOrder(order *Order) error
	CancelOrder(orderID string) ([]byte, error)
	AcquireOrder(executorID string) ([]byte, error)
}

type CostCalculator interface {
	CalculateCost(orderInfo map[string]any) (float32, error)
}
