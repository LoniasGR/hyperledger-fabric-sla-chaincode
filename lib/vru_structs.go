package lib

type Position_s struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Tram_s struct {
	StationID int32      `json:"station_id"`
	Position  Position_s `json:"position"`
}

type OBU_s struct {
	StationID int32      `json:"station_id"`
	Position  Position_s `json:"position"`
	Risk      string     `json:"risk"`
}

type VRU struct {
	Timestamp int64   `json:"timestamp"`
	Tram      Tram_s  `json:"tram"`
	OBUs      []OBU_s `json:"obus"`
}

type Risk struct {
	Critical int `json:"critical"`
	Warning  int `json:"warning"`
	HighRisk int `json:"highRisk"`
	LowRisk  int `json:"lowRisk"`
	NoRisk   int `json:"noRisk"`
}
