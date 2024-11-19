package main

type GetByIdRequest struct {
	ID string `json:"id"`
}

type GetTollRoadsRequest struct {
	ZoneDisplayName string `json:"zone_display_name"`
}
