package types

import (
	"encoding/json"

	"github.com/hashicorp/go-version"
)

type ZBXVersion version.Version

func NewZBXVersion(input string) (*ZBXVersion, error) {
	parsed, err := version.NewVersion(input)
	if err != nil {
		return nil, err
	}

	ver := ZBXVersion(*parsed)
	return &ver, nil
}

func (v *ZBXVersion) String() string {
	ver := version.Version(*v)
	return ver.String()
}

func (v *ZBXVersion) Compare(other *ZBXVersion) int {
	ver := version.Version(*v)
	o := version.Version(*other)
	return ver.Compare(&o)
}

func (v *ZBXVersion) LessThan(o *ZBXVersion) bool {
	return v.Compare(o) < 0
}

func (v *ZBXVersion) UnmarshalJSON(data []byte) error {
	var apiVersion string
	if err := json.Unmarshal(data, &apiVersion); err != nil {
		return err
	}

	parsed, err := version.NewVersion(apiVersion)
	if err != nil {
		return err
	}

	*v = ZBXVersion(*parsed)

	return nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (v *ZBXVersion) UnmarshalText(b []byte) error {
	temp, err := version.NewVersion(string(b))
	if err != nil {
		return err
	}

	*v = ZBXVersion(*temp)
	return nil
}

// MarshalText implements encoding.TextMarshaler interface.
func (v *ZBXVersion) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
