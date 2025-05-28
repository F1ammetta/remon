package types

import "time"

// Service represents a systemd service
type Service struct {
	Name        string    `json:"name"`
	LoadState   string    `json:"load_state"`
	ActiveState string    `json:"active_state"`
	SubState    string    `json:"sub_state"`
	Since       time.Time `json:"since"`
	Description string    `json:"description"`
}
