package config

import "os"

var Host string = "0.0.0.0"
var Port string = "4444"
var GetGeneralOrderInfoEndpoint string
var GetZoneInfoEndpoint string
var GetExecutorProfileEndpoint string
var GetConfigsEndpoint string
var GetTollRoadsInfoEndpoint string

func setIfEnv(x *string, ename string) bool {
	e, ok := os.LookupEnv(ename)
	if !ok {
		return false
	}
	*x = e
	return true
}

func init() {
	setIfEnv(&Host, "MOP_HOST")
	setIfEnv(&Port, "MOP_PORT")
	setIfEnv(&GetGeneralOrderInfoEndpoint, "GET_GENERAL_ORDER_INFO_ENDPOINT")
	setIfEnv(&GetZoneInfoEndpoint, "GET_ZONE_INFO_ENDPOINT")
	setIfEnv(&GetExecutorProfileEndpoint, "GET_EXECUTOR_PROFILE_ENDPOINT")
	setIfEnv(&GetConfigsEndpoint, "GET_CONFIGS_ENDPOINT")
	setIfEnv(&GetTollRoadsInfoEndpoint, "GET_TOLL_ROADS_INFO_ENDPOINT")
}
