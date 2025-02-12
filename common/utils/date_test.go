package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDate(t *testing.T) {
	// Fixed test time for consistent testing
	fixedTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)

	t.Run("ToISOString", func(t *testing.T) {
		isoString := Date.ToISOString(fixedTime)
		expected := "2024-03-15T14:30:00.000Z"
		assert.Equal(t, expected, isoString)
	})

	t.Run("Parse", func(t *testing.T) {
		parsed, err := Date.Parse("2024-03-15T14:30:00.000Z")
		assert.NoError(t, err)
		assert.Equal(t, fixedTime, parsed)

		_, err = Date.Parse("invalid")
		assert.Error(t, err)
	})

	t.Run("FromIsoString", func(t *testing.T) {
		timestamp := Date.FromIsoString("2024-03-15T14:30:00.000Z")
		assert.Equal(t, fixedTime.Unix(), timestamp)

		invalidTimestamp := Date.FromIsoString("invalid")
		assert.Equal(t, int64(0), invalidTimestamp)
	})

	t.Run("ToUnix", func(t *testing.T) {
		unixTime := fixedTime.Unix()
		converted := Date.ToUnix(unixTime)
		assert.Equal(t, fixedTime.Unix(), converted.Unix())
	})

	t.Run("CurrentTimeStampSecond", func(t *testing.T) {
		now := time.Now()
		timestamp := Date.CurrentTimeStampSecond()
		assert.InDelta(t, now.Unix(), timestamp, 1)
	})

	t.Run("CurrentDate", func(t *testing.T) {
		weekday := Date.CurrentDate()
		validDays := map[string]bool{
			"sunday": true, "monday": true, "tuesday": true,
			"wednesday": true, "thursday": true,
			"friday": true, "saturday": true,
		}
		assert.True(t, validDays[weekday])
	})

	t.Run("NextDayAtHour", func(t *testing.T) {
		hour := 10
		nextDay := Date.NextDayAtHour(hour)
		tomorrow := time.Unix(nextDay, 0)

		assert.Equal(t, hour, tomorrow.UTC().Hour())
		assert.Equal(t, 0, tomorrow.UTC().Minute())
		assert.Equal(t, 0, tomorrow.UTC().Second())
	})

	t.Run("NextWeekAtHour", func(t *testing.T) {
		hour := 10
		nextWeek := Date.NextWeekAtHour(hour)
		nextWeekTime := time.Unix(nextWeek, 0)

		assert.Equal(t, hour, nextWeekTime.UTC().Hour())
		assert.Equal(t, 0, nextWeekTime.UTC().Minute())
		assert.Equal(t, 0, nextWeekTime.UTC().Second())

		// Ensure it's next week
		now := time.Now()
		assert.Greater(t, nextWeekTime.Unix(), now.AddDate(0, 0, 7).Unix())
	})
}
