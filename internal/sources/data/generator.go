package main

import (
	"fmt"
	"malomopa/internal/common"
	"math/rand"
	"strconv"
)

func Gen1() {
	it := 100
	r := rand.New(rand.NewSource(42))

	// General info
	general := map[string]common.GeneralOrderInfo{}
	for i := 0; i < it; i++ {
		v := common.GeneralOrderInfo{
			ID:             strconv.Itoa(i),
			UserID:         "some_user",
			ZoneID:         strconv.Itoa(i),
			BaseCoinAmount: 5.2,
		}
		general[v.ID] = v
	}
	common.WriteJSONToFile("/home/kazalika/malomopa/internal/sources/data/gen1/general_info.json", general)

	// Zones info
	zones := map[string]common.ZoneInfo{}
	for i := 0; i < it; i++ {
		v := common.ZoneInfo{
			ID:          strconv.Itoa(i),
			CoinCoeff:   42.52,
			DisplayName: fmt.Sprintf("Display name for zone #%d", i),
		}
		zones[v.ID] = v
	}
	common.WriteJSONToFile("/home/kazalika/malomopa/internal/sources/data/gen1/zones_info.json", zones)

	// Executor profiles
	profiles := map[string]common.ExecutorProfile{}
	for i := 0; i < it; i++ {
		v := common.ExecutorProfile{
			ID:     strconv.Itoa(i),
			Tags:   []string{"gen1"},
			Rating: r.Float32() * 100,
		}
		profiles[v.ID] = v
	}
	common.WriteJSONToFile("/home/kazalika/malomopa/internal/sources/data/gen1/executor_profiles.json", profiles)

	// Configs
	configs := common.AssignOrderConfigs{
		CoinCoeffCfg: &common.CoinCoeffConfig{
			Max: 123.321,
		},
	}
	common.WriteJSONToFile("/home/kazalika/malomopa/internal/sources/data/gen1/configs.json", configs)

	// Toll Roads info
	roads := map[string]common.TollRoadsInfo{}
	for i := 0; i < it; i++ {
		v := common.TollRoadsInfo{
			BonusAmount: 100.0 + float32(i),
		}
		roads[fmt.Sprintf("Display name for zone #%d", i)] = v
	}
	common.WriteJSONToFile("/home/kazalika/malomopa/internal/sources/data/gen1/toll_roads_info.json", roads)
}

func main() {
	Gen1()
}
