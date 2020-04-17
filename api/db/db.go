package db

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"sort"
	"time"
)

// A connection that can perform operations against a database.
type Connection interface {
	Do(command string, args ...interface{}) (interface{}, error)
}

// Get will load and unmarshal a database entry for the provided key into the
// provided object.  If the entry isn't present in the database then ErrMissing
// will be returned.  If the entry can't be can't be properly unmarshalled then
// a json error will be returned.
func Get(c Connection, key string, data interface{}) error {
	bs, err := redis.Bytes(c.Do("GET", key))
	if err == redis.ErrNil {
		// There weren't any entries in redis for this key.  This is okay.
		return nil
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, &data)
}

// GetWithTTLRefresh will load and unmarshal a database entry for the provided
// key into the provided object while also refreshing the TTL of the key in the
// database.  If the entry isn't present in the database then ErrMissing will
// be returned and the TTL wn't be updated.  If the TTL cannot be properly
// updated or the entry cannot be unmarshalled from JSON then an error will be
// returned.
//
// Note, currently the get operation and the TTL update are not atomic
// operations that happen in a transaction.  They are sequential updates which
// means it is possible for the data to be loaded and unmarshalled but the TTL
// not updated (and an error returned).
func GetWithTTLRefresh(c Connection, key string, data interface{}, ttl time.Duration) error {
	// We could create a transaction to read the key and update its expiration
	// time atomically, but that's not really necessary.  The worst that will
	// happen is that we race to update the expiration time and in either case
	// last one wins which is what we want to happen.
	err := Get(c, key, data)
	if err != nil {
		return err
	}

	_, err = c.Do("EXPIRE", key, ttl.Seconds())
	if err != nil {
		return err
	}

	return nil
}

// Set will write the provided entry to the database for the provided key.  If
// the entry can't be marshalled to JSON or is unable to be written to the
// database for some reason then an error will be returned.
func Set(c Connection, key string, data interface{}) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = c.Do("SET", key, bs)
	return err
}

// SetWithTTL will write the provided entry to the database for the provided key
// with a TTL.  Once the TTL's time elapses the key will be automatically
// deleted from the database.  If the entry can't be marshalled to JSON or is
// unable to be written to the database for some reason then an error will be
// returned.
func SetWithTTL(c Connection, key string, data interface{}, ttl time.Duration) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = c.Do("SETEX", key, ttl.Seconds(), bs)
	return err
}

// ScanKeys will scan the database for keys that match the provided key (with
// wildcards).  Each matching key will be returned or an error returned if
// the database couldn't be scanned for some reason.
func ScanKeys(c Connection, key string) ([]string, error) {
	cursor := 0

	var keys []string
	for {
		fields, err := redis.Values(c.Do("SCAN", cursor, "MATCH", key))
		if err != nil {
			return nil, err
		}

		cursor, err = redis.Int(fields[0], nil)
		if err != nil {
			return nil, err
		}

		batch, err := redis.Strings(fields[1], nil)
		if err != nil {
			return nil, err
		}

		keys = append(keys, batch...)

		// Check if we need to stop.
		if cursor == 0 {
			break
		}
	}

	sort.Strings(keys)
	return keys, nil
}
