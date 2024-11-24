package db

import (
	"context"
	"errors"
	"malomopa/internal/common"
	"malomopa/internal/config"
	"time"

	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

type dbProviderImpl struct {
	cluster *gocql.ClusterConfig
}

const (
	dbServiceName = "DB"
)

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
	logger := common.GetRequestLogger(ctx, dbServiceName, "create_order")

	session, err := p.cluster.CreateSession()
	if err != nil {
		logger.Error("failed to create cluster session",
			zap.Error(err),
		)
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

	err = session.Query(
		query,
		order.OrderID,
		order.ExecutorID,
		time.Now(),
		order.Cost,
		order.Payload,
		false,
		false,
	).WithContext(ctx).Exec()

	if err != nil {
		logger.Error("Failed to execute insert order",
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (p *dbProviderImpl) CancelOrder(ctx context.Context, orderID string) (common.OrderPayload, error) {
	logger := common.GetRequestLogger(ctx, dbServiceName, "cancel_order")

	session, err := p.cluster.CreateSession()
	if err != nil {
		logger.Error("failed to create cluster session",
			zap.Error(err),
		)
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
		logger.Error("Failed to execute update order",
			zap.Error(err),
		)
		return nil, err
	}

	if !applied {
		logger.Error("Failed to apply cas in update order",
			zap.Error(ErrNoSuchRowToUpdate),
		)
		return nil, ErrNoSuchRowToUpdate
	}

	query = newSelect().
		columns("payload").
		from(p.cluster.Keyspace, "orders").
		where("order_id = ?").build()

	var payload common.OrderPayload
	err = session.Query(query, orderID).WithContext(ctx).Scan(&payload)
	if err != nil {
		logger.Error("Failed to select order",
			zap.Error(err),
		)
		return nil, err
	}

	return payload, nil
}

func (p *dbProviderImpl) AcquireOrder(ctx context.Context, executorID string) (common.OrderPayload, error) {
	logger := common.GetRequestLogger(ctx, dbServiceName, "acquire_order")

	session, err := p.cluster.CreateSession()
	if err != nil {
		logger.Error("failed to create cluster session",
			zap.Error(err),
		)
		return nil, err
	}
	defer session.Close()

	selectQuery := newSelect().
		columns("payload", "order_id").
		from(p.cluster.Keyspace, "orders").
		where("executor_id = ? AND is_acquired = false AND is_cancelled = false").
		limit(1).
		build()

	var payload common.OrderPayload
	var orderID string
	err = session.Query(selectQuery, executorID).WithContext(ctx).Scan(&payload, &orderID)
	if err != nil {
		logger.Error("Failed to select random order", zap.Error(err))
		return nil, err
	}

	updateQuery := newUpdate(p.cluster.Keyspace, "orders").
		set("is_acquired = true").
		where("order_id = ? and executor_id = ?").
		casIf("is_cancelled = false AND is_acquired = false").
		build()

	applied, err := session.Query(updateQuery, orderID, executorID).WithContext(ctx).MapScanCAS(make(map[string]interface{}))
	if err != nil {
		logger.Error("Failed to execute conditional update on order", zap.Error(err))
		return nil, err
	}

	if !applied {
		logger.Error("Failed to acquire order due to CAS condition", zap.Error(ErrNoSuchRowToUpdate))
		return nil, ErrNoSuchRowToUpdate
	}

	return payload, nil
}
