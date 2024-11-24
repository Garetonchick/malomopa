package sources

import (
	"fmt"
	"log"
	common "malomopa/internal/common"
	"net/http"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
)

type Info interface {
	GeneralInfo(string) (*common.GeneralOrderInfo, error)
	ZoneInfo(string) (*common.ZoneInfo, error)
	ExecutorProfile(string) (*common.ExecutorProfile, error)
	Configs() (*common.CoinCoeffConfig, error)
	TollRoadsInfo(string) (*common.TollRoadsInfo, error)
}

var FakeInfo Info

type HttpServerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type DataPathsConfig struct {
	GeneralOrdersInfoPath string `json:"general_orders_info_path"`
	ZonesInfoPath         string `json:"zones_info_path"`
	ExecutorsProfilesPath string `json:"executors_profiles_path"`
	ConfigsPath           string `json:"configs_path"`
	TollRoadsInfoPath     string `json:"toll_roads_info_path"`
}

type Config struct {
	HttpServer HttpServerConfig `json:"http_server"`
	DataPaths  DataPathsConfig  `json:"data_paths"`
}

type HandlersCountersResponse struct {
	GeneralInfoCounter     int `json:"general_info_counter"`
	ZoneInfoCounter        int `json:"zone_info_counter"`
	ExecutorProfileCounter int `json:"executor_profile_counter"`
	ConfigsCounter         int `json:"configs_counter"`
	TollRoadsInfoCountter  int `json:"toll_roads_info_counter"`
}

type HandlerCounters struct {
	GeneralInfoCounter     atomic.Int32
	ZoneInfoCounter        atomic.Int32
	ExecutorProfileCounter atomic.Int32
	ConfigsCounter         atomic.Int32
	TollRoadsInfoCountter  atomic.Int32
}

type Server struct {
	mux *chi.Mux

	config HttpServerConfig

	ConfigsHandlerAvailability   atomic.Bool
	ZonesInfoHandlerAvailability atomic.Bool

	Counters HandlerCounters
}

func NewServer(cfg HttpServerConfig) (*Server, error) {
	s := &Server{
		mux:    chi.NewRouter(),
		config: cfg,
	}
	s.ConfigsHandlerAvailability.Store(true)
	s.ZonesInfoHandlerAvailability.Store(true)

	s.mux.Get("/ping", s.Ping)
	s.mux.Get("/general_info", s.GetGeneralInfo)
	s.mux.Get("/zone_info", s.GetZoneInfo)
	s.mux.Get("/executor_profile", s.GetExecutorProfile)
	s.mux.Get("/configs", s.GetConfigs)
	s.mux.Get("/toll_roads_info", s.GetTollRoadsInfo)

	s.mux.Post("/zone_info_off", s.TurnOffZonesInfo)
	s.mux.Post("/zone_info_on", s.TurnOnZonesInfo)
	s.mux.Post("/configs_off", s.TurnOffConfigs)
	s.mux.Post("/configs_on", s.TurnOnConfigs)

	s.mux.Get("/counters", s.GetCounters)

	return s, nil
}

func (s *Server) Run() error {
	log.Print("Starting HTTP Server...")
	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.config.Host, s.config.Port), s.mux)
}
