package lib

type Quality struct {
	Total int `json:"total"`
	High  int `json:"high_quality"`
	Low   int `json:"low_quality"`
}

type Part struct {
	MA           string             `json:"MA"`
	Timestamp    string             `json:"TimeStamp"`
	Version      int                `json:"Version"`
	DocumentType string             `json:"DocumentType"`
	DocumentBody Part_document_body `json:"DocumentBody"`
}

type Part_id struct {
	Oid string `json:"$oid"`
}

type Part_timestamp struct {
	Date int `json:"$date"`
}

type Part_document_body struct {
	Start                  Part_timestamp   `json:"Start"`
	Stop                   Part_timestamp   `json:"Stop"`
	CycleTime              float32          `json:"CycleTime"`
	Duration               float32          `json:"Duration"`
	ActiveTime             float32          `json:"ActiveTime"`
	Quality                int              `json:"Quality"`
	ElectricityConsumption []float32        `json:"ElectricityConsumption"`
	LoadingStop            Part_timestamp   `json:"LoadingStop"`
	LoadingTime            float32          `json:"LoadingTime"`
	ClampingStarts         []Part_timestamp `json:"ClampingStarts"`
	ClampingStops          []Part_timestamp `json:"ClampingStops"`
	ClampingTimes          []float32        `json:"ClampingTimes"`
	AdjustingStarts        []Part_timestamp `json:"AdjustingStarts"`
	AdjustingStops         []Part_timestamp `json:"AdjustingStops"`
	AdjustingTimes         []float32        `json:"AdjustingTimes"`
	MachiningStarts        []Part_timestamp `json:"MachiningStarts"`
	MachiningStops         []Part_timestamp `json:"MachiningStops"`
	MachiningTimes         []float32        `json:"MachiningTimes"`
	ReleasingStarts        []Part_timestamp `json:"ReleasingStarts"`
	ReleasingStops         []Part_timestamp `json:"ReleasingStops"`
	ReleasingTimes         []float32        `json:"ReleasingTimes"`
	UnloadingStart         Part_timestamp   `json:"UnloadingStart"`
	UnloadingTime          float32          `json:"UnloadingTime"`
	Pallet                 int              `json:"Pallet"`
	FeedOverride           float32          `json:"FeedOverride"`
	FeedOverrideRapid      float32          `json:"FeedOverrideRapid"`
	SpindleOverride        float32          `json:"SpindleOverride"`
	ToolChangeOveride      float32          `json:"ToolChangeOveride"`
	SpindleNumber          int              `json:"SpindleNumber"`
	ReleaseLocked          bool             `json:"ReleaseLocked"`
	NokTBr                 bool             `json:"NokTBr"`
	NokTMo                 bool             `json:"NokTMo"`
	NokRew                 bool             `json:"NokRew"`
	NokCla                 bool             `json:"NokCla"`
	NokWpc                 bool             `json:"NokWpc"`
	NokNcP                 bool             `json:"NokNcP"`
	ProductionCondUnavail  bool             `json:"ProductionCondUnavail,omitempty"`
	PalletchangeStarts     []Part_timestamp `json:"PalletchangeStarts,omitempty"`
	PalletchangeStops      []Part_timestamp `json:"PalletchangeStops,omitempty"`
	PalletchangeTimes      []float32        `json:"PalletchangeTimes,omitempty"`
	CarrierID              int              `json:"CarrierID"`
	TargetCycleTime        float32          `json:"TargetCycleTime,omitempty"`
	ComponentCode          string           `json:"ComponentCode,omitempty"`
	ComponentName          string           `json:"ComponentName,omitempty"`
	ComponentIdent         string           `json:"ComponentIdent,omitempty"`
	ComponentVersion       string           `json:"ComponentVersion,omitempty"`
	CycleTimeLoss          int              `json:"CycleTimeLoss,omitempty"`
	CycleTimeGain          int              `json:"CycleTimeGain,omitempty"`
	MongoRef               MongRef_s        `json:"mongo_ref"`
}

type MongRef_s struct {
	Id         Part_id `json:"_id"`
	Collection string  `json:"collection"`
}
