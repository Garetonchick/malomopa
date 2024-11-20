package main

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

func readRequestByID(w http.ResponseWriter, r *http.Request) (GetByIdRequest, error) {
	var req GetByIdRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	return req, err
}

//
// Handlers

func Ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ping\n"))
}

func GetGeneralInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetGeneralInfo")

	req, err := readRequestByID(w, r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := fakeInfo.GeneralInfo(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("general info by id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func GetZoneInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetZoneInfo")

	req, err := readRequestByID(w, r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := fakeInfo.ZoneInfo(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("zone info by id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func GetExecutorProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("GetExecutorProfile")

	req, err := readRequestByID(w, r)
	if err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := fakeInfo.ExecutorProfile(req.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("executor with such id=`%s` doesn't exist", req.ID), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func GetConfigs(w http.ResponseWriter, r *http.Request) {
	log.Println("GetConfigs")

	resp, err := fakeInfo.Configs()
	if err != nil {
		http.Error(w, "seems like configs undefined", http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}

func GetTollRoadsInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("GetTollRoadsInfo")

	var req GetTollRoadsRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "request has wrong format", http.StatusBadRequest)
		return
	}

	resp, err := fakeInfo.TollRoadsInfo(req.ZoneDisplayName)
	if err != nil {
		http.Error(w, fmt.Sprintf("zone with such zone_display_name=`%s` doesn't exist", req.ZoneDisplayName), http.StatusExpectationFailed)
		return
	}

	sendResponse(w, resp)
}
