//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	forever.RegisterPotionsAndConjured()
}

func warriorWithConsumes(t *testing.T, consumes *proto.ConsumesSpec, mainHand, offHand int32, mobType proto.MobType) (*core.Simulation, *core.Character) {
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
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63, MobType: mobType}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func bonusAgainstTarget(character *core.Character) stats.Stats {
	target := character.CurrentTarget
	return character.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType]
}

// Potion of Demon Slaying 9224 casts 11406: 180 attack power against Demons (creature type mask 4)
// for 300 s, on potion category 4's 120 s timer.
func TestPotionOfDemonSlayingIsAPotion(t *testing.T) {
	sim, character := warriorWithConsumes(t, &proto.ConsumesSpec{PotId: 9224}, 0, 0, proto.MobType_MobTypeDemon)

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

	want := stats.Stats{stats.AttackPower: 180}
	if got := bonusAgainstTarget(character); got != (stats.Stats{}) {
		t.Errorf("before the potion, the bonus against Demons is %v, want none", got)
	}
	aura := character.GetAuraByID(core.ActionID{SpellID: 11406})
	aura.Activate(sim)
	if aura.Duration != 300*time.Second {
		t.Errorf("the potion's aura lasts %v, want 300 s", aura.Duration)
	}
	if got := bonusAgainstTarget(character); got != want {
		t.Errorf("under the potion, the bonus against Demons is %v, want %v", got, want)
	}
	aura.Deactivate(sim)
	if got := bonusAgainstTarget(character); got != (stats.Stats{}) {
		t.Errorf("after the potion, the bonus against Demons is %v, want none", got)
	}

	undeadSim, againstUndead := warriorWithConsumes(t, &proto.ConsumesSpec{PotId: 9224}, 0, 0, proto.MobType_MobTypeUndead)
	againstUndead.GetAuraByID(core.ActionID{SpellID: 11406}).Activate(undeadSim)
	if got := bonusAgainstTarget(againstUndead); got != (stats.Stats{}) {
		t.Errorf("under the potion, the bonus against Undead is %v, want none", got)
	}
}

// Consecrated Sharpening Stone 23122 casts 28891, which puts enchant 2684 on a weapon of the
// subclasses its SpellEquippedItems mask 1097203 names; the enchant's equip spell 28893 grants 99
// melee and 99 ranged attack power against Undead (creature type mask 32), once per enchanted hand.
func TestConsecratedSharpeningStoneCountsOncePerHandItTakes(t *testing.T) {
	const sword, dagger, shield = 727, 816, 1203
	for _, tc := range []struct {
		name              string
		mainHand, offHand int32
		consumes          *proto.ConsumesSpec
		hands             float64
	}{
		{"on a sword and a dagger", sword, dagger, &proto.ConsumesSpec{MhImbueId: 23122, OhImbueId: 23122}, 2},
		{"on the sword alone", sword, dagger, &proto.ConsumesSpec{MhImbueId: 23122}, 1},
		{"on a sword and a shield it cannot take", sword, shield, &proto.ConsumesSpec{MhImbueId: 23122, OhImbueId: 23122}, 1},
		{"on nothing", sword, dagger, &proto.ConsumesSpec{}, 0},
	} {
		_, againstUndead := warriorWithConsumes(t, tc.consumes, tc.mainHand, tc.offHand, proto.MobType_MobTypeUndead)
		want := stats.Stats{stats.AttackPower: 99 * tc.hands, stats.RangedAttackPower: 99 * tc.hands}
		if got := bonusAgainstTarget(againstUndead); got != want {
			t.Errorf("%s: the bonus against Undead is %v, want %v", tc.name, got, want)
		}
		_, againstDemons := warriorWithConsumes(t, tc.consumes, tc.mainHand, tc.offHand, proto.MobType_MobTypeDemon)
		if got := bonusAgainstTarget(againstDemons); got != (stats.Stats{}) {
			t.Errorf("%s: the bonus against Demons is %v, want none", tc.name, got)
		}
	}
}

// Blessed Wizard Oil 23123 puts enchant 2685 on the weapon, whose equip spell 28890 grants 58 spell
// damage against Undead (creature type mask 32).
func TestBlessedWizardOilGrantsSpellDamageAgainstUndead(t *testing.T) {
	const sword = 727
	_, againstUndead := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: 23123}, sword, 0, proto.MobType_MobTypeUndead)
	if got, want := bonusAgainstTarget(againstUndead), (stats.Stats{stats.SpellDamage: 58}); got != want {
		t.Errorf("the bonus against Undead is %v, want %v", got, want)
	}
	_, againstDemons := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: 23123}, sword, 0, proto.MobType_MobTypeDemon)
	if got := bonusAgainstTarget(againstDemons); got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}
