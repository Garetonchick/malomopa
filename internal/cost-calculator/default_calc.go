package calc

import (
	"fmt"
	"malomopa/internal/common"
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

func (sc *SimpleCostCalculator) CalculateCost(orderInfo common.OrderInfo) (float32, error) {
	generalInfo, err := extractOrderDetails[common.GeneralOrderInfo](orderInfo, common.GeneralOrderInfoKey)
	if err != nil || generalInfo == nil {
		return 0, err
	}

	zoneInfo, err := extractOrderDetails[common.ZoneInfo](orderInfo, common.ZoneInfoKey)
	if err != nil || zoneInfo == nil {
		return 0, err
	}

	tollRoadsInfo, err := extractOrderDetails[common.TollRoadsInfo](orderInfo, common.TollRoadsInfoKey)
	if err != nil || tollRoadsInfo == nil {
		return 0, err
	}

	return generalInfo.BaseCoinAmount*zoneInfo.CoinCoeff + tollRoadsInfo.BonusAmount, nil
}
