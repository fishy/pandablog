package model

import (
	"strings"
	"time"
)

// Tag -
type Tag struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

func (left Tag) Compare(right Tag) int {
	switch l, r := strings.ToLower(left.Name), strings.ToLower(right.Name); {
	default:
		return 0
	case l < r:
		return -1
	case l > r:
		return 1
	}
}
