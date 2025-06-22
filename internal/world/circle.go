package world

// CircleIntersectsCircle checks if two circles intersect.
func CircleIntersectsCircle(x1, y1, r1, x2, y2, r2 float64) bool {
	dx := x2 - x1
	dy := y2 - y1
	distanceSquared := dx*dx + dy*dy
	radiiSum := r1 + r2
	return distanceSquared <= radiiSum*radiiSum
}

func CircleIntersectsBox(cx, cy, cr, bx, by, bw, bh float64) bool {
	closestX := clamp64(cx, bx, bx+bw)
	closestY := clamp64(cy, by, by+bh)

	dx := cx - closestX
	dy := cy - closestY
	return (dx*dx + dy*dy) <= (cr * cr)
}

func clamp64(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
