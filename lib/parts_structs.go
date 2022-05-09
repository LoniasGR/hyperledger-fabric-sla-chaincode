package lib

type Quality struct {
	Total int `json:"total"`
	High  int `json:"high_quality"`
	Low   int `json:"low_quality"`
}

type Part struct {
	Id           Part_id            `json:"_id"`
	MA           string             `json:"MA"`
	Timestamp    Part_timestamp     `json:"TimeStamp"`
	Version      int                `json:"Version"`
	DocumentType string             `json:"DocumentType"`
	DocumentBody Part_document_body `json:"DocumentBody"`
}

type Part_id struct {
	Oid string `json:"$oid"`
}

type Part_timestamp struct {
	Date string `json:"$date"`
}

type Part_document_body struct {
	Start                 Part_timestamp   `json:"Start"`
	Stop                  Part_timestamp   `json:"Stop"`
	CycleTime             int              `json:"CycleTime"`
	Duration              float32          `json:"Duration"`
	ActiveTime            float32          `json:"ActiveTime"`
	Quality               int              `json:"Quality"`
	LoadingStop           Part_timestamp   `json:"LoadingStop"`
	LoadingTime           float32          `json:"LoadingTime"`
	ClampingStarts        []Part_timestamp `json:"ClampingStarts"`
	ClampingStops         []Part_timestamp `json:"ClampingStops"`
	ClampingTimes         []float32        `json:"ClampingTimes"`
	AdjustingStarts       []Part_timestamp `json:"AdjustingStarts"`
	AdjustingStops        []Part_timestamp `json:"AdjustingStops"`
	AdjustingTimes        []float32        `json:"AdjustingTimes"`
	ReleasingStarts       []Part_timestamp `json:"ReleasingStarts"`
	ReleasingStops        []Part_timestamp `json:"ReleasingStops"`
	ReleasingTimes        []float32        `json:"ReleasingTimes"`
	UnloadingStart        Part_timestamp   `json:"UnloadingStart"`
	UnloadingTime         float32          `json:"UnloadingTime"`
	Pallet                int              `json:"Pallet"`
	FeedOverride          float32          `json:"FeedOverride"`
	FeedOverrideRapid     int              `json:"FeedOverrideRapid"`
	SpindleOverride       int              `json:"SpindleOverride"`
	ToolChangeOveride     float32          `json:"ToolChangeOveride"`
	SpindleNumber         int              `json:"SpindleNumber"`
	ReleaseLocked         bool             `json:"ReleaseLocked"`
	NokTBr                bool             `json:"NokTBr"`
	NokTMo                bool             `json:"NokTMo"`
	NokRew                bool             `json:"NokRew"`
	NokCla                bool             `json:"NokCla"`
	NokWpc                bool             `json:"NokWpc"`
	NokNcP                bool             `json:"NokNcP"`
	ProductionCondUnavail bool             `json:"ProductionCondUnavail"`
	PalletchangeStarts    []Part_timestamp `json:"PalletchangeStarts"`
	PalletchangeStops     []Part_timestamp `json:"PalletchangeStops"`
	PalletchangeTimes     []float32        `json:"PalletchangeTimes"`
	CarrierID             int              `json:"CarrierID"`
	TargetCycleTime       *float32         `json:"TargetCycleTime"`
	ComponentCode         string           `json:"ComponentCode"`
	ComponentName         string           `json:"ComponentName"`
	ComponentIdent        string           `json:"ComponentIdent"`
	ComponentVersion      string           `json:"ComponentVersion"`
	CycleTimeLoss         int              `json:"CycleTimeLoss"`
	CycleTimeGain         int              `json:"CycleTimeGain"`
	MongoRef              MongRef_s        `json:"mongo_ref"`
}

type MongRef_s struct {
	Id         Part_id `json:"_id"`
	Collection string  `json:"collection"`
}
