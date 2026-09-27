package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Weaponmaster's sword branch is up only while a sword is equipped: with an axe the aura stays down,
// so its uptime and an APL aura check read the weapon rather than the talent.
func TestWeaponmasterSwordAuraFollowsTheWeapon(t *testing.T) {
	cases := []struct {
		name   string
		weapon int32
		want   bool
	}{
		{"Shadowhide Two-handed Sword", 1460, true},
		{"Lupine Axe", 1220, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			items := make([]*proto.ItemSpec, 17)
			for i := range items {
				items[i] = &proto.ItemSpec{}
			}
			items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: c.weapon}

			sim := core.NewSim(&proto.RaidSimRequest{
				Raid: core.SinglePlayerRaidProto(&proto.Player{
					Name:          "Weaponmaster",
					Race:          proto.Race_RaceHuman,
					Class:         proto.Class_ClassWarrior,
					Equipment:     &proto.EquipmentSpec{Items: items},
					TalentsString: DefaultArmsTalents,
					Spec:          DefaultOptions,
				}, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
				Encounter:  core.MakeSingleTargetEncounter(0),
				SimOptions: &proto.SimOptions{RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()

			war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
			aura := war.GetAura("Weaponmaster (Sword)")
			if aura == nil {
				t.Fatal("Weaponmaster (Sword) is not registered; the talent string no longer takes Weaponmaster")
			}
			if aura.IsActive() != c.want {
				t.Errorf("Weaponmaster (Sword) active = %v, want %v", aura.IsActive(), c.want)
			}
		})
	}
}
