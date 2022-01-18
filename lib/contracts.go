package lib

import (
	"time"
)

type SLA struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	State      string     `json:"state"`
	Assessment Assessment `json:"assessment"`
	Details    Detail     `json:"details"`
}

type Detail struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Name       string      `json:"Name"`
	Provider   Entity      `json:"provider"`
	Client     Entity      `json:"client"`
	Creation   time.Time   `json:"creation"`
	Guarantees []Guarantee `json:"guarantees"`
	Service    string      `json:"service"`
}

type Assessment struct {
	FirstExecution time.Time `json:"first_execution"`
	LastExecution  time.Time `json:"last_execution"`
}

type Entity struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Importance struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint"`
}

type Guarantee struct {
	Name       string       `json:"name"`
	Constraint string       `json:"constraint"`
	Importance []Importance `json:"importance"`
}

type Violation struct {
	ID             string  `json:"id"`
	SLAID          string  `json:"sla_id"`
	GuaranteeID    string  `json:"guarantee_id"`
	Constraint     string  `json:"constraint"`
	Values         []Value `json:"values"`
	ImportanceName string  `json:"importanceName"`
	Importance     int     `json:"importance"`
	AppID          string  `json:"appID"`
}

type Value struct {
	Key      string    `json:"key"`
	Value    int64     `json:"value"`
	Datetime time.Time `json:"datetime"`
}
