package world

// CircleIntersectsCircle checks if two circles intersect.
func CircleIntersectsCircle(x1, y1, r1, x2, y2, r2 float64) bool {
	dx := x2 - x1
	dy := y2 - y1
	distanceSquared := dx*dx + dy*dy
	radiiSum := r1 + r2
	return distanceSquared <= radiiSum*radiiSum
}
