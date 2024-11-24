package test

import "malomopa/internal/common"

type OrderPayload struct {
	Configs          common.AssignOrderConfigs `json:"configs"`
	ExecutorProfile  common.ExecutorProfile    `json:"executor_profile"`
	GeneralOrderInfo common.GeneralOrderInfo   `json:"general_order_info"`
	TollRoadsInfo    common.TollRoadsInfo      `json:"toll_roads_info"`
	ZoneInfo         common.ZoneInfo           `json:"zone_info"`
}
