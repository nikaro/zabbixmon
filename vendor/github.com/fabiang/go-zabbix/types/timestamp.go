package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type ZBXUnixTimestamp time.Time

func (t *ZBXUnixTimestamp) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%d\"", time.Time(*t).Unix())
	return []byte(stamp), nil
}

func (t *ZBXUnixTimestamp) UnmarshalJSON(data []byte) error {
	var timestamp string
	var unix int64

	err := json.Unmarshal(data, &timestamp)
	if err != nil {
		return err
	}

	unix, err = strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return err
	}

	// Fractional seconds are handled implicitly by Parse.
	*t = ZBXUnixTimestamp(time.Unix(unix, 0).UTC())

	return nil
}
