package db

import (
	"encoding/json"
	"github.com/gomodule/redigo/redis"
	"reflect"
	"sort"
	"time"
)

// A connection that can perform operations against a database.
type Connection interface {
	Do(command string, args ...interface{}) (interface{}, error)
}

// Get will load and unmarshal a database entry for the provided key into the
// provided object.  If the entry isn't present in the database then no error
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

// GetAll will load and unmarshal multiple database entries for the provided
// slice of keys and return the values in a map indexed by key.  If a particular
// key is not found then the value of kind, will be used in its place.  For
// each key that is found, a new instance of the same type as kind will be
// created and the json value will be unmarshalled into it.  If a value can't
// be properly unmarshalled then a json error will be returned.
func GetAll(c Connection, keys []string, kind interface{}) (map[string]interface{}, error) {
	var args []interface{}
	for _, key := range keys {
		args = append(args, key)
	}
	if len(args) == 0 {
		return nil, nil
	}

	bss, err := redis.ByteSlices(c.Do("MGET", args...))
	if err != nil {
		return nil, err
	}

	values := make(map[string]interface{})
	for i, bs := range bss {
		key := keys[i]
		if bs == nil {
			values[key] = kind
			continue
		}

		// If we pass a &interface{} into json.Unmarshall then it's just going to
		// unmarshall into a map[string]interface{}.  Because of this we need to
		// pass in a pointer to the underlying type of kind.  We'll use the reflect
		// package to determine what this is.
		ptr := reflect.New(reflect.TypeOf(kind))

		if err := json.Unmarshal(bs, ptr.Interface()); err != nil {
			return nil, err
		}

		// Now that we've unmarshalled into the proper type, again use the reflect
		// package to dereference the pointer and convert it to an interface{} for
		// returning.
		values[key] = reflect.Indirect(ptr).Interface()
	}

	return values, nil
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
