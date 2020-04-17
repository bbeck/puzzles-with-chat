package model

import (
	"encoding/json"
	"fmt"
)

// FontSize is an enumeration representing the supported sizes of a font.  It
// can be marshalled to/from JSON as well as implements the fmt.Stringer
// interface for human readability.
type FontSize int

const (
	FontSizeNormal FontSize = iota
	FontSizeLarge
	FontSizeXLarge
)

func (s FontSize) String() string {
	switch s {
	case FontSizeNormal:
		return "normal"
	case FontSizeLarge:
		return "large"
	case FontSizeXLarge:
		return "xlarge"
	default:
		return "unknown"
	}
}

func (s FontSize) MarshalJSON() ([]byte, error) {
	switch s {
	case FontSizeNormal:
	case FontSizeLarge:
	case FontSizeXLarge:
	default:
		return nil, fmt.Errorf("unrecognized font size: %v", s)
	}

	return json.Marshal(s.String())
}

func (s *FontSize) UnmarshalJSON(bs []byte) error {
	var str string
	if err := json.Unmarshal(bs, &str); err != nil {
		return err
	}

	switch str {
	case "normal":
		*s = FontSizeNormal
	case "large":
		*s = FontSizeLarge
	case "xlarge":
		*s = FontSizeXLarge
	default:
		return fmt.Errorf("unrecognized font size string: %s", str)
	}

	return nil
}
