package feralcat

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func TestManualCrowdPummelerHasteDropsWhenSwappedOff(t *testing.T) {
	if !core.WITH_DB {
		t.Skip("requires the committed game database")
	}
	const mcpID, staffID, hasteSpellID int32 = 9449, 866, 13494
	percent := spelldata.MustFind(hasteSpellID).SpeedPseudoStats()[proto.PseudoStat_PseudoStatMeleeHastePercent]
	if percent <= 0 {
		t.Fatal("the haste spell states no melee haste")
	}
	factor := 1 + percent/100
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	swap := make([]*proto.ItemSpec, len(items))
	for i := range items {
		items[i] = &proto.ItemSpec{}
		swap[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = mcpID
	swap[proto.ItemSlot_ItemSlotMainHand].Id = staffID

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: core.SinglePlayerRaidProto(&proto.Player{
			Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
			Spec: DefaultSpecOptions, Equipment: &proto.EquipmentSpec{Items: items},
			EnableItemSwap: true, ItemSwap: &proto.ItemSwap{Items: swap},
			Consumables: &proto.ConsumesSpec{}, Buffs: &proto.IndividualBuffs{}, Rotation: &proto.APLRotation{},
		}, &proto.PartyBuffs{}, nil, nil),
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)

	use := cat.GetSpell(core.ActionID{ItemID: mcpID})
	if use == nil {
		t.Fatal("Manual Crowd Pummeler has no use")
	}
	before := cat.TotalMeleeHasteMultiplier()
	if !use.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Manual Crowd Pummeler did not cast")
	}
	haste := cat.GetAuraByID(core.ActionID{SpellID: hasteSpellID})
	if !haste.IsActive() {
		t.Fatal("haste aura is not up after the use")
	}
	if got := cat.TotalMeleeHasteMultiplier(); math.Abs(got-before*factor) > 1e-9 {
		t.Fatalf("melee haste after the use = %v, want %v", got, before*factor)
	}

	cat.ItemSwap.SwapItems(sim, proto.APLActionItemSwap_Swap1, false)
	if haste.IsActive() {
		t.Fatal("haste aura stayed up after Manual Crowd Pummeler was swapped off")
	}
	if got := cat.TotalMeleeHasteMultiplier(); math.Abs(got-before) > 1e-9 {
		t.Fatalf("melee haste after the swap = %v, want %v", got, before)
	}
}
