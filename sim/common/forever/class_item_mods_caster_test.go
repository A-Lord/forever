//go:build with_db

package forever_test

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/shaman"
	"github.com/wowsims/forever/sim/shaman/enhancement"
)

func init() {
	enhancement.RegisterEnhancementShaman()
}

// An Enhancement shaman with Maelstrom Weapon, no weapon and one totem, against a level 63 target.
func enhancementShamanWithTotem(t *testing.T, totemID int32) (*core.Simulation, *shaman.Shaman) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotRanged].Id = totemID
	player := &proto.Player{
		Name: "Shaman", Class: proto.Class_ClassShaman, Race: proto.Race_RaceOrc,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Consumables:   &proto.ConsumesSpec{},
		Buffs:         &proto.IndividualBuffs{},
		Rotation:      &proto.APLRotation{},
		TalentsString: "050533102-053030031005112251",
		Spec: &proto.Player_EnhancementShaman{EnhancementShaman: &proto.EnhancementShaman{
			Options: &proto.EnhancementShaman_Options{ClassOptions: &proto.ShamanOptions{}}}},
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Duration: 25000, Targets: []*proto.Target{{Name: "target", Level: 63}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(shaman.ShamanAgent).GetShaman()
}

// Burning Totem 272433's equip spell 1291077 states +3 s on Flame Shock, family 11 mask
// 0x10000000: every rank's dot lasts 3 s longer.
func TestBurningTotemExtendsFlameShock(t *testing.T) {
	_, without := enhancementShamanWithTotem(t, 0)
	_, with := enhancementShamanWithTotem(t, 272433)
	if len(without.FlameShock) == 0 {
		t.Fatal("the shaman registers no Flame Shock")
	}

	for i, spell := range with.FlameShock {
		base := without.FlameShock[i].Dot(without.CurrentTarget).BaseDuration()
		if got, want := spell.Dot(with.CurrentTarget).BaseDuration(), base+3*time.Second; got != want {
			t.Errorf("Flame Shock %v lasts %v with the totem, want %v", spell.ActionID, got, want)
		}
	}
}

// Totem of the Storm 272432 lets a landed Lightning Bolt grant a Maelstrom Weapon stack at half the
// talent's 10 per minute, priced on the bolt's cast time. Chain Lightning and the copy Lightning
// Overload casts never do.
func TestTotemOfTheStormLetsLightningBoltGrantMaelstromWeapon(t *testing.T) {
	sim, wearer := enhancementShamanWithTotem(t, 272432)
	if wearer.Talents.MaelstromWeapon == 0 {
		t.Fatal("the shaman has no Maelstrom Weapon")
	}
	wearer.AutoAttacks.CancelAutoSwing(sim)
	target := wearer.CurrentTarget

	maelstrom := wearer.GetAuraByID(core.ActionID{SpellID: 408505})
	if maelstrom == nil {
		t.Fatal("no Maelstrom Weapon buff 408505 on the shaman")
	}
	var bolt, chain, overload *core.Spell
	for _, spell := range wearer.Spellbook {
		switch {
		case spell.Matches(shaman.SpellMaskLightningBolt):
			bolt = spell
		case spell.Matches(shaman.SpellMaskChainLightning):
			chain = spell
		case spell.Matches(shaman.SpellMaskLightningBoltOverload):
			overload = spell
		}
	}

	// Casts the spell n times and counts the times it left the buff up.
	gains := func(spell *core.Spell, n int) int {
		count := 0
		for range n {
			maelstrom.Deactivate(sim)
			spell.SkipCastAndApplyEffects(sim, target)
			stepTo(t, sim, sim.CurrentTime+time.Second)
			if maelstrom.IsActive() {
				count++
			}
		}
		return count
	}

	if got := gains(chain, 200) + gains(overload, 200); got != 0 {
		t.Errorf("Chain Lightning and the overload granted Maelstrom Weapon %d times, want never", got)
	}

	// Against a level 63 target about one bolt in six misses, so a miss that rolled would show.
	granted := gains(bolt, 20000)
	m := bolt.SpellMetrics[target.UnitIndex]
	landed := m.Hits + m.Crits
	want := 10 * 0.5 * bolt.DefaultCast.CastTime.Minutes()
	if got := float64(granted) / float64(landed); math.Abs(got-want) > 0.01 {
		t.Errorf("%d of %d landed Lightning Bolts (%v cast) granted a stack, %.3f a bolt, want %.3f",
			granted, landed, bolt.DefaultCast.CastTime, got, want)
	}
}
