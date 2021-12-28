package lib

type SLA struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Details Detail `json:"details"`
}

type Detail struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Name       string      `json:"Name"`
	Provider   Entity      `json:"provider"`
	Client     Entity      `json:"client"`
	Creation   string      `json:"creation"`
	Guarantees []Guarantee `json:"guarantees"`
}

type Entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Guarantee struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
}

type Violation struct {
	ID         string `json:"ID"`
	ContractID string `json:"ContractID"`
}
