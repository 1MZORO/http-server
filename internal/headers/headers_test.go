package headers

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
// Test: Valid single header
headers := NewHeader()
data := []byte("Host: localhost:42069\r\n\r\n")
n, done, err := headers.Parse(data)
require.NoError(t, err)
require.NotNil(t, headers)
assert.Equal(t, "localhost:42069", headers.Get("Host"))
assert.Equal(t, 23, n)
assert.False(t, done)

// Test: Invalid spacing header
headers = NewHeader()
data = []byte("       Host : localhost:42069       \r\n\r\n")
n, done, err = headers.Parse(data)
require.Error(t, err)
assert.Equal(t, 0, n)
assert.False(t, done)

headers = NewHeader()
data = []byte("H©st: localhost:42069\r\n\r\n")
n, done, err = headers.Parse(data)
require.Error(t, err)
assert.Equal(t, 0, n)
assert.False(t, done)

headers = NewHeader()
data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\n\r\n")
n, done, err = headers.Parse(data)
require.NotNil(t, headers)
assert.Equal(t, "localhost:42069,localhost:42069", headers.Get("Host"))
assert.False(t, done)
}