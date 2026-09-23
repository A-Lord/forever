//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func warriorWithConsumes(t *testing.T, consumes *proto.ConsumesSpec, mainHand, offHand int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotOffHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = mainHand
	items[proto.ItemSlot_ItemSlotOffHand].Id = offHand
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Consumer", Class: proto.Class_ClassWarrior, Race: proto.Race_RaceOrc,
				Spec: &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{Options: &proto.DpsWarrior_Options{
					ClassOptions: &proto.WarriorOptions{DefaultStance: proto.WarriorStance_WarriorStanceBerserker}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func bonusAgainst(character *core.Character, mobType proto.MobType) stats.Stats {
	return character.AttackTables[character.CurrentTarget.UnitIndex].MobTypeBonusStats[mobType]
}

// Potion of Demonslaying 9224 casts 11406: 108 spell damage and 185 attack power against Demons
// (creature type mask 4) for 120 s, on potion category 4's 120 s timer.
func TestPotionOfDemonslayingIsAPotion(t *testing.T) {
	sim, character := warriorWithConsumes(t, &proto.ConsumesSpec{PotId: 9224}, 0, 0)

	spell := character.GetSpell(core.ActionID{ItemID: 9224})
	if spell == nil {
		t.Fatal("9224 registers no spell")
	}
	if !spell.Flags.Matches(core.SpellFlagPotion | core.SpellFlagCombatPotion) {
		t.Error("9224 is not the combat potion")
	}
	if spell.SharedCD.Timer != character.GetOrInitSpellCategoryTimer(4) || spell.SharedCD.Duration != 120*time.Second {
		t.Errorf("9224's shared cooldown is %v on another timer than category 4's, want 120 s on it", spell.SharedCD.Duration)
	}

	want := stats.Stats{stats.AttackPower: 185, stats.SpellDamage: 108}
	if got := bonusAgainst(character, proto.MobType_MobTypeDemon); got != (stats.Stats{}) {
		t.Errorf("before the potion, the bonus against Demons is %v, want none", got)
	}
	aura := character.GetAuraByID(core.ActionID{SpellID: 11406})
	aura.Activate(sim)
	if aura.Duration != 120*time.Second {
		t.Errorf("the potion's aura lasts %v, want 120 s", aura.Duration)
	}
	if got := bonusAgainst(character, proto.MobType_MobTypeDemon); got != want {
		t.Errorf("under the potion, the bonus against Demons is %v, want %v", got, want)
	}
	if got := bonusAgainst(character, proto.MobType_MobTypeUndead); got != (stats.Stats{}) {
		t.Errorf("under the potion, the bonus against Undead is %v, want none", got)
	}
	aura.Deactivate(sim)
	if got := bonusAgainst(character, proto.MobType_MobTypeDemon); got != (stats.Stats{}) {
		t.Errorf("after the potion, the bonus against Demons is %v, want none", got)
	}
}
