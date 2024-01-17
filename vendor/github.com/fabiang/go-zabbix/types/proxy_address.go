package types

import (
	"encoding/json"
	"strings"
)

// ProxyAddresses IP addresses or DNS names of active Zabbix proxy.
type ZBXProxyAddresses []string

func (addr *ZBXProxyAddresses) UnmarshalJSON(data []byte) error {
	var input string
	err := json.Unmarshal(data, &input)
	if err != nil {
		return err
	}

	*addr = strings.Split(input, ",")
	return nil
}
func (addr *ZBXProxyAddresses) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.Join(*addr, ","))
}
