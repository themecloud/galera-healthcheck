package healthcheck

import (
	"database/sql"
)

const (
	SYNCED_STATE         = "4"
	DONOR_DESYNCED_STATE = "2"
)

type Healthchecker struct {
	db     *sql.DB
	config HealthcheckerConfig
}

type HealthcheckerConfig struct {
	AvailableWhenDonor    bool
	AvailableWhenReadOnly bool
}

func New(db *sql.DB, config HealthcheckerConfig) *Healthchecker {
	return &Healthchecker{
		db:     db,
		config: config,
	}
}

func (h *Healthchecker) Check() (bool, string) {
	var variable_name string
	var value string
	err := h.db.QueryRow("SHOW STATUS LIKE 'wsrep_local_state'").Scan(&variable_name, &value)

	switch {
	case err != nil:
		return false, err.Error()
	case value == SYNCED_STATE || (value == DONOR_DESYNCED_STATE && h.config.AvailableWhenDonor):
		if !h.config.AvailableWhenReadOnly {
			var ro_variable_name string
			var ro_value string
			ro_err := h.db.QueryRow("SHOW GLOBAL VARIABLES LIKE 'read_only'").Scan(&ro_variable_name, &ro_value)
			switch {
			case ro_err != nil:
				return false, ro_err.Error()
			case ro_value == "ON":
				return false, "read-only"
			}
		}
		return true, "synced"
	default:
		return false, "not synced"
	}
}
