//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/core/testagent"
	_ "github.com/wowsims/forever/sim/warrior" // registers the warrior's item effects.
)

// The stand-ins' spells, under the class flags of the spells the items name: Hamstring (family 4
// bit 1), Arcane Shot (family 9 bit 11), Multi-Shot (family 9 bit 12), and the pet's claw.
var (
	stubHamstring  = core.ActionID{SpellID: 993101}
	stubArcaneShot = core.ActionID{SpellID: 993102}
	stubMultiShot  = core.ActionID{SpellID: 993103}
	stubPetClaw    = core.ActionID{SpellID: 993104}
)

func init() {
	// The warrior stand-in: a rage bar, which the game files' tenths of rage are read against, and a
	// Hamstring costing 10 rage.
	testagent.Register(proto.Player_ProtectionWarrior{}, proto.Spec_SpecProtectionWarrior, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_ProtectionWarrior)
	}, func(agent *testagent.Agent) {
		agent.EnableRageBar(core.RageBarOptions{MaxRage: 100, BaseRageMultiplier: 1})
		hamstring := flatHit(stubHamstring, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial, 4, 1)
		hamstring.RageCost = core.RageCostOptions{Cost: 10}
		agent.RegisterSpell(hamstring)
	})

	// The hunter stand-in, on the warlock's spec because the hunter's is taken by the shot stand-in:
	// an Arcane Shot costing 100 mana, a Multi-Shot, and a pet.
	testagent.Register(proto.Player_Warlock{}, proto.Spec_SpecWarlock, func(player *proto.Player, spec any) {
		player.Spec = spec.(*proto.Player_Warlock)
	}, func(agent *testagent.Agent) {
		agent.EnableManaBar()
		arcaneShot := flatHit(stubArcaneShot, core.SpellSchoolArcane, core.ProcMaskRangedSpecial, 9, 11)
		arcaneShot.ManaCost = core.ManaCostOptions{FlatCost: 100}
		agent.RegisterSpell(arcaneShot)
		agent.RegisterSpell(flatHit(stubMultiShot, core.SpellSchoolPhysical, core.ProcMaskRangedSpecial, 9, 12))
		agent.AddPet(&clawPet{Pet: core.NewPet(core.PetConfig{
			Name: "Claw Pet", Owner: &agent.Character, EnabledOnStart: true,
			BaseStats:       stats.Stats{stats.Health: 1000},
			StatInheritance: func(stats.Stats) stats.Stats { return stats.Stats{} },
		})})
	})
}

// A pet whose claw deals 100 physical damage.
type clawPet struct {
	core.Pet
}

func (p *clawPet) GetPet() *core.Pet                        { return &p.Pet }
func (p *clawPet) ApplyTalents()                            {}
func (p *clawPet) Reset(_ *core.Simulation)                 {}
func (p *clawPet) OnEncounterStart(_ *core.Simulation)      {}
func (p *clawPet) ExecuteCustomRotation(_ *core.Simulation) {}

func (p *clawPet) Initialize() {
	p.Pet.Initialize()
	p.RegisterSpell(flatHit(stubPetClaw, core.SpellSchoolPhysical, core.ProcMaskMeleeMHSpecial, 0, 0))
}

// A spell under the family's bit that always hits for 100 no resistance or armor reduces.
func flatHit(id core.ActionID, school core.SpellSchool, mask core.ProcMask, family int32, bit uint) core.SpellConfig {
	config := core.SpellConfig{
		ActionID:         id,
		SpellSchool:      school,
		ProcMask:         mask,
		Flags:            core.SpellFlagIgnoreResists,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, 100, spell.OutcomeAlwaysHit)
		},
	}
	if family != 0 {
		config.ClassFlags = core.ClassFlags{Family: family, Mask: [4]uint32{1 << bit}}
	}
	return config
}

// The class's stand-in wearing one item, or nothing for item 0, against a level 60 target.
func modWearer(t *testing.T, class proto.Class, slot proto.ItemSlot, itemID int32) (*core.Simulation, *core.Character) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[slot].Id = itemID
	player := &proto.Player{
		Name: "Wearer", Class: class, Race: proto.Race_RaceOrc,
		Equipment:   &proto.EquipmentSpec{Items: items},
		Consumables: &proto.ConsumesSpec{},
		Buffs:       &proto.IndividualBuffs{},
		Rotation:    &proto.APLRotation{},
	}
	if class == proto.Class_ClassWarrior {
		core.WithSpec(player, &proto.Player_ProtectionWarrior{ProtectionWarrior: &proto.ProtectionWarrior{}})
	} else {
		core.WithSpec(player, &proto.Player_Warlock{Warlock: &proto.Warlock{}})
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 180},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].GetCharacter()
}

func dealtBy(sim *core.Simulation, unit *core.Unit, id core.ActionID) float64 {
	spell := unit.GetSpell(id)
	target := sim.Encounter.ActiveTargetUnits[0]
	before := spell.SpellMetrics[target.UnitIndex].TotalDamage
	spell.SkipCastAndApplyEffects(sim, target)
	return spell.SpellMetrics[target.UnitIndex].TotalDamage - before
}

// The PvP plate gloves' 22778 states -30 and Rage of Mugamba's 24428 -20 on Hamstring's cost, in
// tenths of rage: 3 and 2 off a warrior's Hamstring.
func TestPlateGlovesAndRageOfMugambaCutHamstringsRage(t *testing.T) {
	for _, c := range []struct {
		name string
		slot proto.ItemSlot
		item int32
		want float64
	}{
		{"nothing", proto.ItemSlot_ItemSlotHands, 0, 10},
		{"Knight-Lieutenant's Plate Gauntlets", proto.ItemSlot_ItemSlotHands, 16406, 7},
		{"Premier Plate Gauntlets", proto.ItemSlot_ItemSlotHands, 272717, 7},
		{"Rage of Mugamba", proto.ItemSlot_ItemSlotNeck, 19577, 8},
	} {
		_, warrior := modWearer(t, proto.Class_ClassWarrior, c.slot, c.item)
		if got := warrior.GetSpell(stubHamstring).Cost.GetCurrentCost(); got != c.want {
			t.Errorf("wearing %s, Hamstring costs %v rage, want %v", c.name, got, c.want)
		}
	}
}

// The PvP chain gloves' 23157 takes 15 mana off Arcane Shot; the Premier chain gloves' 28539 raises
// Multi-Shot's direct damage by 4% and leaves Arcane Shot alone.
func TestChainGlovesModifyTheirShot(t *testing.T) {
	_, cheap := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotHands, 16403)
	if got := cheap.GetSpell(stubArcaneShot).Cost.GetCurrentCost(); got != 85 {
		t.Errorf("wearing Knight-Lieutenant's Chain Gauntlets, Arcane Shot costs %v mana, want 85", got)
	}

	sim, hunter := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotHands, 272495)
	if got := dealtBy(sim, &hunter.Unit, stubMultiShot); !core.WithinToleranceFloat64(104, got, 1e-6) {
		t.Errorf("wearing Premier Chain Gauntlets, Multi-Shot dealt %v, want 104", got)
	}
	if got := dealtBy(sim, &hunter.Unit, stubArcaneShot); !core.WithinToleranceFloat64(100, got, 1e-6) {
		t.Errorf("wearing Premier Chain Gauntlets, Arcane Shot dealt %v, want 100", got)
	}
}

// Ferra's Collar's 22854 adds 2 to Tamed Pet Passive's damage done, which is 0 on its own: the pet
// deals 2% more, where the tooltip says 4%.
func TestFerrasCollarRaisesThePetsDamage(t *testing.T) {
	for _, c := range []struct {
		item int32
		want float64
	}{{0, 100}, {18355, 102}} {
		sim, hunter := modWearer(t, proto.Class_ClassHunter, proto.ItemSlot_ItemSlotTrinket1, c.item)
		if got := dealtBy(sim, &hunter.Pets[0].Unit, stubPetClaw); !core.WithinToleranceFloat64(c.want, got, 1e-6) {
			t.Errorf("wearing item %d, the pet's claw dealt %v, want %v", c.item, got, c.want)
		}
	}
}
