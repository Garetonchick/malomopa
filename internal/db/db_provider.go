package db

import (
	"context"
	"errors"
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

func (p *dbProviderImpl) CreateOrder(ctx context.Context, order *common.Order) error {
	session, err := p.cluster.CreateSession()
	if err != nil {
		return err
	}
	defer session.Close()

	query := newInsert(p.cluster.Keyspace, "orders").columns(
		"order_id",
		"executor_id",
		"created_at",
		"cost",
		"payload",
		"is_acquired",
		"is_cancelled",
	).build()

	return session.Query(
		query,
		order.OrderID,
		order.ExecutorID,
		time.Now(),
		order.Cost,
		order.Payload,
		false,
		false,
	).WithContext(ctx).Exec()
}

func (p *dbProviderImpl) CancelOrder(ctx context.Context, orderID string) (common.OrderPayload, error) {
	session, err := p.cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	query := newUpdate(p.cluster.Keyspace, "orders").
		set("is_cancelled = true").
		where("order_id = ?").
		casIf("is_cancelled = false AND is_acquired = false AND created_at >= ?").build()

	applied, err := session.Query(
		query,
		orderID,
		time.Now().UTC().Add(-10*time.Minute),
	).WithContext(ctx).MapScanCAS(make(map[string]interface{}))
	if err != nil {
		return nil, err
	}

	if !applied {
		return nil, ErrNoSuchRowToUpdate
	}

	query = newSelect().
		columns("payload").
		from(p.cluster.Keyspace, "orders").
		where("order_id = ?").build()

	var payload common.OrderPayload
	err = session.Query(query, orderID).WithContext(ctx).Scan(&payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (p *dbProviderImpl) AcquireOrder(ctx context.Context, executorID string) (common.OrderPayload, error) {
	// ArtNext
	return nil, nil
}
