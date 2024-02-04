// Code generated by "enumer -type Status -linecomment -json -output health_string.go"; DO NOT EDIT.

package health

import (
	"encoding/json"
	"fmt"
)

const _StatusName = "UPDOWNOUT OF SERVICEUNKNOWN"

var _StatusIndex = [...]uint8{0, 2, 6, 20, 27}

func (i Status) String() string {
	if i < 0 || i >= Status(len(_StatusIndex)-1) {
		return fmt.Sprintf("Status(%d)", i)
	}
	return _StatusName[_StatusIndex[i]:_StatusIndex[i+1]]
}

var _StatusValues = []Status{0, 1, 2, 3}

var _StatusNameToValueMap = map[string]Status{
	_StatusName[0:2]:   0,
	_StatusName[2:6]:   1,
	_StatusName[6:20]:  2,
	_StatusName[20:27]: 3,
}

// StatusString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func StatusString(s string) (Status, error) {
	if val, ok := _StatusNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to Status values", s)
}

// StatusValues returns all values of the enum
func StatusValues() []Status {
	return _StatusValues
}

// IsAStatus returns "true" if the value is listed in the enum definition. "false" otherwise
func (i Status) IsAStatus() bool {
	for _, v := range _StatusValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for Status
func (i Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Status
func (i *Status) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("Status should be a string, got %s", data)
	}

	var err error
	*i, err = StatusString(s)
	return err
}
