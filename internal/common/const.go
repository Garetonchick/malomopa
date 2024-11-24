package common

const (
	OrderIDQueryParam    = "order-id"
	ExecutorIDQueryParam = "executor-id"
)

type DataSourcesKeys struct {
	GeneralOrderInfo   string
	ZoneInfo           string
	ExecutorProfile    string
	AssignOrderConfigs string
	TollRoadsInfo      string
}

var Keys = DataSourcesKeys{
	GeneralOrderInfo:   "general_order_info",
	ZoneInfo:           "zone_info",
	ExecutorProfile:    "executor_profile",
	AssignOrderConfigs: "assign_order_configs",
	TollRoadsInfo:      "toll_roads_info",
}
