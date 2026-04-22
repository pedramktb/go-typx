package typx_test

import (
	"encoding/json"
	"testing"

	typx "github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
)

func Test_URL_RoundTrip(t *testing.T) {
	raw := "https://example.com/path?query=1#fragment"
	u, err := typx.URLFrom(raw)
	require.NoError(t, err)
	require.Equal(t, raw, u.String())
}

func Test_URL_MarshalUnmarshalText(t *testing.T) {
	raw := "https://example.com/path?query=1#fragment"
	u, err := typx.URLFrom(raw)
	require.NoError(t, err)

	b, err := u.MarshalText()
	require.NoError(t, err)

	var u2 typx.URL
	err = u2.UnmarshalText(b)
	require.NoError(t, err)
	require.Equal(t, u.String(), u2.String())
}

func Test_URL_JSONMarshalUnmarshal(t *testing.T) {
	raw := "https://example.com/path?query=1"
	u, err := typx.URLFrom(raw)
	require.NoError(t, err)

	b, err := json.Marshal(u)
	require.NoError(t, err)
	require.Equal(t, `"`+raw+`"`, string(b))

	var u2 typx.URL
	err = json.Unmarshal(b, &u2)
	require.NoError(t, err)
	require.Equal(t, u.String(), u2.String())
}

func Test_URL_Scan_String(t *testing.T) {
	raw := "https://example.com/path"
	var u typx.URL
	err := u.Scan(raw)
	require.NoError(t, err)
	require.Equal(t, raw, u.String())
}

func Test_URL_Scan_Bytes(t *testing.T) {
	raw := "https://example.com/path"
	var u typx.URL
	err := u.Scan([]byte(raw))
	require.NoError(t, err)
	require.Equal(t, raw, u.String())
}

func Test_URL_Scan_Nil(t *testing.T) {
	var u typx.URL
	err := u.Scan(nil)
	require.NoError(t, err)
	require.Equal(t, "", u.String())
}

func Test_URL_Scan_UnsupportedType(t *testing.T) {
	var u typx.URL
	err := u.Scan(42)
	require.Error(t, err)
}

func Test_URL_Scan_InvalidURL(t *testing.T) {
	var u typx.URL
	err := u.Scan("://invalid%gg")
	require.Error(t, err)
}

func Test_URL_Value(t *testing.T) {
	raw := "https://example.com/path?query=1"
	u, err := typx.URLFrom(raw)
	require.NoError(t, err)

	val, err := u.Value()
	require.NoError(t, err)
	require.Equal(t, raw, val)
}

func Test_URL_Value_RoundTrip_Via_Scan(t *testing.T) {
	raw := "https://user:pass@example.com:8080/path?k=v#anchor"
	u, err := typx.URLFrom(raw)
	require.NoError(t, err)

	val, err := u.Value()
	require.NoError(t, err)

	var u2 typx.URL
	err = u2.Scan(val)
	require.NoError(t, err)
	require.Equal(t, u.String(), u2.String())
}
