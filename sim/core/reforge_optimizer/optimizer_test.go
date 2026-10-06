//go:build with_db

package reforgeoptimizer

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	protopkg "google.golang.org/protobuf/proto"
)

func enchantRequest(raid *proto.Raid, weights map[proto.Stat]float64, options ...*proto.BulkEnchantSlotOptions) *proto.ReforgeOptimizeRequest {
	ep := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name)), PseudoStats: make([]float64, len(proto.PseudoStat_name))}
	for stat, w := range weights {
		ep.Stats[stat] = w
	}
	return &proto.ReforgeOptimizeRequest{Raid: raid, PreCapEpWeights: ep, Settings: &proto.ReforgeSettings{}, EnchantOptions: options}
}

func wrist(ids ...int32) *proto.BulkEnchantSlotOptions {
	return &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotWrist, EffectIds: ids}
}

// Makes the player's current Agility an upper bound, so any Agility on top
// of it breaks the cap. A cap the request does not list as an undershoot cap is a target instead.
func capAgilityAtTheSheet(t *testing.T, request *proto.ReforgeOptimizeRequest) {
	t.Helper()
	capAgilityBelowTheSheet(t, request, 0)
}

func capAgilityBelowTheSheet(t *testing.T, request *proto.ReforgeOptimizeRequest, below float64) {
	t.Helper()
	caps := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name))}
	caps.Stats[proto.Stat_StatAgility] = sheetStats(t, request.Raid).Stats[stats.Agility] - below
	request.Settings.StatCaps = caps
	upperBound := &proto.UnitStats{Stats: make([]float64, len(proto.Stat_name))}
	upperBound.Stats[proto.Stat_StatAgility] = 1
	request.UndershootCaps = upperBound
}

func optimize(t *testing.T, request *proto.ReforgeOptimizeRequest) *proto.ReforgeOptimizeResult {
	t.Helper()
	result := OptimizeAsync(request, simsignals.CreateSignals())
	if err := result.GetError(); err != nil {
		t.Fatalf("OptimizeAsync: %s", err.GetMessage())
	}
	return result
}

func slotEnchant(result *proto.ReforgeOptimizeResult, slot proto.ItemSlot) int32 {
	return result.GetOptimizedGear().Items[slot].GetEnchant()
}

var agilityOverStrength = map[proto.Stat]float64{proto.Stat_StatAgility: 2, proto.Stat_StatStrength: 1}

func TestOptimizePicksTheHigherEPEnchant(t *testing.T) {
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(927, 1887))
	result := optimize(t, request)
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != 1887 {
		t.Errorf("wrist enchant %d, want 1887 (Greater Agility)", got)
	}
	if result.GetScore() <= 0 {
		t.Errorf("score %v, want above 0", result.GetScore())
	}
	if again := optimize(t, request); !protopkg.Equal(again.GetOptimizedGear(), result.GetOptimizedGear()) {
		t.Errorf("second run gave %v, the first %v", again.GetOptimizedGear(), result.GetOptimizedGear())
	}
}

func TestOptimizeRespectsAStatCap(t *testing.T) {
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(927, 1887))
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != 927 {
		t.Errorf("wrist enchant %d, want 927 (Greater Strength)", got)
	}
}

// Greater Agility breaks the Agility ceiling on the first solve, so a second one picks Greater Strength.
func TestOptimizeReportsItsSolvePasses(t *testing.T) {
	uncapped := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(927, 1887))
	if got := optimize(t, uncapped).GetPassesDone(); got != 1 {
		t.Errorf("uncapped: %d passes, want 1", got)
	}
	capped := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(927, 1887))
	capAgilityAtTheSheet(t, capped)
	if got := optimize(t, capped).GetPassesDone(); got <= 1 {
		t.Errorf("capped: %d passes, want more than 1", got)
	}
}

func TestOptimizeLeavesASlotEmptyWhenEveryCandidateBreaksACap(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = 927
	request := enchantRequest(raid, map[proto.Stat]float64{proto.Stat_StatAgility: 2}, wrist(1887))
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != 0 {
		t.Errorf("wrist enchant %d, want none", got)
	}
}

// The baseline already exceeds the ceiling, so the ceiling only stops more Agility.
func TestOptimizeAllowsACeilingTheBaselineExceeds(t *testing.T) {
	request := enchantRequest(protectionWarriorRaid("preraid"), agilityOverStrength, wrist(927, 1887))
	capAgilityBelowTheSheet(t, request, 10)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotWrist); got != 927 {
		t.Errorf("wrist enchant %d, want 927 (Greater Strength)", got)
	}
}

// Minor Speed gives no stats, so the boots keep it when the only candidate breaks the cap.
func TestOptimizeKeepsAPermanentEffectEnchantWhenNothingIsPicked(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotFeet].Enchant = 911
	feet := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotFeet, EffectIds: []int32{1887}}
	request := enchantRequest(raid, map[proto.Stat]float64{proto.Stat_StatAgility: 2}, feet)
	capAgilityAtTheSheet(t, request)
	if got := slotEnchant(optimize(t, request), proto.ItemSlot_ItemSlotFeet); got != 911 {
		t.Errorf("feet enchant %d, want 911 (Minor Speed)", got)
	}
}

// Might of the Scourge's 1% melee and ranged crit counts as 14 crit rating: 26 AP + 14 × 4 = 82,
// against Zandalar Signet of Might's 30 AP.
func TestOptimizeValuesAPercentEnchantThroughItsRating(t *testing.T) {
	weights := map[proto.Stat]float64{proto.Stat_StatStrength: 2.2, proto.Stat_StatAttackPower: 1, proto.Stat_StatCritRating: 4}
	shoulder := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotShoulder, EffectIds: []int32{2606, 2717}}
	result := optimize(t, enchantRequest(protectionWarriorRaid("preraid"), weights, shoulder))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotShoulder); got != 2717 {
		t.Errorf("shoulder enchant %d, want 2717 (Might of the Scourge)", got)
	}
	if got := result.GetScore(); math.Abs(got-82) > 1e-9 {
		t.Errorf("score %v, want 82", got)
	}
}

// Crusader is not a candidate, so the main hand keeps it while the wrist is optimized.
func TestOptimizeDropsProcCandidates(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 1900
	crusader := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotMainHand, EffectIds: []int32{1900}}
	result := optimize(t, enchantRequest(raid, agilityOverStrength, crusader, wrist(1887)))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotMainHand); got != 1900 {
		t.Errorf("main hand enchant %d, want 1900 (Crusader)", got)
	}
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != 1887 {
		t.Errorf("wrist enchant %d, want 1887 (Greater Agility)", got)
	}
}

// A listed slot is optimized from its options even when its equipped enchant is a proc the sim
// implements.
func TestOptimizeReplacesAnImplementedProcEnchant(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Enchant = 1900
	agility := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotMainHand, EffectIds: []int32{2564}}
	result := optimize(t, enchantRequest(raid, agilityOverStrength, agility))
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotMainHand); got != 2564 {
		t.Errorf("main hand enchant %d, want 2564 (Superior Agility)", got)
	}
}

// The request's own database is merged before the candidates resolve, so an enchant only it
// carries is a candidate.
func TestOptimizeResolvesAnEnchantFromTheRequestDatabase(t *testing.T) {
	const effectID = 910201
	agility := make([]float64, len(proto.Stat_name))
	agility[proto.Stat_StatAgility] = 20
	raid := protectionWarriorRaid("preraid")
	player := raid.Parties[0].Players[0]
	player.Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = 0
	player.Database = &proto.SimDatabase{Enchants: []*proto.SimEnchant{{EffectId: effectID, Type: proto.ItemType_ItemTypeWrist, Stats: agility}}}
	if got := slotEnchant(optimize(t, enchantRequest(raid, agilityOverStrength, wrist(effectID))), proto.ItemSlot_ItemSlotWrist); got != effectID {
		t.Errorf("wrist enchant %d, want %d", got, effectID)
	}
}

// With no EP on either stat, the wrist keeps Greater Stamina rather than switching or ending empty.
func TestOptimizeKeepsTheEquippedEnchantOnATie(t *testing.T) {
	for _, options := range [][]int32{{929, 1887}, {929}} {
		raid := protectionWarriorRaid("preraid")
		raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotWrist].Enchant = 929
		result := optimize(t, enchantRequest(raid, nil, wrist(options...)))
		if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != 929 {
			t.Errorf("options %v: wrist enchant %d, want 929 (Greater Stamina)", options, got)
		}
		if got := result.GetScore(); got != 0 {
			t.Errorf("options %v: score %v, want 0", options, got)
		}
	}
}

func TestOptimizeLeavesFrozenAndUnlistedSlots(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	items := raid.Parties[0].Players[0].Equipment.Items
	items[proto.ItemSlot_ItemSlotHead].Enchant = 1506
	items[proto.ItemSlot_ItemSlotWrist].Enchant = 927
	feet := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotFeet, EffectIds: []int32{1887}}
	request := enchantRequest(raid, agilityOverStrength, wrist(1887), feet)
	request.Settings = &proto.ReforgeSettings{FreezeItemSlots: true, FrozenItemSlots: []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}}

	result := optimize(t, request)
	for _, check := range []struct {
		slot proto.ItemSlot
		want int32
	}{
		{proto.ItemSlot_ItemSlotWrist, 927},
		{proto.ItemSlot_ItemSlotHead, 1506},
		{proto.ItemSlot_ItemSlotFeet, 1887},
	} {
		if got := slotEnchant(result, check.slot); got != check.want {
			t.Errorf("%s enchant %d, want %d", check.slot, got, check.want)
		}
	}
}

func TestOptimizeSkipsAnEmptySlot(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	raid.Parties[0].Players[0].Equipment.Items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{}
	shield := &proto.BulkEnchantSlotOptions{Slot: proto.ItemSlot_ItemSlotOffHand, EffectIds: []int32{929}}
	result := optimize(t, enchantRequest(raid, agilityOverStrength, wrist(1887), shield))
	if got := result.GetOptimizedGear().Items[proto.ItemSlot_ItemSlotOffHand]; got.GetId() != 0 || got.GetEnchant() != 0 {
		t.Errorf("off hand %v, want empty", got)
	}
	if got := slotEnchant(result, proto.ItemSlot_ItemSlotWrist); got != 1887 {
		t.Errorf("wrist enchant %d, want 1887 (Greater Agility)", got)
	}
}

func TestOptimizeReturnsTheGearUnchanged(t *testing.T) {
	raid := protectionWarriorRaid("preraid")
	result := OptimizeAsync(&proto.ReforgeOptimizeRequest{Raid: raid}, simsignals.CreateSignals())
	if err := result.GetError(); err != nil {
		t.Fatalf("OptimizeAsync: %s", err.GetMessage())
	}

	if gear := result.GetOptimizedGear(); !protopkg.Equal(gear, raid.Parties[0].Players[0].Equipment) {
		t.Errorf("optimized gear %v, want the input gear", gear)
	}
	if result.GetScore() != 0 || result.GetPassesDone() != 0 {
		t.Errorf("score %v after %d passes, want 0 after 0", result.GetScore(), result.GetPassesDone())
	}
}
