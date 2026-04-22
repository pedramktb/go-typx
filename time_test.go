package typx_test

import (
	"testing"
	"time"

	typx "github.com/pedramktb/go-typx"
	"github.com/stretchr/testify/require"
)

func dt(year int, month time.Month, day, hour, min int) typx.DateTime {
	return typx.DateTime{Time: time.Date(year, month, day, hour, min, 0, 0, time.UTC)}
}

func Test_DateTime_Add_TimeOnly(t *testing.T) {
	base := dt(2026, time.March, 15, 10, 0)
	result := base.AddDuration(typx.Duration{Time: 2 * time.Hour})
	require.Equal(t, dt(2026, time.March, 15, 12, 0), result)
}

func Test_DateTime_Add_DaysOnly(t *testing.T) {
	base := dt(2026, time.March, 15, 10, 0)
	result := base.AddDuration(typx.Duration{Day: 10})
	require.Equal(t, dt(2026, time.March, 25, 10, 0), result)
}

func Test_DateTime_Add_MonthsOnly(t *testing.T) {
	base := dt(2026, time.January, 15, 10, 0)
	result := base.AddDuration(typx.Duration{Month: 3})
	require.Equal(t, dt(2026, time.April, 15, 10, 0), result)
}

func Test_DateTime_Add_MonthsAcrossYear(t *testing.T) {
	base := dt(2025, time.November, 1, 0, 0)
	result := base.AddDuration(typx.Duration{Month: 3})
	require.Equal(t, dt(2026, time.February, 1, 0, 0), result)
}

func Test_DateTime_Add_Mixed(t *testing.T) {
	base := dt(2026, time.January, 10, 6, 0)
	d := typx.Duration{Month: 2, Day: 5, Time: 3 * time.Hour}
	result := base.AddDuration(d)
	// Add applies time.Duration first, then AddDate(0, months, days):
	// Jan 10 06:00 + 3h = Jan 10 09:00, + 2 months = Mar 10 09:00, + 5 days = Mar 15 09:00
	require.Equal(t, dt(2026, time.March, 15, 9, 0), result)
}

func Test_DateTime_Add_MonthOverflowDay(t *testing.T) {
	// Jan 31 + 1 month: Feb 31 overflows to Mar 3 (2026 is not a leap year)
	base := dt(2026, time.January, 31, 0, 0)
	result := base.AddDuration(typx.Duration{Month: 1})
	require.Equal(t, dt(2026, time.March, 3, 0, 0), result)
}

func Test_DateTime_Add_Zero(t *testing.T) {
	base := dt(2026, time.June, 15, 12, 30)
	result := base.AddDuration(typx.Duration{})
	require.Equal(t, base, result)
}

func Test_DateTime_Sub_RoundTrip_TimeOnly(t *testing.T) {
	t2 := dt(2026, time.March, 15, 10, 0)
	d := typx.Duration{Time: 5 * time.Hour}
	t1 := t2.AddDuration(d)
	require.Equal(t, t1, t2.AddDuration(t1.SubDateTime(t2)))
}

func Test_DateTime_Sub_RoundTrip_DaysOnly(t *testing.T) {
	t2 := dt(2026, time.March, 10, 0, 0)
	d := typx.Duration{Day: 7}
	t1 := t2.AddDuration(d)
	require.Equal(t, t1, t2.AddDuration(t1.SubDateTime(t2)))
}

func Test_DateTime_Sub_RoundTrip_MonthsOnly(t *testing.T) {
	t2 := dt(2026, time.January, 15, 0, 0)
	d := typx.Duration{Month: 4}
	t1 := t2.AddDuration(d)
	require.Equal(t, t1, t2.AddDuration(t1.SubDateTime(t2)))
}

func Test_DateTime_Sub_RoundTrip_Mixed(t *testing.T) {
	t2 := dt(2026, time.January, 10, 6, 0)
	d := typx.Duration{Month: 2, Day: 5, Time: 3 * time.Hour}
	t1 := t2.AddDuration(d)
	require.Equal(t, t1, t2.AddDuration(t1.SubDateTime(t2)))
}
