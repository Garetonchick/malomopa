package sources

import (
	"errors"
	"malomopa/internal/common"
)

type infoImpl struct {
	generalInfo       map[string]common.GeneralOrderInfo
	zonesInfo         map[string]common.ZoneInfo
	executorsProfiles map[string]common.ExecutorProfile
	configs           common.CoinCoeffConfig
	tollRoadsInfo     map[string]common.TollRoadsInfo
}

func NewFakeInfo(config DataPathsConfig) (*infoImpl, error) {
	var i infoImpl
	var err error

	i.generalInfo, err = common.ReadJSONFromFile[map[string]common.GeneralOrderInfo](config.GeneralOrdersInfoPath)
	if err != nil {
		return nil, err
	}

	i.zonesInfo, err = common.ReadJSONFromFile[map[string]common.ZoneInfo](config.ZonesInfoPath)
	if err != nil {
		return nil, err
	}

	i.executorsProfiles, err = common.ReadJSONFromFile[map[string]common.ExecutorProfile](config.ExecutorsProfilesPath)
	if err != nil {
		return nil, err
	}

	i.configs, err = common.ReadJSONFromFile[common.CoinCoeffConfig](config.ConfigsPath)
	if err != nil {
		return nil, err
	}

	i.tollRoadsInfo, err = common.ReadJSONFromFile[map[string]common.TollRoadsInfo](config.TollRoadsInfoPath)
	if err != nil {
		return nil, err
	}

	return &i, nil
}

func (i *infoImpl) GeneralInfo(id string) (*common.GeneralOrderInfo, error) {
	v, ok := i.generalInfo[id]
	if !ok {
		return nil, errors.New("element not found")
	}
	return &v, nil
}

func (i *infoImpl) ZoneInfo(id string) (*common.ZoneInfo, error) {
	v, ok := i.zonesInfo[id]
	if !ok {
		return nil, errors.New("element not found")
	}
	return &v, nil
}

func (i *infoImpl) ExecutorProfile(id string) (*common.ExecutorProfile, error) {
	v, ok := i.executorsProfiles[id]
	if !ok {
		return nil, errors.New("element not found")
	}
	return &v, nil
}

func (i *infoImpl) Configs() (*common.CoinCoeffConfig, error) {
	return &i.configs, nil
}

func (i *infoImpl) TollRoadsInfo(displayName string) (*common.TollRoadsInfo, error) {
	v, ok := i.tollRoadsInfo[displayName]
	if !ok {
		return nil, errors.New("element not found")
	}
	return &v, nil
}
