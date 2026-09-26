package dps

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

func TestTouchOfTheGraveProcsOnlyFromDamagingAbilities(t *testing.T) {
	setup := func(t *testing.T, spellID int32) (*core.Simulation, *core.Unit, *core.Spell, *core.Aura) {
		t.Helper()
		player := &proto.Player{
			Name:          "Undead",
			Race:          proto.Race_RaceUndead,
			Class:         proto.Class_ClassWarrior,
			Equipment:     &proto.EquipmentSpec{},
			TalentsString: DefaultArmsTalents,
			Spec:          DefaultOptions,
		}
		sim := core.NewSim(&proto.RaidSimRequest{
			Raid:       core.SinglePlayerRaidProto(player, &proto.PartyBuffs{}, &proto.RaidBuffs{}, &proto.Debuffs{}),
			Encounter:  core.MakeSingleTargetEncounter(0),
			SimOptions: &proto.SimOptions{RandomSeed: 1},
		}, simsignals.CreateSignals())
		sim.Reset()

		war := sim.Raid.Parties[0].Players[0].(*DpsWarrior)
		spell := war.GetSpell(core.ActionID{SpellID: spellID})
		aura := war.GetAura("Touch of the Grave")
		if spell == nil || aura == nil || !aura.IsActive() || aura.Icd == nil {
			t.Fatalf("spell %d %v, Touch of the Grave %v; want both registered and the aura active", spellID, spell, aura)
		}
		return sim, sim.Encounter.AllTargetUnits[0], spell, aura
	}

	// Time never advances, so the first proc leaves the cooldown running for good.
	procsWithin := func(sim *core.Simulation, target *core.Unit, spell *core.Spell, aura *core.Aura, tries int) bool {
		for range tries {
			spell.SkipCastAndApplyEffects(sim, target)
			if !aura.Icd.IsReady(sim) {
				return true
			}
		}
		return false
	}

	t.Run("Heroic Strike triggers it", func(t *testing.T) {
		sim, target, heroicStrike, aura := setup(t, 25286)
		if !procsWithin(sim, target, heroicStrike, aura, 400) {
			t.Error("no Touch of the Grave in 400 Heroic Strikes")
		}
	})

	t.Run("Sunder Armor does not", func(t *testing.T) {
		sim, target, sunderArmor, aura := setup(t, 11597)
		if procsWithin(sim, target, sunderArmor, aura, 400) {
			t.Error("Sunder Armor deals no damage but triggered Touch of the Grave")
		}
	})
}
