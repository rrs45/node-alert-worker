package types

import (
	"time"
)

//Status struct represents the state of remediations running currently
type Status struct {
	Action string // Ansible play name
	Params string
	Timestamp time.Time
}

//NewStatus returns pointer to an empty Status
func NewStatus() *Status {
	return &Status{}
}