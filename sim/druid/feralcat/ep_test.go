//go:build with_db

package feralcat

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

func TestDefaultEp(t *testing.T) {
	if os.Getenv("EP_SWEEP") == "" {
		t.Skip("set EP_SWEEP=1 (and EP_SWEEP_ITERATIONS) to compute default EP")
	}
	iterations := int32(12500)
	if n, err := strconv.Atoi(os.Getenv("EP_SWEEP_ITERATIONS")); err == nil && n > 0 {
		iterations = int32(n)
	}
	base := buildRequest(t, "raid")
	if input := os.Getenv("EP_SWEEP_REQUEST"); input != "" {
		data, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := protojson.Unmarshal(data, base); err != nil {
			t.Fatal(err)
		}
	}
	base.SimOptions.Iterations = iterations
	base.SimOptions.RandomSeed = 101
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
	request := &proto.StatWeightsRequest{
		Player:             base.Raid.Parties[0].Players[0],
		RaidBuffs:          base.Raid.Buffs,
		PartyBuffs:         base.Raid.Parties[0].Buffs,
		Debuffs:            base.Raid.Debuffs,
		Encounter:          base.Encounter,
		SimOptions:         base.SimOptions,
		StatsToWeigh:       weigh,
		PseudoStatsToWeigh: []proto.PseudoStat{proto.PseudoStat_PseudoStatMainHandDps},
		EpReferenceStat:    proto.Stat_StatAgility,
	}
	writeCatEvidence(t, "ep-request", request)
	result := core.StatWeights(request)
	writeCatEvidence(t, "ep-result", result)
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
	i := int(proto.PseudoStat_PseudoStatMainHandDps)
	fmt.Printf("%-22s %12.4f %12.4f %10.4f %10.4f\n", "Main Hand DPS",
		result.Dps.Weights.PseudoStats[i], result.Dps.WeightsStdev.PseudoStats[i],
		result.Dps.EpValues.PseudoStats[i], result.Dps.EpValuesStdev.PseudoStats[i])
}

func writeCatEvidence(t testing.TB, name string, message goproto.Message) {
	t.Helper()
	dir := os.Getenv("CAT_EVIDENCE_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	data, err := (protojson.MarshalOptions{Indent: "  "}).Marshal(message)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), data, 0644); err != nil {
		t.Fatal(err)
	}
}
