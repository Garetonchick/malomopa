package config

type CacheServiceConfig struct {
	GetGeneralOrderInfoEndpoint string `json:"get_general_order_info_endpoint"`
	GetZoneInfoEndpoint         string `json:"get_zone_info_endpoint"`
	GetExecutorProfileEndpoint  string `json:"get_executor_profile_endpoint"`
	GetConfigsEndpoint          string `json:"get_configs_endpoint"`
	GetTollRoadsInfoEndpoint    string `json:"get_toll_roads_info_endpoint"`
}
