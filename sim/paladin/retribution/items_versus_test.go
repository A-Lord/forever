//go:build with_db

package retribution

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

func paladinAgainstUndead(t *testing.T, equipped map[proto.ItemSlot]int32) *core.Character {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	for slot, id := range equipped {
		items[slot].Id = id
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Paladin", Class: proto.Class_ClassPaladin, Race: proto.Race_RaceHuman,
				Spec: &proto.Player_RetributionPaladin{RetributionPaladin: &proto.RetributionPaladin{
					Options: &proto.RetributionPaladin_Options{ClassOptions: &proto.PaladinOptions{}}}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Duration: 180, Targets: []*proto.Target{
			{Name: "target", Level: 63, MobType: proto.MobType_MobTypeUndead}}},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Justice Battlegear's 4-piece bonus 1301084 states A_MOD_MELEE_ATTACK_POWER_VERSUS 36 against mask
// 32, Undead, and no ranged attack power.
func TestJusticeBattlegearFourPieceGrantsAttackPowerAgainstUndead(t *testing.T) {
	character := paladinAgainstUndead(t, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotLegs:     273908,
		proto.ItemSlot_ItemSlotHands:    273909,
		proto.ItemSlot_ItemSlotShoulder: 273910,
		proto.ItemSlot_ItemSlotFeet:     273911,
	})
	table := character.AttackTables[character.CurrentTarget.UnitIndex]

	if got, want := table.MobTypeBonusStats[proto.MobType_MobTypeUndead], (stats.Stats{stats.AttackPower: 36}); got != want {
		t.Errorf("the bonus against Undead is %v, want %v", got, want)
	}
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon]; got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}

// Tenets of the Silver Hand 249397's equip spell 1302545 states A_MOD_DAMAGE_DONE_VERSUS 1 against
// mask 32: 1% more damage against Undead, with crit damage following the damage.
func TestTenetsOfTheSilverHandMultipliesDamageAgainstUndead(t *testing.T) {
	without := paladinAgainstUndead(t, nil)
	with := paladinAgainstUndead(t, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 249397})
	base := without.AttackTables[without.CurrentTarget.UnitIndex]
	table := with.AttackTables[with.CurrentTarget.UnitIndex]

	if got, want := table.DamageDealtMultiplier, base.DamageDealtMultiplier*1.01; got != want {
		t.Errorf("the damage multiplier against Undead is %v, want %v", got, want)
	}
	if got, want := table.CritMultiplier, base.CritMultiplier; got != want {
		t.Errorf("the crit multiplier against Undead is %v, want %v", got, want)
	}
}
