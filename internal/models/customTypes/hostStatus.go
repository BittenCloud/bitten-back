package customTypes

import (
	"database/sql/driver"
	"fmt"
)

// HostStatus defines the possible operational statuses of a host.
type HostStatus string

// Defines the set of valid host statuses.
const (
	StatusUnknown     HostStatus = "unknown"     // Status is not determined or is ambiguous.
	StatusActive      HostStatus = "active"      // Host is operational and actively serving.
	StatusInactive    HostStatus = "inactive"    // Host is intentionally not operational.
	StatusMaintenance HostStatus = "maintenance" // Host is temporarily down for maintenance.
)

// String satisfies the fmt.Stringer interface, returning the string representation of the HostStatus.
func (hs *HostStatus) String() string {
	return string(*hs)
}

// IsValid checks if the HostStatus value is one of the predefined valid statuses.
func (hs *HostStatus) IsValid() bool {
	switch *hs {
	case StatusUnknown, StatusActive, StatusInactive, StatusMaintenance:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface.
// This method defines how HostStatus will be stored in the database.
func (hs *HostStatus) Value() (driver.Value, error) {
	if !hs.IsValid() {
		return nil, fmt.Errorf("invalid HostStatus value for database storage: %s", *hs)
	}
	return string(*hs), nil
}

// Scan implements the sql.Scanner interface.
// This method defines how HostStatus will be read from the database.
func (hs *HostStatus) Scan(value interface{}) error {
	if value == nil {
		// If the database value is NULL, set to StatusUnknown as a default.
		*hs = StatusUnknown
		return nil
	}

	var strValue string
	switch v := value.(type) {
	case []byte:
		strValue = string(v)
	case string:
		strValue = v
	default:
		return fmt.Errorf("failed to scan HostStatus: unsupported type %T", value)
	}

	scannedStatus := HostStatus(strValue)

	if !scannedStatus.IsValid() {
		*hs = StatusUnknown
		return nil
	}
	*hs = scannedStatus
	return nil
}
