package balance

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

// Self keeps the mana. Someone else still spends the cast and the cooldown; a dummy receives the mana.
func TestGiftedInnervateCastsWithoutTheMana(t *testing.T) {
	for _, c := range []struct {
		name     string
		gift     bool
		wantMana bool
	}{
		{"self", false, true},
		{"someone else", true, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			sim, druid := balanceDruid(t, c.gift, 0)
			spell := druid.GetSpell(core.ActionID{SpellID: 29166})
			if spell == nil {
				t.Fatal("Innervate is not registered")
			}

			before := druid.ManaRegenPerSecondWhileNotCasting()
			dummy := sim.Raid.GetFirstTargetDummy()
			if dummy == nil {
				t.Fatal("Balance has no dummy")
			}

			target := sim.Encounter.ActiveTargetUnits[0]
			if !spell.Cast(sim, target) {
				t.Fatal("Innervate did not cast")
			}
			if spell.IsReady(sim) {
				t.Error("Innervate is still ready")
			}
			if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
				t.Errorf("Innervate casts = %d, want 1", got)
			}

			active := druid.HasActiveAuraWithTag(buffs.InnervatesCategory)
			gotMana := druid.ManaRegenPerSecondWhileNotCasting() != before
			if active != c.wantMana || gotMana != c.wantMana {
				t.Errorf("mana buff active = %v, regen changed = %v, want %v", active, gotMana, c.wantMana)
			}
			if gotDummy := dummy.HasActiveAuraWithTag(buffs.InnervatesCategory); gotDummy == c.wantMana {
				t.Errorf("dummy Innervate = %v, want %v", gotDummy, !c.wantMana)
			}
		})
	}

	t.Run("someone else with an incoming innervate", func(t *testing.T) {
		sim, druid := balanceDruid(t, true, 1)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		incoming := druid.GetAura("Innervates (External)")
		if incoming == nil {
			t.Fatal("the incoming Innervate is not registered")
		}
		incoming.Activate(sim)
		if !druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Fatal("the incoming Innervate is not up")
		}
		target := sim.Encounter.ActiveTargetUnits[0]
		if !spell.Cast(sim, target) {
			t.Fatal("Innervate did not cast")
		}
		if spell.IsReady(sim) {
			t.Error("Innervate is still ready")
		}
		if got := spell.SpellMetrics[target.UnitIndex].Casts; got != 1 {
			t.Errorf("Innervate casts = %d, want 1", got)
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})

	t.Run("cast at the dummy", func(t *testing.T) {
		sim, druid := balanceDruid(t, false, 0)
		spell := druid.GetSpell(core.ActionID{SpellID: 29166})
		if spell == nil {
			t.Fatal("Innervate is not registered")
		}
		dummy := sim.Raid.GetFirstTargetDummy()
		if dummy == nil {
			t.Fatal("Balance has no dummy")
		}
		before := druid.ManaRegenPerSecondWhileNotCasting()
		if !spell.Cast(sim, &dummy.Unit) {
			t.Fatal("Innervate did not cast")
		}
		if druid.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the caster has Innervate")
		}
		if druid.ManaRegenPerSecondWhileNotCasting() != before {
			t.Error("the caster's regen changed")
		}
		if !dummy.HasActiveAuraWithTag(buffs.InnervatesCategory) {
			t.Error("the dummy has no Innervate")
		}
	})
}

func balanceDruid(t *testing.T, gift bool, incoming int32) (*core.Simulation, *core.Character) {
	t.Helper()
	target := &proto.UnitReference{}
	if gift {
		target = &proto.UnitReference{Type: proto.UnitReference_Player, Index: 1}
	}
	player := core.WithSpec(&proto.Player{
		Class:         proto.Class_ClassDruid,
		Race:          proto.Race_RaceNightElf,
		Equipment:     &proto.EquipmentSpec{},
		TalentsString: DefaultTalents,
		Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
		Buffs:         &proto.IndividualBuffs{Innervates: incoming},
	}, &proto.Player_BalanceDruid{
		BalanceDruid: &proto.BalanceDruid{
			Options: &proto.BalanceDruid_Options{
				ClassOptions: &proto.DruidOptions{InnervateTarget: target},
			},
		},
	})
	raid := core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{})
	raid.TargetDummies = 1
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid:       raid,
		Encounter:  core.MakeSingleTargetEncounter(0),
		SimOptions: &proto.SimOptions{RandomSeed: 1},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}
