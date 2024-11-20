package main

import (
	"flag"
	"fmt"
	"log"
	common "malomopa/internal/common"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Info interface {
	GeneralInfo(string) (*common.GeneralOrderInfo, error)
	ZoneInfo(string) (*common.ZoneInfo, error)
	ExecutorProfile(string) (*common.ExecutorProfile, error)
	Configs() (*map[string]any, error)
	TollRoadsInfo(string) (*common.TollRoadsInfo, error)
}

var fakeInfo Info

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

type Server struct {
	mux *chi.Mux

	config HttpServerConfig
}

func NewServer(cfg HttpServerConfig) (*Server, error) {
	s := &Server{
		mux:    chi.NewRouter(),
		config: cfg,
	}

	s.mux.Get("/ping", Ping)
	s.mux.Get("/general_info", GetGeneralInfo)
	s.mux.Get("/zone_info", GetZoneInfo)
	s.mux.Get("/executor_profile", GetExecutorProfile)
	s.mux.Get("/configs", GetConfigs)
	s.mux.Get("/toll_roads_info", GetTollRoadsInfo)

	return s, nil
}

func (s *Server) Run() error {
	log.Print("Starting HTTP Server...")
	return http.ListenAndServe(fmt.Sprintf("%s:%s", s.config.Host, s.config.Port), s.mux)
}

func main() {
	configPath := flag.String("config", "", "Path to order assigner config")
	flag.Parse()
	if configPath == nil {
		log.Fatal("no config file provided")
	}

	cfg, err := common.ReadJSONFromFile[Config](*configPath)
	if err != nil {
		log.Fatal("config has wrong format or doesn't exist")
	}

	fakeInfo, err = NewFakeInfo(cfg.DataPaths)
	if err != nil {
		log.Fatal("fake info initiation failed :", err.Error())
	}

	s, err := NewServer(cfg.HttpServer)
	if err != nil {
		log.Fatal("something gone wrong: ", err.Error())
	}

	if err := s.Run(); err != nil {
		log.Fatal("`server.Run()` finished with error: ", err.Error())
	}
}
