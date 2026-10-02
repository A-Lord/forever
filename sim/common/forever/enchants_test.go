//go:build with_db

package forever_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
)

const revelationEnchant = 8217

func init() {
	core.RegisterAgentFactory(
		proto.Player_BalanceDruid{},
		proto.Spec_SpecBalanceDruid,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &testDruid{Character: *character}
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_BalanceDruid)
		},
	)
}

// A caster with two druid spells: Wrath, which Revelation's druid buff 1323400 names (family 7 bit 0),
// and Moonfire, which it does not (bit 1).
type testDruid struct {
	core.Character
	wrath, moonfire *core.Spell
	last            *core.SpellResult
}

func (d *testDruid) GetCharacter() *core.Character       { return &d.Character }
func (d *testDruid) ApplyTalents()                       {}
func (d *testDruid) Reset(_ *core.Simulation)            {}
func (d *testDruid) OnGCDReady(_ *core.Simulation)       {}
func (d *testDruid) OnEncounterStart(_ *core.Simulation) {}

func (d *testDruid) Initialize() {
	nuke := func(id int32, bit uint) *core.Spell {
		return d.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      core.SpellSchoolNature,
			DefenseType:      core.DefenseTypeMagic,
			ProcMask:         core.ProcMaskSpellDamage,
			ClassFlags:       core.ClassFlags{Family: 7, Mask: [4]uint32{1 << bit}},
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				d.last = spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeMagicHitAndCrit)
			},
		})
	}
	d.wrath = nuke(5176, 0)
	d.moonfire = nuke(8921, 1)
}

func revelationWearer(t *testing.T, class proto.Class, spec any) (*core.Simulation, *core.Character) {
	t.Helper()
	items := weaponSlots(873, 0)
	items[proto.ItemSlot_ItemSlotMainHand].Enchant = revelationEnchant
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceTauren,
		Equipment: &proto.EquipmentSpec{Items: items},
		Buffs:     &proto.IndividualBuffs{},
		Rotation:  &proto.APLRotation{},
	}
	switch s := spec.(type) {
	case *proto.Player_BalanceDruid:
		player.Spec = s
	case *proto.Player_DpsWarrior:
		player.Spec = s
	}

	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{player},
			Buffs:   &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

// Enchant Weapon - Revelation 8217 gives a druid its class's buff 1323400 on a direct spell that lands
// without a crit: one charge of +100% crit for 15 s on the spells its mask names, spent by the next of
// them and not by a spell outside the mask. A warrior has no Revelation buff.
func TestRevelationGrantsTheWearersClassBuff(t *testing.T) {
	sim, character := revelationWearer(t, proto.Class_ClassDruid, &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{}})
	druid := sim.Raid.Parties[0].Players[0].(*testDruid)
	target := character.CurrentTarget

	buff := character.GetAuraByID(core.ActionID{SpellID: 1323400})
	if buff == nil {
		t.Fatal("no Revelation buff 1323400 on the druid")
	}
	if buff.Duration != 15*time.Second || buff.MaxStacks != 1 {
		t.Errorf("buff lasts %v with %d charges, want 15s and 1", buff.Duration, buff.MaxStacks)
	}

	// A proc's handler runs one spell batch window after the hit, so each cast is followed by a step.
	castAndStep := func(spell *core.Spell) *core.SpellResult {
		spell.SkipCastAndApplyEffects(sim, target)
		sim.Step()
		return druid.last
	}

	for casts := 0; !buff.IsActive(); casts++ {
		if casts > 300 {
			t.Fatal("300 Moonfires and Revelation never procced")
		}
		castAndStep(druid.moonfire)
	}

	castAndStep(druid.moonfire)
	if !buff.IsActive() {
		t.Fatal("Moonfire, which the buff's mask does not name, spent its charge")
	}
	if result := castAndStep(druid.wrath); !result.DidCrit() {
		t.Error("Wrath with Revelation up did not crit")
	}
	if buff.IsActive() {
		t.Error("Wrath did not spend Revelation's charge")
	}

	_, warrior := revelationWearer(t, proto.Class_ClassWarrior, &proto.Player_DpsWarrior{DpsWarrior: &proto.DpsWarrior{
		Options: &proto.DpsWarrior_Options{ClassOptions: &proto.WarriorOptions{}}}})
	if warrior.GetAura("Revelation") != nil {
		t.Error("a warrior has a Revelation buff")
	}
}
