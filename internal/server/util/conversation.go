package util

import "github.com/google/uuid"

// DirectKey returns the same private-conversation key regardless of user order.
func DirectKey(a, b uuid.UUID) string {
	x, y := a.String(), b.String()
	if x > y {
		x, y = y, x
	}
	return x + ":" + y
}
