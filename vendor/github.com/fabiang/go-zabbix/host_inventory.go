package zabbix

import (
	"bytes"
	"encoding/json"
)

type HostInventory map[string]string

var phpEmptyArray = []byte("[]")

func (hi *HostInventory) UnmarshalJSON(data []byte) error {
	// fix for PHP maps are turned to array when they are empty
	if bytes.Equal(data, phpEmptyArray) {
		return nil
	}

	var inv map[string]string
	if err := json.Unmarshal(data, &inv); err != nil {
		return err
	}

	*hi = inv

	return nil
}
