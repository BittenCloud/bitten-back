package customTypes

import (
	"database/sql/driver"
	"fmt"
)

// DurationUnit defines the unit of measurement for a duration.
type DurationUnit string

// Defines the possible values for DurationUnit.
const (
	UnitDay   DurationUnit = "day"
	UnitMonth DurationUnit = "month"
	UnitYear  DurationUnit = "year"
)

// String satisfies the fmt.Stringer interface.
func (du *DurationUnit) String() string {
	return string(*du)
}

// IsValid checks if the DurationUnit value is one of the defined valid units.
func (du *DurationUnit) IsValid() bool {
	switch *du {
	case UnitDay, UnitMonth, UnitYear:
		return true
	default:
		return false
	}
}

// Value implements the driver.Valuer interface.
// This method defines how DurationUnit will be stored in the database.
func (du *DurationUnit) Value() (driver.Value, error) {
	if !du.IsValid() {
		return nil, fmt.Errorf("invalid DurationUnit value for database storage: %s", *du)
	}
	return string(*du), nil
}

// Scan implements the sql.Scanner interface.
// This method defines how DurationUnit will be read from the database.
func (du *DurationUnit) Scan(value interface{}) error {
	if value == nil {
		// Handle NULL from database; perhaps set to a default.
		*du = ""
		return nil
	}

	var strValue string
	switch v := value.(type) {
	case []byte:
		strValue = string(v)
	case string:
		strValue = v
	default:
		return fmt.Errorf("failed to scan DurationUnit: unsupported type %T", value)
	}

	scannedUnit := DurationUnit(strValue)

	if strValue != "" && !scannedUnit.IsValid() {
		return fmt.Errorf("invalid DurationUnit value '%s' from database", strValue)
	}
	*du = scannedUnit
	return nil
}
