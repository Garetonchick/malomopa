package db

import (
	"errors"
	"fmt"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"time"

	"github.com/gocql/gocql"
)

type dbProviderImpl struct {
	cluster *gocql.ClusterConfig
}

var (
	ErrDBMisconfigured   = errors.New("db misconfigured")
	ErrNoSuchRowToUpdate = errors.New("no such row to update")
)

func MakeDBProvider(cfg *config.ScyllaConfig) (common.DBProvider, error) {
	if cfg == nil {
		return nil, ErrDBMisconfigured
	}

	dbProvider := &dbProviderImpl{}

	dbProvider.cluster = gocql.NewCluster(cfg.Nodes...)
	dbProvider.cluster.Port = cfg.Port
	dbProvider.cluster.Keyspace = cfg.Keyspace
	err := dbProvider.cluster.Consistency.UnmarshalText([]byte(cfg.Consistency))
	if err != nil {
		return nil, err
	}
	dbProvider.cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: cfg.NumRetries}

	return dbProvider, nil
}

func (p *dbProviderImpl) CreateOrder(order *common.Order) error {
	session, err := p.cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	query := fmt.Sprintf(
		"INSERT INTO %s.orders (order_id, executor_id, created_at, cost, payload, is_acquired, is_cancelled) VALUES (?, ?, ?, ?, ?, ?, ?)",
		p.cluster.Keyspace,
	)
	createdAt := time.Now()
	isAcquired := false
	isCancelled := false

	return session.Query(query, order.OrderID, order.ExecutorID, createdAt, order.Cost, order.Payload, isAcquired, isCancelled).Exec()
}

func (p *dbProviderImpl) CancelOrder(orderID string) ([]byte, error) {
	session, err := p.cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	tsBound := time.Now().Add(-10 * time.Minute)
	query := fmt.Sprintf(
		`UPDATE %s.orders`+
			`SET is_cancelled = true`+
			`WHERE order_id = ?`+
			`IF is_cancelled = false AND is_acquired = false AND created_at >= ?`,
		p.cluster.Keyspace,
	)

	applied, err := session.Query(query, orderID, tsBound).ScanCAS()
	if err != nil {
		return nil, err
	}

	if !applied {
		return nil, ErrNoSuchRowToUpdate
	}

	var payload []byte
	query = fmt.Sprintf("SELECT payload FROM %s.orders WHERE order_id = ?", p.cluster.Keyspace)
	err = session.Query(query, orderID).Scan(&payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (p *dbProviderImpl) AcquireOrder(executorID string) ([]byte, error) {
	// ArtNext
	return nil, nil
}
