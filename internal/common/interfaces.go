package common

import "context"

type CacheServiceProvider interface {
	GetOrderInfo(ctx context.Context, orderID, executorID string) (OrderInfo, error)
}

type DBProvider interface {
	CreateOrder(order *Order) error
	CancelOrder(orderID string) (OrderPayload, error)
	AcquireOrder(executorID string) (OrderPayload, error)
}

type CostCalculator interface {
	CalculateCost(orderInfo OrderInfo) (float32, error)
}
