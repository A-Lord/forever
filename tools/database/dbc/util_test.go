package dbc

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func TestGetRacesFromRaceMask(t *testing.T) {
	for _, tc := range []struct {
		mask int64
		want []proto.Race
	}{
		{1<<1 | 1<<9, []proto.Race{proto.Race_RaceBloodElf, proto.Race_RaceOrc}},
		{1<<10 | 1<<0, []proto.Race{proto.Race_RaceDraenei, proto.Race_RaceHuman}},
		{-1, nil},
		{0b110_1111_1111, nil},
	} {
		if got := GetRacesFromRaceMask(tc.mask); !slices.Equal(got, tc.want) {
			t.Errorf("mask %#x: %v, want %v", tc.mask, got, tc.want)
		}
	}
}
