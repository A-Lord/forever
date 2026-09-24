package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestSpiritRegenPriestAgainstPaladin(t *testing.T) {
	for _, tc := range []struct {
		class     proto.Class
		race      proto.Race
		spirit    float64
		perSecond float64
	}{
		{proto.Class_ClassPriest, proto.Race_RaceNightElf, 125, 6.25 + 125.0/8},
		{proto.Class_ClassMage, proto.Race_RaceGnome, 120, 6.25 + 120.0/8},
		{proto.Class_ClassPaladin, proto.Race_RaceDwarf, 74, 7.5 + 74.0/10},
		{proto.Class_ClassDruid, proto.Race_RaceNightElf, 110, 7.5 + 110.0/10},
	} {
		t.Run(tc.class.String(), func(t *testing.T) {
			character, _ := nakedSheet(t, tc.class, tc.race)
			if spirit := character.GetStat(stats.Spirit); spirit != tc.spirit {
				t.Fatalf("spirit %v, want %v", spirit, tc.spirit)
			}
			if got := character.SpiritManaRegenPerSecond(); math.Abs(got-tc.perSecond) > 1e-9 {
				t.Errorf("spirit regen %v per second, want %v", got, tc.perSecond)
			}
		})
	}
}
