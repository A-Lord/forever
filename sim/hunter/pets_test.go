package hunter

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// The tallstrider's Dust Cloud bids in the slot the raid's Dust Cloud, Faerie Fire and Curse of
// Recklessness share: it lands beside Expose Armor and both apply, and beside the raid's own Dust
// Cloud it lands too, with the target still losing one -505.
func TestPetDustCloudSharesTheRaidsDustCloudSlot(t *testing.T) {
	for _, c := range []struct {
		name    string
		debuffs *proto.Debuffs
		other   string
		armor   float64
	}{
		{"beside Expose Armor", &proto.Debuffs{ExposeArmor: true}, "Expose Armor (External)", -505},
		{"against the raid's Dust Cloud", &proto.Debuffs{DustCloud: true}, "Dust Cloud (External)", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			player := &proto.Player{
				Name:          "Dust Cloud",
				Race:          proto.Race_RaceOrc,
				Class:         proto.Class_ClassHunter,
				Equipment:     &proto.EquipmentSpec{},
				TalentsString: DefaultBMTalents,
				Spec: &proto.Player_Hunter{Hunter: &proto.Hunter{Options: &proto.Hunter_Options{
					ClassOptions: &proto.HunterOptions{PetType: proto.HunterOptions_Tallstrider, PetUptime: 1},
				}}},
				Rotation: &proto.APLRotation{},
			}
			sim := core.NewSim(&proto.RaidSimRequest{
				Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, c.debuffs),
				Encounter:  core.MakeSingleTargetEncounter(0),
				SimOptions: &proto.SimOptions{RandomSeed: 1},
			}, simsignals.CreateSignals())
			sim.Reset()
			pet := sim.Raid.Parties[0].Players[0].(*Hunter).Pet
			target := sim.Encounter.AllTargetUnits[0]

			spell := pet.GetSpell(core.ActionID{SpellID: spellData.DustCloudTriggered.Highest().ID})
			aura := target.GetAura("Dust Cloud")
			if spell == nil || aura == nil {
				t.Fatalf("spell %v, aura %v; the tallstrider has no Dust Cloud", spell, aura)
			}
			if !spell.ExtraCastCondition(sim, target) {
				t.Error("the pet cannot cast Dust Cloud, want it castable over the other debuff")
			}

			other := target.GetAura(c.other)
			before := target.GetStat(stats.Armor)
			aura.Activate(sim)
			if !aura.IsActive() || !other.IsActive() {
				t.Errorf("the pet's Dust Cloud active %v, %s active %v; want both on the target",
					aura.IsActive(), c.other, other.IsActive())
			}
			if got := target.GetStat(stats.Armor) - before; got != c.armor {
				t.Errorf("the pet's Dust Cloud moved armor by %v, want %v", got, c.armor)
			}

			other.Deactivate(sim)
			withPet := target.GetStat(stats.Armor)
			aura.Deactivate(sim)
			if got := withPet - target.GetStat(stats.Armor); got != -505 {
				t.Errorf("with %s off, the pet's Dust Cloud is worth %v, want -505", c.other, got)
			}
		})
	}
}

// The generated family table and the options enum name the same families: a family the client
// adds needs its enum value, and an enum value the client no longer describes has to go.
func TestPetFamiliesMatchTheOptions(t *testing.T) {
	for name := range petFamilies {
		if _, ok := proto.HunterOptions_PetType_value[name]; !ok {
			t.Errorf("pet family %q has no HunterOptions.PetType value", name)
		}
	}
	for name, value := range proto.HunterOptions_PetType_value {
		if value == 0 {
			continue
		}
		if _, ok := petFamilies[name]; !ok {
			t.Errorf("HunterOptions.PetType %s names no generated pet family", name)
		}
	}
	for name, value := range proto.HunterOptions_PetAttackSpeed_value {
		if value == 0 {
			continue
		}
		if _, ok := petAttackSpeeds[proto.HunterOptions_PetAttackSpeed(value)]; !ok {
			t.Errorf("HunterOptions.PetAttackSpeed %s names no passive", name)
		}
	}
}

// The generated arrow table and the options enum name the same arrows, and every value is its
// arrow's item id, which is what the UI lists them by.
func TestArrowsMatchTheOptions(t *testing.T) {
	for name, arrow := range arrows {
		if value, ok := proto.HunterOptions_Ammo_value[name]; !ok {
			t.Errorf("arrow %q has no HunterOptions.Ammo value", name)
		} else if value != arrow.ItemID {
			t.Errorf("HunterOptions.Ammo %s is %d, not the arrow's item id %d", name, value, arrow.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_Ammo_value {
		if value == 0 {
			continue
		}
		if _, ok := arrows[name]; !ok {
			t.Errorf("HunterOptions.Ammo %s names no generated arrow", name)
		}
	}
}

// The same for the quivers.
func TestQuiversMatchTheOptions(t *testing.T) {
	for name, quiver := range quivers {
		if value, ok := proto.HunterOptions_QuiverBonus_value[name]; !ok {
			t.Errorf("quiver %q has no HunterOptions.QuiverBonus value", name)
		} else if value != quiver.ItemID {
			t.Errorf("HunterOptions.QuiverBonus %s is %d, not the quiver's item id %d", name, value, quiver.ItemID)
		}
	}
	for name, value := range proto.HunterOptions_QuiverBonus_value {
		if value == 0 {
			continue
		}
		if _, ok := quivers[name]; !ok {
			t.Errorf("HunterOptions.QuiverBonus %s names no generated quiver", name)
		}
	}
}
