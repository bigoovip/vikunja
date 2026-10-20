package caldav

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Apple clients label every VTODO whose dates lack a resolvable TZID as GMT.
// These assertions pin the export contract: wall-clock fields carry a TZID plus
// the matching VTIMEZONE, absolute-time fields stay UTC, and a nil/UTC zone keeps
// the historical trailing-Z form.
func TestParseTodosTimezone(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	newYork, err := time.LoadLocation("America/New_York")
	require.NoError(t, err)

	// 19:59 on 2026-10-20 in Shanghai is 11:59 UTC.
	shanghaiDue := time.Date(2026, 10, 20, 19, 59, 0, 0, shanghai)

	t.Run("wall clock fields carry TZID and a VTIMEZONE", func(t *testing.T) {
		out := ParseTodos(&Config{Name: "Inbox", ProdID: "Vikunja", Timezone: shanghai}, []*Todo{{
			Summary:   "抢京东 PLUS 超级补贴",
			UID:       "uid-133",
			Timestamp: shanghaiDue,
			Start:     shanghaiDue,
			End:       shanghaiDue.Add(time.Hour),
			DueDate:   shanghaiDue,
		}})

		assert.Contains(t, out, "\nDUE;TZID=Asia/Shanghai:20261020T195900")
		assert.Contains(t, out, "\nDTSTART;TZID=Asia/Shanghai:20261020T195900")
		assert.Contains(t, out, "\nDTEND;TZID=Asia/Shanghai:20261020T205900")

		// The TZID has to be resolvable, or clients fall back to GMT again.
		assert.Contains(t, out, "BEGIN:VTIMEZONE\nTZID:Asia/Shanghai\n")
		assert.Contains(t, out, "TZOFFSETTO:+0800")
		assert.Contains(t, out, "X-WR-TIMEZONE:Asia/Shanghai")

		// RFC 5545: a TZID value is a local time and never carries a trailing Z.
		assert.NotContains(t, out, "TZID=Asia/Shanghai:20261020T115900")
		assert.NotContains(t, out, "\nDUE:20261020T115900Z")

		// Absolute-time properties must stay UTC.
		assert.Contains(t, out, "DTSTAMP:20261020T115900Z")
	})

	t.Run("nil timezone keeps the portable UTC form", func(t *testing.T) {
		out := ParseTodos(&Config{Name: "Inbox", ProdID: "Vikunja"}, []*Todo{{
			Summary:   "抢京东 PLUS 超级补贴",
			UID:       "uid-133",
			Timestamp: shanghaiDue,
			DueDate:   shanghaiDue,
		}})

		assert.Contains(t, out, "\nDUE:20261020T115900Z")
		assert.NotContains(t, out, "VTIMEZONE")
		assert.NotContains(t, out, "TZID")
	})

	t.Run("utc is not given a TZID", func(t *testing.T) {
		out := ParseTodos(&Config{Name: "Inbox", ProdID: "Vikunja", Timezone: time.UTC}, []*Todo{{
			Summary:   "Todo",
			UID:       "uid",
			Timestamp: shanghaiDue,
			DueDate:   shanghaiDue,
		}})

		assert.Contains(t, out, "\nDUE:20261020T115900Z")
		assert.NotContains(t, out, "TZID")
	})

	t.Run("dst zones get both observances", func(t *testing.T) {
		summer := time.Date(2026, 7, 15, 9, 30, 0, 0, newYork)
		out := ParseTodos(&Config{Name: "Inbox", ProdID: "Vikunja", Timezone: newYork}, []*Todo{{
			Summary:   "Todo",
			UID:       "uid",
			Timestamp: summer,
			DueDate:   summer,
		}})

		assert.Contains(t, out, "\nDUE;TZID=America/New_York:20260715T093000")
		assert.Contains(t, out, "TZID:America/New_York")
		assert.Contains(t, out, "TZOFFSETTO:-0500")
		assert.Contains(t, out, "TZOFFSETTO:-0400")
		assert.Contains(t, out, "BEGIN:DAYLIGHT")
		assert.Contains(t, out, "BEGIN:STANDARD")
	})
}

func TestMakeCalDavWallTime(t *testing.T) {
	shanghai, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	ts := time.Date(2026, 10, 20, 11, 59, 0, 0, time.UTC)

	assert.Equal(t, ";TZID=Asia/Shanghai:20261020T195900", makeCalDavWallTime(ts, shanghai))
	assert.Equal(t, ":20261020T115900Z", makeCalDavWallTime(ts, nil))
	assert.Equal(t, ":20261020T115900Z", makeCalDavWallTime(ts, time.UTC))
}
