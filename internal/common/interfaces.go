package common

import "context"

type CacheServiceProvider interface {
	GetOrderInfo(ctx context.Context, orderID, executorID string) (*Order, error)
}

type DBProvider interface {
	CreateOrder(orderID, executorID string, order *Order) error
	CancelOrder(orderID string) (*Order, error)
	AcquireOrder(executorID string) (*Order, error)
}
