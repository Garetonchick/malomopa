package common

import "context"

type CacheServiceProvider interface {
	GetOrderInfo(ctx context.Context, orderID, executorID string) (OrderInfo, error)
}

type DBProvider interface {
	CreateOrder(ctx context.Context, order *Order) error
	CancelOrder(ctx context.Context, orderID string) (OrderPayload, error)
	AcquireOrder(ctx context.Context, executorID string) (OrderPayload, error)
}

type CostCalculator interface {
	CalculateCost(orderInfo OrderInfo) (float32, error)
}
