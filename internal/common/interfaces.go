package common

type CacheServiceProvider interface {
	GetOrderInfo(orderID string) (*Order, error)
}

type DBProvider interface {
	CreateOrder(order *Order) error
	CancelOrder(orderID string) (*Order, error)
	AcquireOrder(executorID string) (*Order, error)
}
