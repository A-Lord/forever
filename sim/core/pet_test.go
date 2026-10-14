package core_test

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// An owner whose dynamic pet takes the better of the owner's strength and agility as attack power.
type inheritingOwner struct {
	core.Character
	pet *inheritingPet
}

type inheritingPet struct {
	core.Pet
}

func init() {
	core.RegisterAgentFactory(
		proto.Player_FeralCatDruid{},
		proto.Spec_SpecFeralCatDruid,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			owner := &inheritingOwner{Character: *character}
			owner.pet = &inheritingPet{Pet: core.NewPet(core.PetConfig{
				Name:  "Inheriting Pet",
				Owner: &owner.Character,
				StatInheritance: func(ownerStats stats.Stats) stats.Stats {
					return stats.Stats{stats.AttackPower: max(ownerStats[stats.Strength], ownerStats[stats.Agility])}
				},
				IsDynamic: true,
			})}
			owner.AddPet(owner.pet)
			return owner
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_FeralCatDruid)
		},
	)
}

func (c *inheritingOwner) GetCharacter() *core.Character       { return &c.Character }
func (c *inheritingOwner) ApplyTalents()                       {}
func (c *inheritingOwner) Reset(_ *core.Simulation)            {}
func (c *inheritingOwner) OnGCDReady(_ *core.Simulation)       {}
func (c *inheritingOwner) OnEncounterStart(_ *core.Simulation) {}
func (c *inheritingOwner) Initialize()                         {}

func (p *inheritingPet) GetPet() *core.Pet                        { return &p.Pet }
func (p *inheritingPet) Reset(_ *core.Simulation)                 {}
func (p *inheritingPet) OnEncounterStart(_ *core.Simulation)      {}
func (p *inheritingPet) ExecuteCustomRotation(_ *core.Simulation) {}

// A dynamic pet inherits from its owner's totals: strength the owner gains reaches the pet at the
// next pet heartbeat, and comes back off at the heartbeat after the owner loses it.
func TestDynamicPetInheritsFromTheOwnersTotals(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{Name: "Owner", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				Spec: &proto.Player_FeralCatDruid{}, Equipment: &proto.EquipmentSpec{}}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()
	owner := sim.Raid.Parties[0].Players[0].(*inheritingOwner)
	pet := owner.pet
	pet.Enable(sim, pet)

	inherited := func() float64 { return max(owner.GetStat(stats.Strength), owner.GetStat(stats.Agility)) }
	start := pet.GetStat(stats.AttackPower)
	if want := inherited(); start != want {
		t.Fatalf("the pet starts with %v attack power, want the owner's better of strength and agility, %v", start, want)
	}

	owner.AddStatsDynamic(sim, stats.Stats{stats.Strength: 100})
	stepTo(t, sim, sim.CurrentTime+core.PetUpdateInterval)
	if got, want := pet.GetStat(stats.AttackPower), inherited(); math.Abs(got-want) > 1e-9 {
		t.Errorf("with +100 strength on the owner the pet has %v attack power, want %v", got, want)
	}

	owner.AddStatsDynamic(sim, stats.Stats{stats.Strength: -100})
	stepTo(t, sim, sim.CurrentTime+core.PetUpdateInterval)
	if got := pet.GetStat(stats.AttackPower); math.Abs(got-start) > 1e-9 {
		t.Errorf("with the strength gone the pet has %v attack power, want the %v it started with", got, start)
	}
}
