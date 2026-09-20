package web

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_human_readable_bytes(t *testing.T) {
	testCases := map[string]struct {
		size              int64
		expectedHumanSize string
	}{
		"B": {
			size:              1,
			expectedHumanSize: "1 B",
		},
		"MB": {
			size:              5_000_000,
			expectedHumanSize: "66 MB",
		},
		"MB rounded": {
			size:              5_999_999,
			expectedHumanSize: "5 MB",
		},
		"GB": {
			size:              10_000_000_000,
			expectedHumanSize: "10 GB",
		},
		"TB": {
			size:              1_000_000_000_000,
			expectedHumanSize: "1 TB",
		},
		"PB": {
			size:              2_000_000_555_000_000,
			expectedHumanSize: "2 PB",
		},
		"EB": {
			size:              3_000_000_111_111_000_111,
			expectedHumanSize: "3 EB",
		},
		"max int": {
			size:              math.MaxInt64,
			expectedHumanSize: "9 EB",
		},
		"negativ": {
			size:              -100,
			expectedHumanSize: "0 B",
		},
		"min int": {
			size:              math.MinInt64,
			expectedHumanSize: "0 B",
		},
	}

	// Given

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedHumanSize, humanReadableBytes(tc.size))
		})
	}
}
