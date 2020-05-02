package db

import (
	"encoding/json"
	"errors"
	"github.com/alicebob/miniredis"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	type Entry struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name     string
		initial  map[string]Entry // Entries that should be present for the test.
		key      string           // The key of the entry to retrieve.
		expected Entry            // The entry that we expect to retrieve.
	}{
		{
			name:     "missing entry",
			key:      "1",
			expected: Entry{},
		},
		{
			name: "single entry",
			initial: map[string]Entry{
				"1": {1},
			},
			key:      "1",
			expected: Entry{1},
		},
		{
			name: "multiple entries",
			initial: map[string]Entry{
				"1": {1},
				"2": {2},
				"3": {3},
			},
			key:      "1",
			expected: Entry{1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, entry := range test.initial {
				bs, err := json.Marshal(entry)
				require.NoError(t, err)

				err = server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// Get the entry we care about.
			var entry Entry
			err := Get(conn, test.key, &entry)

			// Verify we received our entry.
			require.NoError(t, err)
			assert.Equal(t, test.expected, entry)
		})
	}
}

func TestGet_Error(t *testing.T) {
	type Entry struct{}

	tests := []struct {
		name       string
		connection ConnectionFunc
		initial    map[string][]byte // Entries that should be present for the test.
		key        string            // The key of the entry to retrieve.
		expected   error
	}{
		{
			name: "conn.Do error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("forced error")
			},
			key:      "key",
			expected: errors.New("forced error"),
		},
		{
			name: "json.Unmarshal error",
			initial: map[string][]byte{
				"key": []byte(""),
			},
			key: "key",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, bs := range test.initial {
				err := server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server.
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			err := Get(connection, test.key, &Entry{})

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	type Entry struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name     string
		initial  map[string]Entry       // Entries that should be present for the test.
		keys     []string               // The keys that should be retrieved.
		expected map[string]interface{} // The entries expected to be returned.
	}{
		{
			name: "no keys",
			keys: []string{},
		},
		{
			name: "missing all entries",
			keys: []string{"1", "2"},
			expected: map[string]interface{}{
				"1": Entry{},
				"2": Entry{},
			},
		},
		{
			name: "one key",
			initial: map[string]Entry{
				"1": {1},
			},
			keys: []string{"1"},
			expected: map[string]interface{}{
				"1": Entry{1},
			},
		},
		{
			name: "multiple keys",
			initial: map[string]Entry{
				"1": {1},
				"2": {2},
			},
			keys: []string{"1", "2"},
			expected: map[string]interface{}{
				"1": Entry{1},
				"2": Entry{2},
			},
		},
		{
			name: "some keys missing",
			initial: map[string]Entry{
				"1": {1},
				"3": {3},
			},
			keys: []string{"1", "2", "3"},
			expected: map[string]interface{}{
				"1": Entry{1},
				"2": Entry{},
				"3": Entry{3},
			},
		},
		{
			name: "extra keys in database",
			initial: map[string]Entry{
				"1": {1},
				"2": {2},
				"3": {3},
			},
			keys: []string{"1", "2"},
			expected: map[string]interface{}{
				"1": Entry{1},
				"2": Entry{2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, entry := range test.initial {
				bs, err := json.Marshal(entry)
				require.NoError(t, err)

				err = server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// Get the entries we care about.
			values, err := GetAll(conn, test.keys, Entry{})

			// Verify our results.
			require.NoError(t, err)
			assert.Equal(t, test.expected, values)
		})
	}
}

func TestGetAll_Error(t *testing.T) {
	type Entry struct{}

	tests := []struct {
		name       string
		connection ConnectionFunc
		initial    map[string][]byte // Entries that should be present for the test.
		keys       []string          // The keys of the entries to retrieve.
		expected   error
	}{
		{
			name: "conn.Do error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("forced error")
			},
			keys:     []string{"key"},
			expected: errors.New("forced error"),
		},
		{
			name: "json.Unmarshal error",
			initial: map[string][]byte{
				"key": []byte(""),
			},
			keys: []string{"key"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, bs := range test.initial {
				err := server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server.
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			_, err := GetAll(connection, test.keys, Entry{})

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

func TestGetWithTTLRefresh(t *testing.T) {
	type Entry struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name        string
		initial     map[string]Entry // Entries that should be present for the test.
		key         string           // The key of the entry to retrieve.
		ttl         time.Duration    // The TTL to set when retrieving the entry.
		expectedTTL time.Duration    // The TTL that the key is expected to have.
	}{
		{
			name:        "missing entry",
			key:         "1",
			ttl:         time.Hour,
			expectedTTL: 0, // the key doesn't exist, thus it can't have an expiration
		},
		{
			name: "single entry",
			initial: map[string]Entry{
				"1": {1},
			},
			key:         "1",
			ttl:         time.Hour,
			expectedTTL: time.Hour,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, entry := range test.initial {
				bs, err := json.Marshal(entry)
				require.NoError(t, err)

				err = server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// Get the entry we care about
			err := GetWithTTLRefresh(conn, test.key, &Entry{}, test.ttl)

			// Verify we received our entry and it has the correct TTL.
			require.NoError(t, err)
			assert.Equal(t, test.expectedTTL, server.TTL(test.key))
		})
	}
}

func TestGetWithTTLRefresh_Error(t *testing.T) {
	server, conn := NewMiniredis(t)

	type Entry struct{}

	tests := []struct {
		name       string
		connection ConnectionFunc
		initial    map[string][]byte // Entries that should be present for the test.
		key        string            // The key of the entry to retrieve.
		expected   error
	}{
		{
			name: "conn.Do error on EXPIRE",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				if command == "EXPIRE" {
					return nil, errors.New("forced error")
				}
				return conn.Do(command, args...)
			},
			initial: map[string][]byte{
				"key": []byte("{}"),
			},
			key:      "key",
			expected: errors.New("forced error"),
		},
		{
			name: "json.Unmarshal error",
			initial: map[string][]byte{
				"key": []byte(""),
			},
			key: "key",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server.FlushAll()

			// Write all of the initial entries into the database.
			for key, bs := range test.initial {
				err := server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			err := GetWithTTLRefresh(connection, test.key, &Entry{}, time.Hour)

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

func TestSet(t *testing.T) {
	type Entry struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name     string
		initial  map[string]Entry // Entries that should be present for the test.
		key      string           // The key of the data to write.
		entry    Entry            // The entry to write.
		expected string
	}{
		{
			name:     "missing entry",
			key:      "key",
			entry:    Entry{1},
			expected: `{"id":1}`,
		},
		{
			name: "existing overwritten",
			initial: map[string]Entry{
				"key": {0},
			},
			key:      "key",
			entry:    Entry{1},
			expected: `{"id":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, entry := range test.initial {
				bs, err := json.Marshal(entry)
				require.NoError(t, err)

				err = server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// Write the entry we care about.
			err := Set(conn, test.key, test.entry)

			// Verify we wrote our entry.
			require.NoError(t, err)
			server.CheckGet(t, test.key, test.expected)
		})
	}
}

func TestSet_Error(t *testing.T) {
	tests := []struct {
		name       string
		connection ConnectionFunc
		key        string      // The key of the data to write.
		data       interface{} // The data to write.
		expected   error
	}{
		{
			name: "json.Marshal error",
			key:  "key",
			data: make(chan int), // Channels are not able to be marshalled to JSON.
		},
		{
			name: "conn.Do error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("forced error")
			},
			key:      "key",
			expected: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, conn := NewMiniredis(t)

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server.
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			err := Set(connection, test.key, test.data)

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

func TestSetWithTTL(t *testing.T) {
	type Entry struct {
		Id int `json:"id"`
	}

	tests := []struct {
		name     string
		initial  map[string]Entry // Entries that should be present for the test.
		key      string           // The key of the data to write.
		entry    Entry            // The entry to write.
		ttl      time.Duration    // The TTL to write the entry with.
		expected string
	}{
		{
			name:     "missing entry",
			key:      "key",
			entry:    Entry{1},
			ttl:      time.Hour,
			expected: `{"id":1}`,
		},
		{
			name: "existing overwritten",
			initial: map[string]Entry{
				"key": {0},
			},
			key:      "key",
			entry:    Entry{1},
			ttl:      time.Hour,
			expected: `{"id":1}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for key, entry := range test.initial {
				bs, err := json.Marshal(entry)
				require.NoError(t, err)

				err = server.Set(key, string(bs))
				require.NoError(t, err)
			}

			// Write the entry we care about.
			err := SetWithTTL(conn, test.key, test.entry, test.ttl)

			// Verify we wrote our entry.
			require.NoError(t, err)
			server.CheckGet(t, test.key, test.expected)
			assert.Equal(t, test.ttl, server.TTL(test.key))
		})
	}
}

func TestSetWithTTL_Error(t *testing.T) {
	tests := []struct {
		name       string
		connection ConnectionFunc
		key        string      // The key of the data to write.
		data       interface{} // The data to write.
		expected   error
	}{
		{
			name: "json.Marshal error",
			key:  "key",
			data: make(chan int), // Channels are not able to be marshalled to JSON.
		},
		{
			name: "conn.Do error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("forced error")
			},
			key:      "key",
			expected: errors.New("forced error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, conn := NewMiniredis(t)

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server.
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			err := SetWithTTL(connection, test.key, test.data, time.Hour)

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

func TestScanKeys(t *testing.T) {
	tests := []struct {
		name     string
		initial  []string // Keys that should be present for the test.
		key      string   // The key of the data to scan for.
		expected []string // The keys expected to be returned.
	}{
		{
			name: "no existing keys",
			key:  "key",
		},
		{
			name:    "no overlapping keys",
			initial: []string{"prefix2:key1", "prefix2:key2"},
			key:     "prefix1:*",
		},
		{
			name:     "one existing key",
			initial:  []string{"prefix1:key1", "prefix2:key1"},
			key:      "prefix1:*",
			expected: []string{"prefix1:key1"},
		},
		{
			name:     "multiple existing keys",
			initial:  []string{"prefix1:key2", "prefix1:key1", "prefix2:key1"},
			key:      "prefix1:*",
			expected: []string{"prefix1:key1", "prefix1:key2"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server, conn := NewMiniredis(t)

			// Write all of the initial entries into the database.
			for _, key := range test.initial {
				err := server.Set(key, "")
				require.NoError(t, err)
			}

			// Scan
			keys, err := ScanKeys(conn, test.key)

			// Verify we found the keys we were expecting.
			require.NoError(t, err)
			assert.Equal(t, test.expected, keys)
		})
	}
}

func TestScanKeys_Error(t *testing.T) {
	tests := []struct {
		name       string
		connection ConnectionFunc
		key        string // The key of the data to scan for.
		expected   error  // The error we expect to receive.
	}{
		{
			name: "conn.Do error",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				return nil, errors.New("forced error")
			},
			key:      "key",
			expected: errors.New("forced error"),
		},
		{
			name: "field 0 not an int64",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				var data []interface{}
				data = append(data, "not an int64")
				return data, nil
			},
		},
		{
			name: "field 1 not a []string",
			connection: func(command string, args ...interface{}) (interface{}, error) {
				var data []interface{}
				data = append(data, int64(0)) // cursor id, 0 means all data returned
				data = append(data, 0)        // should be []string
				return data, nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, conn := NewMiniredis(t)

			// If we weren't provided a connection to use, then use the one connected
			// to the miniredis server.
			var connection Connection = test.connection
			if test.connection == nil {
				connection = conn
			}

			_, err := ScanKeys(connection, test.key)

			// Verify we got the error we expected.
			assert.Error(t, err)
			if test.expected != nil {
				assert.Equal(t, test.expected, err)
			}
		})
	}
}

type ConnectionFunc func(command string, args ...interface{}) (interface{}, error)

func (cf ConnectionFunc) Do(command string, args ...interface{}) (interface{}, error) {
	return cf(command, args...)
}

func NewMiniredis(t *testing.T) (*miniredis.Miniredis, redis.Conn) {
	server, err := miniredis.Run()
	require.NoError(t, err)

	connection, err := redis.Dial("tcp", server.Addr())
	require.NoError(t, err)

	t.Cleanup(func() {
		connection.Close()
		server.Close()
	})

	return server, connection
}
