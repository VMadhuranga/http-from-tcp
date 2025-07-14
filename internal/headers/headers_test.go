package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte(" Host: localhost:42069 \r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 25, n)
	assert.False(t, done)

	// Test: Valid single header with existing headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nUser-Agent: curl/8.11.1\r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Valid single header with multiple values
	headers = NewHeaders()
	data = []byte("Set-Person: lane-loves-go\r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go", headers["set-person"])
	assert.Equal(t, 27, n)
	assert.False(t, done)

	data = []byte("Set-Person: prime-loves-zig\r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go, prime-loves-zig", headers["set-person"])
	assert.Equal(t, 29, n)
	assert.False(t, done)

	data = []byte("Set-Person: tj-loves-ocaml\r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "lane-loves-go, prime-loves-zig, tj-loves-ocaml", headers["set-person"])
	assert.Equal(t, 28, n)
	assert.False(t, done)

	// Test: Valid done
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(string(data))
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header key character
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(string(data))
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
