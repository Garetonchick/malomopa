package calc

import (
	"context"
	"fmt"
	"malomopa/internal/common"

	"go.uber.org/zap"
)

const (
	costCalcServiceName = "cost_calc"
)

type SimpleCostCalculator struct {
}

func MakeSimpleCostCalculator() (common.CostCalculator, error) {
	return &SimpleCostCalculator{}, nil
}

func extractOrderDetails[T any](orderInfo common.OrderInfo, key string) (*T, error) {
	orderDetailsRaw, ok := orderInfo[key]
	if !ok {
		return nil, fmt.Errorf("missing key %s in orderInfo", key)
	}

	orderDetails, ok := orderDetailsRaw.(T)
	if !ok {
		return nil, fmt.Errorf("missing key %s in orderInfo", key)
	}

	return &orderDetails, nil
}

func (sc *SimpleCostCalculator) CalculateCost(ctx context.Context, orderInfo common.OrderInfo) (float32, error) {
	logger := common.GetRequestLogger(ctx, costCalcServiceName, "calculate_cost")

	generalInfo, err := extractOrderDetails[common.GeneralOrderInfo](orderInfo, common.Keys.GeneralOrderInfo)
	if err != nil || generalInfo == nil {
		logger.Error("failed to extract general order info",
			zap.Error(err),
		)
		return 0, err
	}

	zoneInfo, err := extractOrderDetails[common.ZoneInfo](orderInfo, common.Keys.ZoneInfo)
	if err != nil || zoneInfo == nil {
		logger.Error("failed to extract zone info",
			zap.Error(err),
		)
		return 0, err
	}

	tollRoadsInfo, err := extractOrderDetails[common.TollRoadsInfo](orderInfo, common.Keys.TollRoadsInfo)
	if err != nil || tollRoadsInfo == nil {
		logger.Error("failed to extract toll roads info",
			zap.Error(err),
		)
		return 0, err
	}

	assignOrderConfigs, err := extractOrderDetails[common.AssignOrderConfigs](orderInfo, common.Keys.AssignOrderConfigs)
	if err != nil || assignOrderConfigs == nil {
		logger.Error("failed to extract assign order configs",
			zap.Error(err),
		)
		return 0, err
	}

	coinCoeff := zoneInfo.CoinCoeff

	if assignOrderConfigs.CoinCoeffCfg != nil {
		coinCoeff = min(coinCoeff, assignOrderConfigs.CoinCoeffCfg.Max)
	}

	return generalInfo.BaseCoinAmount*coinCoeff + tollRoadsInfo.BonusAmount, nil
}
