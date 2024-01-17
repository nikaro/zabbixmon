package types

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ZBXBoolean bool

func (bit *ZBXBoolean) UnmarshalJSON(data []byte) error {
	asString := string(data)
	if asString == "1" || asString == "true" {
		*bit = true
	} else if asString == "0" || asString == "false" {
		*bit = false
	} else {
		return errors.New(fmt.Sprintf("Boolean unmarshal error: invalid input %s", asString))
	}
	return nil
}

func (bit ZBXBoolean) MarshalJSON() ([]byte, error) {
	if bit == true {
		return json.Marshal("1")
	}

	return json.Marshal("0")
}
