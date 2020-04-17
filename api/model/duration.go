package model

import (
	"encoding/json"
	"time"
)

// Duration is an aliasing of time.Duration that supports marshalling to/from
// JSON.
type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(bs []byte) error {
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return err
	}

	td, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration{td}
	return nil
}
