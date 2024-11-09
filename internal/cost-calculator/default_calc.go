package calc

import (
	"errors"
	"malomopa/internal/common"
)

var (
	ErrMissingGeneralInfo   = errors.New("missing general info in orderInfo")
	ErrMissingZoneInfo      = errors.New("missing zone info in orderInfo")
	ErrMissingTollRoadsInfo = errors.New("missing toll roads info in orderInfo")
)

type SimpleCostCalculator struct {
}

func MakeSimpleCostCalculator() (common.CostCalculator, error) {
	return &SimpleCostCalculator{}, nil
}

func extractOrderDetails[T any](orderInfo common.OrderInfo, key string, errToReturn error) (*T, error) {
	orderDetailsRaw, ok := orderInfo[key]
	if !ok {
		return nil, errToReturn
	}

	orderDetails, ok := orderDetailsRaw.(T)
	if !ok {
		return nil, errToReturn
	}

	return &orderDetails, nil
}

func (sc *SimpleCostCalculator) CalculateCost(orderInfo common.OrderInfo) (float32, error) {
	generalInfo, err := extractOrderDetails[common.GeneralOrderInfo](orderInfo, common.GeneralOrderInfoKey, ErrMissingGeneralInfo)
	if err != nil || generalInfo == nil {
		return 0, err
	}

	zoneInfo, err := extractOrderDetails[common.ZoneInfo](orderInfo, common.ZoneInfoKey, ErrMissingZoneInfo)
	if err != nil || zoneInfo == nil {
		return 0, err
	}

	tollRoadsInfo, err := extractOrderDetails[common.TollRoadsInfo](orderInfo, common.TollRoadsInfoKey, ErrMissingTollRoadsInfo)
	if err != nil || tollRoadsInfo == nil {
		return 0, err
	}

	return generalInfo.BaseCoinAmount*zoneInfo.CoinCoeff + tollRoadsInfo.BonusAmount, nil
}
