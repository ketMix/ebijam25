package game

import (
	"image/color"
	"strconv"
)

func stringToColor(s string, fallback color.NRGBA) color.NRGBA {
	if s == "" {
		return fallback
	}
	if s[0] == '#' {
		if len(s) == 4 { // Allow lazy RGB->RRGGBB
			rr := string(s[1] + s[1])
			gg := string(s[2] + s[2])
			bb := string(s[3] + s[3])
			s = "#" + rr + gg + bb
		}
		if len(s) == 7 {
			r, _ := strconv.ParseInt(s[1:3], 16, 0)
			g, _ := strconv.ParseInt(s[3:5], 16, 0)
			b, _ := strconv.ParseInt(s[5:7], 16, 0)
			return color.NRGBA{uint8(r), uint8(g), uint8(b), 255}
		} else if len(s) == 9 {
			r, _ := strconv.ParseInt(s[1:3], 16, 0)
			g, _ := strconv.ParseInt(s[3:5], 16, 0)
			b, _ := strconv.ParseInt(s[5:7], 16, 0)
			a, _ := strconv.ParseInt(s[7:9], 16, 0)
			return color.NRGBA{uint8(r), uint8(g), uint8(b), uint8(a)}
		}
	}
	// Simple crummy name parsing.
	switch s {
	case "black":
		return color.NRGBA{0, 0, 0, 255}
	case "white":
		return color.NRGBA{255, 255, 255, 255}
	case "red":
		return color.NRGBA{255, 0, 0, 255}
	case "green":
		return color.NRGBA{0, 255, 0, 255}
	case "blue":
		return color.NRGBA{0, 0, 255, 255}
	}
	return color.NRGBA{0, 0, 0, 255}
}
