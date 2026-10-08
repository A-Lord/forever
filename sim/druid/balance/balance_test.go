package balance

import (
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get caster sets included. (we use spellfire here)
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
)

func TestMoonkinThickHideTracksDefenseAndForm(t *testing.T) {
	player := &proto.Player{
		Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		Spec: &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{
			Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}},
		}},
		Equipment: &proto.EquipmentSpec{}, Consumables: &proto.ConsumesSpec{},
		Rotation: &proto.APLRotation{},
	}
	character := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, player)
	character.Env = &core.Environment{MeasuringStats: true}
	moonkin := NewBalanceDruid(&character, player)
	moonkin.Talents.MoonkinForm = true
	moonkin.Talents.ThickHide = 3
	moonkin.RegisterMoonkinFormAura()
	moonkin.ApplyTalents()
	moonkin.FinalizeStatDeps()
	sim := &core.Simulation{Environment: character.Env}
	armor := func() float64 { return moonkin.ApplyStatDependencies(moonkin.GetStats())[stats.Armor] }
	before := armor()
	moonkin.MoonkinFormAura.Activate(sim)
	if got, want := armor()-before, 60*3*4.6; math.Abs(got-want) > 1e-9 {
		t.Fatalf("Moonkin Thick Hide armor = %g, want %g", got, want)
	}
	beforeDefense := armor()
	moonkin.AddStatDynamic(sim, stats.DefenseRating, 4)
	if got, want := armor()-beforeDefense, 4*2*4.6; math.Abs(got-want) > 1e-9 {
		t.Fatalf("armor from four Defense = %g, want %g", got, want)
	}
	moonkin.ClearForm(sim)
	if got := armor(); math.Abs(got-before) > 1e-9 {
		t.Fatalf("humanoid armor = %g, want %g", got, before)
	}
	moonkin.AddStatDynamic(sim, stats.DefenseRating, -4)
	moonkin.MoonkinFormAura.Activate(sim)
	if got, want := armor()-before, 60*3*4.6; math.Abs(got-want) > 1e-9 {
		t.Fatalf("Moonkin armor on re-entry = %g, want %g", got, want)
	}
	t.Log("Moonkin Thick Hide adds 828 armor; four Defense adds 36.8; leaving the form removes both")
}

func init() {
	RegisterBalanceDruid()
	common.RegisterAllEffects()
}

func TestBalance(t *testing.T) {
	core.RunTestSuite(t, t.Name(), core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{
		{
			Class:      proto.Class_ClassDruid,
			Race:       proto.Race_RaceNightElf,
			OtherRaces: []proto.Race{proto.Race_RaceTauren},
			SpecOptions: core.SpecOptionsCombo{Label: "Standard", SpecOptions: &proto.Player_BalanceDruid{
				BalanceDruid: &proto.BalanceDruid{
					Options: &proto.BalanceDruid_Options{
						ClassOptions: &proto.DruidOptions{},
					},
				},
			}},
			GearSet: core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1"),
			OtherGearSets: []core.GearSetCombo{
				core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "prebis"),
			},
			Talents:  DefaultTalents,
			Rotation: core.GetAplRotation("../../../ui/specs/druid/balance/apls", "default"),
			ItemFilter: core.ItemFilter{
				WeaponTypes:       DefaultWeaponTypes,
				ArmorType:         DefaultArmorType,
				RangedWeaponTypes: DefaultRangedWeaponTypes,
			},
		},
	}))
}

const DefaultTalents = "5532220115501351--505"

const DefaultArmorType = proto.ArmorType_ArmorTypeLeather

var DefaultWeaponTypes = []proto.WeaponType{
	proto.WeaponType_WeaponTypeDagger,
	proto.WeaponType_WeaponTypeMace,
	proto.WeaponType_WeaponTypeStaff,
	proto.WeaponType_WeaponTypeOffHand,
}

var DefaultRangedWeaponTypes = []proto.RangedWeaponType{
	proto.RangedWeaponType_RangedWeaponTypeIdol,
}

func TestClearcastingProcDuringPaidHurricaneIsRetained(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name: "Channel case", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
			TalentsString: DefaultTalents,
			Spec:          &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
			Equipment:     core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1").GearSet,
			Consumables:   &proto.ConsumesSpec{}, Rotation: &proto.APLRotation{},
		}}}}},
		Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	moonkin := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	moonkin.OmenOfClarityAura.Deactivate(sim)
	if moonkin.ClearcastingAura.IsActive() {
		t.Fatal("Clearcasting is active before the paid channel")
	}
	before := moonkin.CurrentMana()
	if !moonkin.Hurricane.Cast(sim, moonkin.CurrentTarget) {
		t.Fatal("Hurricane did not cast")
	}
	t.Logf("channel start: mana spent=%g, charge=%v", before-moonkin.CurrentMana(), moonkin.ClearcastingAura.IsActive())
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: time.Second, OnAction: func(sim *core.Simulation) {
		moonkin.ClearcastingAura.Activate(sim)
		moonkin.ClearcastingAura.SetStacks(sim, moonkin.ClearcastingAura.MaxStacks)
		t.Logf("mid-channel: channel=%v, charge=%v", moonkin.IsChanneling(), moonkin.ClearcastingAura.IsActive())
	}}))
	reached := false
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: 11 * time.Second, OnAction: func(*core.Simulation) { reached = true }}))
	for steps := 0; !reached && steps < 10000; steps++ {
		sim.Step()
	}
	if !reached {
		t.Fatal("the sim did not reach the channel-end observation")
	}
	t.Logf("after channel: channel=%v, charge=%v, next Starfire cost=%g", moonkin.IsChanneling(), moonkin.ClearcastingAura.IsActive(), moonkin.Starfire[len(moonkin.Starfire)-1].Cost.GetCurrentCost())
	if !moonkin.ClearcastingAura.IsActive() {
		t.Fatal("a charge gained during the paid channel was consumed before the next paid spell")
	}
}

func TestClearcastingProcDuringFreeHurricaneIsRetained(t *testing.T) {
	sim := core.NewSim(&proto.RaidSimRequest{
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name: "Channel case", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
			TalentsString: DefaultTalents,
			Spec:          &proto.Player_BalanceDruid{BalanceDruid: &proto.BalanceDruid{Options: &proto.BalanceDruid_Options{ClassOptions: &proto.DruidOptions{}}}},
			Equipment:     core.GetGearSet("../../../ui/specs/druid/balance/gear_sets", "p1").GearSet,
			Consumables:   &proto.ConsumesSpec{}, Rotation: &proto.APLRotation{},
		}}}}},
		Encounter:  &proto.Encounter{Duration: 30, Targets: []*proto.Target{core.NewDefaultTarget()}},
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
	}, simsignals.CreateSignals())
	sim.Reset()
	moonkin := sim.Raid.Parties[0].Players[0].(druid.DruidAgent).GetDruid()
	moonkin.OmenOfClarityAura.Deactivate(sim)
	if moonkin.ClearcastingAura.IsActive() {
		t.Fatal("Clearcasting is active before the paid channel")
	}
	moonkin.ClearcastingAura.Activate(sim)
	moonkin.ClearcastingAura.SetStacks(sim, moonkin.ClearcastingAura.MaxStacks)
	before := moonkin.CurrentMana()
	if !moonkin.Hurricane.Cast(sim, moonkin.CurrentTarget) {
		t.Fatal("Hurricane did not cast")
	}
	if spent := before - moonkin.CurrentMana(); spent != 0 {
		t.Fatalf("the initial charge did not cover the channel cost: spent %g mana", spent)
	}
	if moonkin.ClearcastingAura.IsActive() {
		t.Fatal("the charge paying for Hurricane was not spent at cast completion")
	}
	t.Logf("channel start: mana spent=%g, charge=%v", before-moonkin.CurrentMana(), moonkin.ClearcastingAura.IsActive())
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: time.Second, OnAction: func(sim *core.Simulation) {
		moonkin.ClearcastingAura.Activate(sim)
		moonkin.ClearcastingAura.SetStacks(sim, moonkin.ClearcastingAura.MaxStacks)
		t.Logf("mid-channel: channel=%v, charge=%v", moonkin.IsChanneling(), moonkin.ClearcastingAura.IsActive())
	}}))
	reached := false
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: 11 * time.Second, OnAction: func(*core.Simulation) { reached = true }}))
	for steps := 0; !reached && steps < 10000; steps++ {
		sim.Step()
	}
	if !reached {
		t.Fatal("the sim did not reach the channel-end observation")
	}
	t.Logf("after channel: channel=%v, charge=%v, next Starfire cost=%g", moonkin.IsChanneling(), moonkin.ClearcastingAura.IsActive(), moonkin.Starfire[len(moonkin.Starfire)-1].Cost.GetCurrentCost())
	if !moonkin.ClearcastingAura.IsActive() {
		t.Fatal("a charge gained during the free channel was consumed before the next paid spell")
	}
}
