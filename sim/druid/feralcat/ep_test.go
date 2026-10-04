//go:build with_db

package feralcat

import (
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Prints DPS EP for the default cat: Level 60 gear, Rip at 5, Rake kept, normalized to agility.
// Pass the UI's iteration count; stat weights then split it above and below the baseline.
func TestDefaultEp(t *testing.T) {
	if os.Getenv("EP_SWEEP") == "" {
		t.Skip("set EP_SWEEP=1 (and EP_SWEEP_ITERATIONS) to compute default EP")
	}
	iterations := int32(12500)
	if n, err := strconv.Atoi(os.Getenv("EP_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	base := sweepRequest(powershiftTalents, defaultCatRotation(), false, iterations)
	// IsTest caps the sim at 3 workers. Labeled RNG is turned on inside stat weights.
	base.SimOptions.IsTest = false
	weigh := []proto.Stat{
		proto.Stat_StatStrength,
		proto.Stat_StatAgility,
		proto.Stat_StatAttackPower,
		proto.Stat_StatFeralAttackPower,
		proto.Stat_StatPhysicalDamage,
		proto.Stat_StatHitRating,
		proto.Stat_StatExpertiseRating,
		proto.Stat_StatCritRating,
		proto.Stat_StatHasteRating,
		proto.Stat_StatArmorPenetration,
	}
	result := core.StatWeights(&proto.StatWeightsRequest{
		Player:          base.Raid.Parties[0].Players[0],
		RaidBuffs:       base.Raid.Buffs,
		PartyBuffs:      base.Raid.Parties[0].Buffs,
		Debuffs:         base.Raid.Debuffs,
		Encounter:       base.Encounter,
		SimOptions:      base.SimOptions,
		StatsToWeigh:    weigh,
		EpReferenceStat: proto.Stat_StatAgility,
	})
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	fmt.Printf("%-22s %12s %12s %10s %10s\n", "stat", "dps/point", "stdev", "ep", "ep stdev")
	for _, s := range weigh {
		i := int(s)
		fmt.Printf("%-22s %12.4f %12.4f %10.4f %10.4f\n",
			stats.Stat(s).StatName(),
			result.Dps.Weights.Stats[i], result.Dps.WeightsStdev.Stats[i],
			result.Dps.EpValues.Stats[i], result.Dps.EpValuesStdev.Stats[i])
	}
}
