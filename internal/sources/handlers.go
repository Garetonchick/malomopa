package sources

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

//
// Helpers

func sendResponse(w http.ResponseWriter, resp any) {
	respJSON, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "error while marshaling response", http.StatusInternalServerError)
		return
	}
	if _, err := w.Write(respJSON); err != nil {
		http.Error(w, "error while writing response", http.StatusInternalServerError)
	}
}

func readRequestByID(r *http.Request) (GetByIdRequest, error) {
	var req GetByIdRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	return req, err
}

//
// Handlers

func (s *Server) Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping\n"))
}

func (s *Server) GetGeneralInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetGeneralInfo handler got a request!")
	s.Counters.GeneralInfoCounter.Add(1)

	req, err := readRequestByID(r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := FakeInfo.GeneralInfo(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("general info by id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func (s *Server) GetZoneInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetZoneInfo handler got a request!")
	if !s.ZonesInfoHandlerAvailability.Load() {
		log.Println("GetZoneInfo handler is turned off!")
		// Do nothing or return some error?
		return
	}
	s.Counters.ZoneInfoCounter.Add(1)

	req, err := readRequestByID(r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := FakeInfo.ZoneInfo(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("zone info by id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func (s *Server) GetExecutorProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("GetExecutorProfile handler got a request!")
	s.Counters.ExecutorProfileCounter.Add(1)

	req, err := readRequestByID(r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := FakeInfo.ExecutorProfile(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("executor with such id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func (s *Server) GetConfigs(w http.ResponseWriter, r *http.Request) {
	log.Println("GetConfigs handler got a request!")
	if !s.ZonesInfoHandlerAvailability.Load() {
		log.Println("GetConfigs handler is turned off!")
		// Do nothing or return some error?
		return
	}
	s.Counters.ConfigsCounter.Add(1)

	resp, err := FakeInfo.Configs()
	if err != nil {
		http.Error(w, "seems like configs undefined", http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func (s *Server) GetTollRoadsInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetTollRoadsInfo handler got a request!")
	s.Counters.TollRoadsInfoCountter.Add(1)

	var req GetTollRoadsRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := FakeInfo.TollRoadsInfo(req.ZoneDisplayName)
	if err != nil {
		http.Error(w, fmt.Sprintf("zone with such zone_display_name=`%s` doesn't exist", req.ZoneDisplayName), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func (s *Server) TurnOffConfigs(w http.ResponseWriter, r *http.Request) {
	log.Println("Turning off `Configs` handler")

	s.ConfigsHandlerAvailability.Store(false)
}

func (s *Server) TurnOnConfigs(w http.ResponseWriter, r *http.Request) {
	log.Println("Turning on `Configs` handler")

	s.ConfigsHandlerAvailability.Store(true)
}

func (s *Server) TurnOffZonesInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Turning off `ZoneInfo` handler")

	s.ZonesInfoHandlerAvailability.Store(false)
}

func (s *Server) TurnOnZonesInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Turning on `ZoneInfo` handler")

	s.ZonesInfoHandlerAvailability.Store(true)
}

func (s *Server) GetCounters(w http.ResponseWriter, r *http.Request) {
	log.Println("Got `GetCounters` request")

	res := HandlersCountersResponse{
		GeneralInfoCounter:     int(s.Counters.GeneralInfoCounter.Load()),
		ZoneInfoCounter:        int(s.Counters.ZoneInfoCounter.Load()),
		ExecutorProfileCounter: int(s.Counters.ExecutorProfileCounter.Load()),
		ConfigsCounter:         int(s.Counters.ConfigsCounter.Load()),
		TollRoadsInfoCountter:  int(s.Counters.TollRoadsInfoCountter.Load()),
	}

	b, err := json.Marshal(res)
	if err != nil {
		http.Error(w, fmt.Sprintf("error while marshaling counter: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
