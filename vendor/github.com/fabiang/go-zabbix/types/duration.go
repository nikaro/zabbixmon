package types

import (
	"regexp"
	"strconv"
	"time"
)

type ZBXDuration time.Duration

var durationRegexp = regexp.MustCompile(`^\"(?P<time>\d+)(?P<suffix>(s|m|h|d|w|M|y))?\"$`)

//func (t ZBXDuration) MarshalJSON() ([]byte, error) {
//
//}

func (t *ZBXDuration) UnmarshalJSON(data []byte) (err error) {
	match := durationRegexp.FindSubmatch(data)

	if len(match) == 0 {
		*t = 0
		return nil
	}

	time, err := strconv.ParseInt(string(match[1]), 10, 32)
	if err != nil {
		return err
	}

	var suffix rune
	if len(match[2]) > 0 {
		suffix = rune(match[2][0])
	}

	// to nano
	time *= 1_000_000_000

	switch suffix {
	case 's': // already in seconds
	case '\x00': // no suffix means seconds
	case 'm':
		time *= 60
	case 'h':
		time *= 60 * 60
	case 'd':
		time *= 60 * 60 * 24
	case 'w':
		time *= 60 * 60 * 24 * 7
	case 'M':
		time *= 60 * 60 * 24 * 30
	case 'y':
		time *= 60 * 60 * 24 * 365
	}

	*t = ZBXDuration(time)
	return nil
}
