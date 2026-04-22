package typx_test

import (
	"encoding/json"
	"testing"
	"time"

	typx "github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
)

func Test_Duration_String(t *testing.T) {
	tests := []struct {
		name string
		d    typx.Duration
		want string
	}{
		{
			name: "zero",
			d:    typx.Duration{},
			want: "0M0D0s",
		},
		{
			name: "months only",
			d:    typx.Duration{Month: 3},
			want: "3M0D0s",
		},
		{
			name: "days only",
			d:    typx.Duration{Day: 5},
			want: "0M5D0s",
		},
		{
			name: "time only",
			d:    typx.Duration{Time: 2*time.Hour + 30*time.Minute},
			want: "0M0D2h30m0s",
		},
		{
			name: "all components",
			d:    typx.Duration{Month: 2, Day: 10, Time: time.Hour},
			want: "2M10D1h0m0s",
		},
		{
			name: "negative",
			d:    typx.Duration{Month: -1, Day: -3, Time: -time.Hour},
			want: "-1M-3D-1h0m0s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.String())
		})
	}
}

func Test_Duration_Abs(t *testing.T) {
	tests := []struct {
		name string
		d    typx.Duration
		want typx.Duration
	}{
		{
			name: "zero unchanged",
			d:    typx.Duration{},
			want: typx.Duration{},
		},
		{
			name: "positive unchanged",
			d:    typx.Duration{Month: 2, Day: 10, Time: time.Hour},
			want: typx.Duration{Month: 2, Day: 10, Time: time.Hour},
		},
		{
			name: "all negative inverted",
			d:    typx.Duration{Month: -2, Day: -10, Time: -time.Hour},
			want: typx.Duration{Month: 2, Day: 10, Time: time.Hour},
		},
		{
			name: "mixed signs",
			d:    typx.Duration{Month: -3, Day: 4, Time: -time.Minute},
			want: typx.Duration{Month: 3, Day: 4, Time: time.Minute},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.Abs())
		})
	}
}

func Test_Duration_Add(t *testing.T) {
	tests := []struct {
		name  string
		d     typx.Duration
		other typx.Duration
		want  typx.Duration
	}{
		{
			name:  "zero + zero",
			d:     typx.Duration{},
			other: typx.Duration{},
			want:  typx.Duration{},
		},
		{
			name:  "all components",
			d:     typx.Duration{Month: 1, Day: 2, Time: time.Hour},
			other: typx.Duration{Month: 3, Day: 4, Time: 2 * time.Hour},
			want:  typx.Duration{Month: 4, Day: 6, Time: 3 * time.Hour},
		},
		{
			name:  "negative other",
			d:     typx.Duration{Month: 5, Day: 10, Time: 3 * time.Hour},
			other: typx.Duration{Month: -2, Day: -4, Time: -time.Hour},
			want:  typx.Duration{Month: 3, Day: 6, Time: 2 * time.Hour},
		},
		{
			name:  "add zero",
			d:     typx.Duration{Month: 1, Day: 2, Time: time.Hour},
			other: typx.Duration{},
			want:  typx.Duration{Month: 1, Day: 2, Time: time.Hour},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.Add(tt.other))
		})
	}
}

func Test_Duration_Sub(t *testing.T) {
	tests := []struct {
		name  string
		d     typx.Duration
		other typx.Duration
		want  typx.Duration
	}{
		{
			name:  "zero - zero",
			d:     typx.Duration{},
			other: typx.Duration{},
			want:  typx.Duration{},
		},
		{
			name:  "all components",
			d:     typx.Duration{Month: 5, Day: 10, Time: 3 * time.Hour},
			other: typx.Duration{Month: 2, Day: 4, Time: time.Hour},
			want:  typx.Duration{Month: 3, Day: 6, Time: 2 * time.Hour},
		},
		{
			name:  "resulting in negatives",
			d:     typx.Duration{Month: 1, Day: 2, Time: time.Hour},
			other: typx.Duration{Month: 3, Day: 4, Time: 2 * time.Hour},
			want:  typx.Duration{Month: -2, Day: -2, Time: -time.Hour},
		},
		{
			name:  "self subtraction",
			d:     typx.Duration{Month: 3, Day: 7, Time: 5 * time.Hour},
			other: typx.Duration{Month: 3, Day: 7, Time: 5 * time.Hour},
			want:  typx.Duration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.d.Sub(tt.other))
		})
	}
}

func Test_Duration_MarshalText(t *testing.T) {
	tests := []struct {
		name string
		d    typx.Duration
		want string
	}{
		{
			name: "zero",
			d:    typx.Duration{},
			want: "PT0S",
		},
		{
			name: "months only — less than a year",
			d:    typx.Duration{Month: 3},
			want: "P3M",
		},
		{
			name: "months only — more than a year",
			d:    typx.Duration{Month: 14},
			want: "P1Y2M",
		},
		{
			name: "months only — exact years",
			d:    typx.Duration{Month: 24},
			want: "P2Y",
		},
		{
			name: "days only",
			d:    typx.Duration{Day: 5},
			want: "P5D",
		},
		{
			name: "hours only",
			d:    typx.Duration{Time: 4 * time.Hour},
			want: "PT4H",
		},
		{
			name: "minutes only",
			d:    typx.Duration{Time: 30 * time.Minute},
			want: "PT30M",
		},
		{
			name: "whole seconds only",
			d:    typx.Duration{Time: 45 * time.Second},
			want: "PT45S",
		},
		{
			name: "fractional seconds",
			d:    typx.Duration{Time: 1500 * time.Millisecond},
			want: "PT1.5S",
		},
		{
			name: "nanosecond precision",
			d:    typx.Duration{Time: time.Second + 123456789},
			want: "PT1.123456789S",
		},
		{
			name: "hours minutes seconds",
			d:    typx.Duration{Time: 4*time.Hour + 5*time.Minute + 6*time.Second},
			want: "PT4H5M6S",
		},
		{
			name: "all components",
			d:    typx.Duration{Month: 14, Day: 3, Time: 4*time.Hour + 5*time.Minute + 6*time.Second},
			want: "P1Y2M3DT4H5M6S",
		},
		{
			name: "negative months",
			d:    typx.Duration{Month: -14},
			want: "P-1Y-2M",
		},
		{
			name: "negative days",
			d:    typx.Duration{Day: -7},
			want: "P-7D",
		},
		{
			name: "negative time",
			d:    typx.Duration{Time: -(90*time.Minute + 30*time.Second)},
			want: "PT-1H-30M-30S",
		},
		{
			name: "negative fractional seconds",
			d:    typx.Duration{Time: -1500 * time.Millisecond},
			want: "PT-1.5S",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := tt.d.MarshalText()
			require.NoError(t, err)
			require.Equal(t, tt.want, string(b))
		})
	}
}

func Test_Duration_UnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    typx.Duration
		wantErr bool
	}{
		{
			name:  "zero PT0S",
			input: "PT0S",
			want:  typx.Duration{},
		},
		{
			name:  "zero P0D",
			input: "P0D",
			want:  typx.Duration{},
		},
		{
			name:  "months less than year",
			input: "P3M",
			want:  typx.Duration{Month: 3},
		},
		{
			name:  "years and months",
			input: "P1Y2M",
			want:  typx.Duration{Month: 14},
		},
		{
			name:  "years only",
			input: "P2Y",
			want:  typx.Duration{Month: 24},
		},
		{
			name:  "days only",
			input: "P5D",
			want:  typx.Duration{Day: 5},
		},
		{
			name:  "hours only",
			input: "PT4H",
			want:  typx.Duration{Time: 4 * time.Hour},
		},
		{
			name:  "minutes only",
			input: "PT30M",
			want:  typx.Duration{Time: 30 * time.Minute},
		},
		{
			name:  "whole seconds",
			input: "PT45S",
			want:  typx.Duration{Time: 45 * time.Second},
		},
		{
			name:  "fractional seconds",
			input: "PT1.5S",
			want:  typx.Duration{Time: 1500 * time.Millisecond},
		},
		{
			name:  "all components",
			input: "P1Y2M3DT4H5M6S",
			want:  typx.Duration{Month: 14, Day: 3, Time: 4*time.Hour + 5*time.Minute + 6*time.Second},
		},
		{
			name:  "negative prefix",
			input: "-P1Y2M3DT4H5M6S",
			want:  typx.Duration{Month: -14, Day: -3, Time: -(4*time.Hour + 5*time.Minute + 6*time.Second)},
		},
		{
			name:  "individual negative components",
			input: "P-1Y-2M",
			want:  typx.Duration{Month: -14},
		},
		{
			name:    "invalid format",
			input:   "1Y2M3D",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got typx.Duration
			err := got.UnmarshalText([]byte(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_Duration_MarshalText_RoundTrip(t *testing.T) {
	tests := []typx.Duration{
		{},
		{Month: 14, Day: 3, Time: 4*time.Hour + 5*time.Minute + 6*time.Second},
		{Month: -14, Day: -3, Time: -(4*time.Hour + 5*time.Minute + 6*time.Second)},
		{Time: 1500 * time.Millisecond},
		{Time: time.Second + 123456789},
		{Month: 24},
		{Day: -7},
	}
	for _, d := range tests {
		b, err := d.MarshalText()
		require.NoError(t, err)
		var got typx.Duration
		require.NoError(t, got.UnmarshalText(b))
		require.Equal(t, d, got)
	}
}

func Test_Duration_JSON_RoundTrip(t *testing.T) {
	d := typx.Duration{Month: 14, Day: 3, Time: 4*time.Hour + 5*time.Minute + 6*time.Second}
	b, err := json.Marshal(d)
	require.NoError(t, err)
	require.Equal(t, `"P1Y2M3DT4H5M6S"`, string(b))

	var got typx.Duration
	require.NoError(t, json.Unmarshal(b, &got))
	require.Equal(t, d, got)
}
