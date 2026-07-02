package ical

import (
	"strings"
	"time"
)

// tzInfo holds everything needed to render times in a supported IANA zone
// without depending on the host's tzdata: a fixed-offset location and the
// matching VTIMEZONE component body.
type tzInfo struct {
	location  *time.Location
	vtimezone string // LF-separated VTIMEZONE lines (folded and CRLF-joined later)
}

// builtinZones maps IANA zone ids to their rendering data. The system operates
// in a single fixed-offset zone (Europe/Moscow, which has observed no DST since
// 2014); adding an entry here exposes another zone to the feed. Zones absent
// from this map fall back to UTC rendering.
var builtinZones = map[string]tzInfo{
	"Europe/Moscow": {
		location: time.FixedZone("MSK", 3*3600),
		vtimezone: strings.Join([]string{
			"BEGIN:VTIMEZONE",
			"TZID:Europe/Moscow",
			"BEGIN:STANDARD",
			"DTSTART:19700101T000000",
			"TZOFFSETFROM:+0300",
			"TZOFFSETTO:+0300",
			"TZNAME:MSK",
			"END:STANDARD",
			"END:VTIMEZONE",
		}, "\n"),
	},
}

// lookupZone returns the rendering data for an IANA zone id and whether it is
// known. An empty or unknown id falls back to UTC rendering (ok == false).
func lookupZone(tzid string) (tzInfo, bool) {
	info, ok := builtinZones[tzid]
	return info, ok
}

// Location returns the *time.Location for a builtin zone id, or UTC for an
// unknown id. Callers building event times for Render must use this location so
// their wall-clock times match the zone the renderer emits.
func Location(tzid string) *time.Location {
	if info, ok := builtinZones[tzid]; ok {
		return info.location
	}
	return time.UTC
}
