package lib

type Tram_s struct {
	StationID  int32   `json:"station_id"`
	Latitude   float64 `json:"latitude"`
	Longditute float64 `json:"longitude"`
}

type OBU_s struct {
	StationID  int32   `json:"station_id"`
	Latitude   float64 `json:"latitude"`
	Longditute float64 `json:"longitude"`
	Risk       string  `json:"risk"`
}

type VRU struct {
	Timestamp int64  `json:"timestamp"`
	Tram      Tram_s `json:"tram"`
	OBUs      []OBU_s `json:"obus"`
}
