package typx_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
)

func Test_Str64_RoundTrip(t *testing.T) {
	uid := uuid.New()

	// Cast raw bytes directly — no encoding step needed.
	s := typx.Str64(uid[:])
	require.Len(t, s, 16)

	// 16 bytes → ceil(16*8/6) chars per block logic → 8+8+6 = 22 chars total.
	str := s.String()
	require.Len(t, str, 22)

	// FromString recovers exactly the original 16 bytes — no slicing required.
	s2, err := typx.FromString(str)
	require.NoError(t, err)
	require.Equal(t, s, s2)

	uid2, err := uuid.FromBytes(s2)
	require.NoError(t, err)
	require.Equal(t, uid, uid2)
}

func Test_Str64_MarshalUnmarshalText(t *testing.T) {
	uid := uuid.New()
	original := typx.Str64(uid[:])

	text, err := original.MarshalText()
	require.NoError(t, err)
	require.Len(t, text, 22)

	var restored typx.Str64
	err = restored.UnmarshalText(text)
	require.NoError(t, err)
	require.Equal(t, original, restored)
}

func Test_Str64_UnsupportedCharacter(t *testing.T) {
	_, err := typx.FromString("hello!")
	require.Error(t, err)
}
