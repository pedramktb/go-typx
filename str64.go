package typx

import (
	"encoding"
	"encoding/base64"
	"fmt"
)

/*
[Str64] is a named []byte type that encodes arbitrary byte slices as unpadded base64url strings
(alphabet: A-Z, a-z, 0-9, -, _) via [base64.RawURLEncoding]. Because it implements
[encoding.TextMarshaler] / [encoding.TextUnmarshaler], it serialises automatically in
JSON, and any other text-based transport or storage without extra conversion,
while having the possibility of being saved as raw bytes, such as in Redis or a database.

Typical use cases:

 1. Shorter UUID representation.
    A UUID is 16 bytes. As a standard hyphenated string it is 36 characters;
    as [Str64] it is 22 characters — a 39 % reduction — while still being URL-safe.

 2. Custom Variable-length identifiers.
    When a UUID's 128 bits of entropy is more (or less) than needed, [Str64] works
    for any byte length: 8 bytes (64-bit) encodes to 11 chars, 32 bytes to 43 chars.

 3. Compact string storage.
    If a string's content is limited to the [Str64] alphabet (A-Z, a-z, 0-9, -, _),
    storing it as [Str64] bytes saves 25 % of space (6 bits stored per byte instead of 8).

 4. Loggable binary values.
    Any []byte field (hashes, tokens, raw IDs) becomes a human-readable, URL- and
    filename-safe string automatically in logs, JSON responses, and etc.
*/
type Str64 []byte

// [Str64From] decodes a base64 RawURL string back to the original raw bytes.
func Str64From(s string) (Str64, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return Str64(b), nil
}

var _ fmt.Stringer = Str64{}

// String encodes the raw bytes to their base64 RawURL text representation.
func (s Str64) String() string {
	return base64.RawURLEncoding.EncodeToString(s)
}

var _ encoding.TextMarshaler = Str64{}

// MarshalText implements [encoding.TextMarshaler].
func (s Str64) MarshalText() ([]byte, error) {
	out := make([]byte, base64.RawURLEncoding.EncodedLen(len(s)))
	base64.RawURLEncoding.Encode(out, s)
	return out, nil
}

var _ encoding.TextUnmarshaler = (*Str64)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (s *Str64) UnmarshalText(text []byte) error {
	val, err := Str64From(string(text))
	if err != nil {
		return err
	}
	*s = val
	return nil
}
