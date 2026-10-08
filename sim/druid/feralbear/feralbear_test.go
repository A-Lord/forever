//go:build with_db

package feralbear

import (
	"os"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/druid"
	"google.golang.org/protobuf/encoding/protojson"
)

func init() {
	RegisterFeralBearDruid()
	common.RegisterAllEffects()
}

func TestFeralBear(t *testing.T) {
	generators := core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{starterConfig()})
	generators = append(generators, &core.SingleDpsTestGenerator{Name: "StarterCasts", Request: starterRequest(t)})
	core.RunTestSuite(t, t.Name(), generators)
}

func starterConfig() core.CharacterSuiteConfig {
	reference := proto.Stat_StatAttackPower
	weigh := []proto.Stat{proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatStamina}
	config := core.GetTestBuildFromJSON(
		proto.Class_ClassDruid, "../../../ui/specs/druid/feralbear/builds", "starter",
		core.ItemFilter{
			ArmorType: proto.ArmorType_ArmorTypeLeather,
			WeaponTypes: []proto.WeaponType{
				proto.WeaponType_WeaponTypeDagger, proto.WeaponType_WeaponTypeFist,
				proto.WeaponType_WeaponTypeMace, proto.WeaponType_WeaponTypeStaff,
			},
			RangedWeaponTypes: []proto.RangedWeaponType{proto.RangedWeaponType_RangedWeaponTypeIdol},
		}, &reference, &weigh,
	)
	config.OtherRaces = []proto.Race{proto.Race_RaceTauren}
	return config
}

func BenchmarkSimulate(b *testing.B) {
	core.CharacterBenchmark(b, starterConfig())
}

// The same talent string the cat coverage uses. The opener leaves dire bear and comes back so both
// form spells are casts, then the bear abilities run.
const coverageTalents = "0000000000000000-00000000100100000001-0500300000000000"

func TestFeralBearCoverage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), []core.TestGenerator{castsGen{bearCoverageRequest()}})
}

type castsGen struct {
	rsr *proto.RaidSimRequest
}

func (g castsGen) NumTests() int { return 1 }

func (g castsGen) GetTest(int) (string, *proto.ComputeStatsRequest, *proto.StatWeightsRequest, *proto.RaidSimRequest) {
	return "Casts", nil, nil, g.rsr
}

func bearCoverageRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: coverageTalents,
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 100},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation: &proto.APLRotation{
					Type: proto.APLRotation_TypeAPL,
					PriorityList: []*proto.APLListItem{
						// One of each before Berserk removes Primal Bite's cooldown and it spends the rage.
						sequence(9908, 1235827, 8983, 5209, 9898, 1238073),
						cast(9881),
						cast(5229),
						cast(22842),
						cast(6795),
						cast(1238122),
						cast(22812),
						castUnlessOnTarget(9907),
						cast(417141),
						sequence(768, 9634),
					},
				},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}
}

func cast(id int32) *proto.APLListItem {
	return &proto.APLListItem{Action: spellAction(id)}
}

// A debuff the druid can refresh at any time, cast only while the target does not have it.
func castUnlessOnTarget(id int32) *proto.APLListItem {
	action := spellAction(id)
	action.Condition = &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: &proto.APLValue{
		Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
			SourceUnit: &proto.UnitReference{Type: proto.UnitReference_CurrentTarget},
			AuraId:     &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
		}},
	}}}}
	return &proto.APLListItem{Action: action}
}

func sequence(ids ...int32) *proto.APLListItem {
	actions := make([]*proto.APLAction, len(ids))
	for i, id := range ids {
		actions[i] = spellAction(id)
	}
	return &proto.APLListItem{Action: &proto.APLAction{Action: &proto.APLAction_StrictSequence{
		StrictSequence: &proto.APLActionStrictSequence{Actions: actions},
	}}}
}

func spellAction(id int32) *proto.APLAction {
	return &proto.APLAction{Action: &proto.APLAction_CastSpell{CastSpell: &proto.APLActionCastSpell{
		SpellId: &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func newBear(t *testing.T) (*core.Simulation, *GuardianDruid) {
	return newBearWithTalents(t, "0000000000000000-00000000000000000000-0000000000000000")
}

func newBearWithTalents(t *testing.T, talents string) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	return newBearAgainst(t, talents, 1)
}

func newBearAgainst(t *testing.T, talents string, numTargets int) (*core.Simulation, *GuardianDruid) {
	t.Helper()
	targets := make([]*proto.Target, numTargets)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{StartingRage: 100},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: &proto.ConsumesSpec{},
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*GuardianDruid)
}

func TestDemoralizingRoarReachesEveryTarget(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000000000000-0000000000000000", 3)
	targets := sim.Encounter.ActiveTargetUnits
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		debuffed := 0
		for _, target := range targets {
			if bear.DemoralizingRoarAuras.Get(target).IsActive() {
				debuffed++
			}
		}
		if debuffed == len(targets) {
			return
		}
	}
	t.Error("demoralizing roar never debuffed every target")
}

func TestDemoralizingRoarIgnoresDodge(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	for range 20 {
		bear.AddRage(sim, 50, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
		bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget)
		sim.CurrentTime += 2 * time.Second
		if m := bear.DemoralizingRoar.SpellMetrics[target.UnitIndex]; m.Dodges != 0 {
			t.Fatalf("demoralizing roar was dodged %d times", m.Dodges)
		}
		if bear.DemoralizingRoarAuras.Get(target).IsActive() {
			return
		}
	}
	t.Error("demoralizing roar never landed against a target that dodges every melee attack")
}

func TestBerserkPrimalBiteStrikesThreeTargets(t *testing.T) {
	sim, bear := newBearAgainst(t, "0000000000000000-00000000000100000001-0000000000000000", 5)
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	bear.BerserkAura.Activate(sim)
	bear.AddRage(sim, 100, metrics)
	if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
		t.Fatal("primal bite did not cast under berserk")
	}
	struck := 0
	for _, target := range sim.Encounter.ActiveTargetUnits {
		if m := bear.PrimalBite.SpellMetrics[target.UnitIndex]; m.Hits+m.Crits+m.Misses+m.Dodges+m.Parries+m.Glances+m.Blocks > 0 {
			struck++
		}
	}
	if struck != 3 {
		t.Errorf("primal bite under berserk struck %d targets, want 3", struck)
	}
}

// Bash 8983 states a melee defense type and a miss refund: a dodged Bash stuns nothing and gives rage back.
func TestDodgedBashStunsNothingAndRefunds(t *testing.T) {
	sim, bear := newBear(t)
	target := sim.Encounter.ActiveTargetUnits[0]
	target.PseudoStats.BaseDodgeChance = 1
	before := bear.CurrentRage()
	if !bear.Bash.Cast(sim, bear.CurrentTarget) {
		t.Fatal("bash did not cast")
	}
	if m := bear.Bash.SpellMetrics[target.UnitIndex]; m.Dodges != 1 {
		t.Fatalf("bash dodges = %d, want 1", m.Dodges)
	}
	for _, aura := range bear.Bash.RelatedAuraArrays {
		if aura.Get(target).IsActive() {
			t.Error("a dodged bash stunned the target")
		}
	}
	if spent := before - bear.CurrentRage(); spent >= bear.Bash.DefaultCast.Cost {
		t.Errorf("a dodged bash spent %v of its %v rage", spent, bear.Bash.DefaultCast.Cost)
	}
}

// Berserk effect 2 is -100% on Primal Bite's 6 s cooldown, while Berserk is up.
func TestBerserkLiftsPrimalBiteCooldownOnlyWhileUp(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000000100000001-0000000000000000")
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	cast := func(when string) {
		t.Helper()
		bear.AddRage(sim, 100, metrics)
		if !bear.PrimalBite.Cast(sim, bear.CurrentTarget) {
			t.Fatalf("primal bite did not cast %s", when)
		}
		sim.CurrentTime += 2 * time.Second
	}

	bear.BerserkAura.Activate(sim)
	cast("under berserk")
	if !bear.PrimalBite.IsReady(sim) {
		t.Errorf("primal bite is on cooldown until %v under berserk", bear.PrimalBite.ReadyAt())
	}
	bear.BerserkAura.Deactivate(sim)
	cast("after berserk")
	if got, want := bear.PrimalBite.ReadyAt(), sim.CurrentTime+4*time.Second; got != want {
		t.Errorf("primal bite is ready at %v after berserk, want %v", got, want)
	}
}

func TestFrenziedRegenerationSpendsRageForHealth(t *testing.T) {
	sim, bear := newBear(t)
	bear.AddRage(sim, 100, bear.NewRageMetrics(core.ActionID{SpellID: 1}))
	before := bear.CurrentRage()
	if !bear.FrenziedRegeneration.Cast(sim, bear.CurrentTarget) {
		t.Fatal("frenzied regeneration did not cast")
	}
	for sim.CurrentTime < 1100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if spent := before - bear.CurrentRage(); spent < 10 {
		t.Errorf("frenzied regeneration spent %v rage in its first tick, want 10", spent)
	}
	healed := 0.0
	for _, m := range bear.FrenziedRegeneration.SpellMetrics {
		healed += m.TotalHealing
	}
	if healed <= 0 {
		t.Error("frenzied regeneration healed nothing")
	}
}

func TestFirstLacerateHitsForNothing(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		if bear.Lacerate.Dot(target).IsActive() {
			break
		}
		sim.CurrentTime += 2 * time.Second
	}
	if !bear.Lacerate.Dot(target).IsActive() {
		t.Fatal("lacerate never landed")
	}
	if got := bear.Lacerate.SpellMetrics[target.UnitIndex].TotalDamage; got != 0 {
		t.Errorf("the first lacerate dealt %v, want 0 with no stack already on the target", got)
	}
}

func TestLacerateStacksToFive(t *testing.T) {
	sim, bear := newBear(t)
	if !bear.InForm(druid.Bear) {
		t.Fatal("the bear did not start in bear form")
	}
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 40 {
		bear.AddRage(sim, 50, metrics)
		bear.Lacerate.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Lacerate.Dot(target).GetStacks() >= 5 {
			break
		}
	}
	if got := bear.Lacerate.Dot(target).GetStacks(); got != 5 {
		t.Errorf("lacerate stacks = %d, want 5", got)
	}
}

func TestMaulReplacesTheNextSwing(t *testing.T) {
	sim, bear := newBear(t)
	if bear.TryMaul(sim, nil) != nil {
		t.Fatal("maul replaced a swing before it was queued")
	}
	if !bear.Maul.Cast(sim, bear.CurrentTarget) {
		t.Fatal("maul did not queue")
	}
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	replaced := bear.TryMaul(sim, nil)
	if replaced == nil || replaced.ActionID.SpellID != 9881 {
		t.Fatalf("the queued swing was %v, want maul 9881", replaced)
	}
}

func TestFeralChargeCastsOutsideMelee(t *testing.T) {
	sim, bear := newBearWithTalents(t, "0000000000000000-00000000100000000000-0000000000000000")
	bear.DistanceFromTarget = 10
	if !bear.FeralCharge.Cast(sim, bear.CurrentTarget) {
		t.Fatal("feral charge did not cast at 10 yards")
	}
	if bear.FeralCharge.CD.Timer != nil {
		bear.FeralCharge.CD.Reset()
	}
	bear.DistanceFromTarget = 0
	if bear.FeralCharge.CanCast(sim, bear.CurrentTarget) {
		t.Error("feral charge could be cast in melee")
	}
}

func TestSwipeDoesNotScaleWithAttackPower(t *testing.T) {
	sim, bear := newBear(t)
	target := bear.CurrentTarget
	metrics := bear.NewRageMetrics(core.ActionID{SpellID: 1})
	for range 20 {
		bear.AddRage(sim, 50, metrics)
		bear.Swipe.Cast(sim, target)
		sim.CurrentTime += 2 * time.Second
		if bear.Swipe.SpellMetrics[target.UnitIndex].Hits > 0 {
			break
		}
	}
	hits := bear.Swipe.SpellMetrics[target.UnitIndex]
	if hits.Hits == 0 {
		t.Fatal("swipe never hit")
	}
	perHit := hits.TotalDamage / float64(hits.Hits)
	// The row's damage is 83. Attack power on a bear is far larger than the gap this allows.
	if perHit < 40 || perHit > 160 {
		t.Errorf("swipe hit for %.1f, want the row's damage without attack power", perHit)
	}
}

const bearTalents = "0000000000000000-05030300100100010500-0000000000000000"

func TestBearSimpleRotationDealsDamage(t *testing.T) {
	result := core.RunSim(bearSimpleRequest(), nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	var casts int32
	for _, id := range []int32{1238073, 1235827, 9881, 9908} {
		casts += bearCastsOf(player.Actions, id)
	}
	if casts == 0 {
		t.Fatal("primal bite, lacerate, maul and swipe were never cast")
	}
}

func TestBearShippedAplRuns(t *testing.T) {
	req := bearSimpleRequest()
	req.Raid.Parties[0].Players[0].Rotation = core.GetAplRotation("../../../ui/specs/druid/feralbear/apls", "default").Rotation
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	for _, id := range []int32{1238073, 9881, 22812} {
		if bearCastsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped bear APL never casts %d", id)
		}
	}
}

func bearSimpleRequest() *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Bear", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: bearTalents,
				Equipment:     &proto.EquipmentSpec{Items: items},
				Consumables:   &proto.ConsumesSpec{},
				Buffs:         &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
					Options: &proto.FeralBearDruid_Options{},
				}},
				Rotation: &proto.APLRotation{
					PriorityList: []*proto.APLListItem{{
						Action: &proto.APLAction{
							Action: &proto.APLAction_BearOptimalRotationAction{
								BearOptimalRotationAction: &proto.APLActionBearOptimalRotationAction{
									MaintainFaerieFire:       true,
									MaintainDemoralizingRoar: true,
									MaulRageThreshold:        50,
									SwipeUsage:               proto.FeralBearDruid_Rotation_SwipeUsage_WithEnoughAP,
									SwipeApThreshold:         2700,
								},
							},
						},
					}},
				},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}
}

func bearCastsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetSpellId() != spellID {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../../../ui/specs/druid/feralbear/builds/starter.build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	raid := core.SinglePlayerRaidProto(settings.Player, settings.PartyBuffs, settings.RaidBuffs, settings.Debuffs)
	raid.Tanks = settings.Tanks
	if settings.Encounter == nil {
		settings.Encounter = &proto.Encounter{Duration: core.LongDuration, Targets: []*proto.Target{core.FreshDefaultTargetConfig()}}
	}
	return &proto.RaidSimRequest{Raid: raid, Encounter: settings.Encounter, SimOptions: &proto.SimOptions{Iterations: 1, RandomSeed: 1, IsTest: true}}
}

func TestFeralBearStarterCasts(t *testing.T) {
	request := starterRequest(t)
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			request.Raid.Parties[0].Players[0].Race = race
			stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid})
			for _, action := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
				if len(action.Validations) > 0 {
					t.Fatalf("rotation validation: %v", action.Validations)
				}
			}
			result := core.RunRaidSim(request)
			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			player := result.RaidMetrics.Parties[0].Players[0]
			casts := make(map[int32]int32)
			for _, action := range player.Actions {
				for _, target := range action.Targets {
					casts[action.Id.GetSpellId()] += target.Casts
				}
			}
			for _, id := range []int32{1238073, 1235827, 9881, 9898} {
				if casts[id] == 0 {
					t.Errorf("starter rotation never casts spell %d", id)
				}
			}
			if player.Dtps.Avg <= 0 {
				t.Fatal("assigned bear receives no damage")
			}
			t.Logf("casts=%v DPS=%g DTPS=%g", casts, player.Dps.Avg, player.Dtps.Avg)
		})
	}
}

func TestClearcastingSurvivesFreeFaerieFire(t *testing.T) {
	sim, bear := newBear(t)
	bear.OmenOfClarityAura.Deactivate(sim)
	if cost := bear.FaerieFire.Cost.GetCurrentCost(); cost != 0 {
		t.Fatalf("Faerie Fire in Bear Form costs %g before Clearcasting, want zero", cost)
	}
	bear.ClearcastingAura.Activate(sim)
	bear.ClearcastingAura.SetStacks(sim, bear.ClearcastingAura.MaxStacks)
	if !bear.FaerieFire.Cast(sim, bear.CurrentTarget) {
		t.Fatal("Faerie Fire did not cast")
	}
	if !bear.ClearcastingAura.IsActive() {
		t.Fatal("a spell already free in Bear Form consumed Clearcasting")
	}
}

func TestClearcastingCoversAndSpendsDemoralizingRoar(t *testing.T) {
	sim, bear := newBear(t)
	bear.OmenOfClarityAura.Deactivate(sim)
	before := bear.DemoralizingRoar.Cost.GetCurrentCost()
	bear.ClearcastingAura.Activate(sim)
	bear.ClearcastingAura.SetStacks(sim, bear.ClearcastingAura.MaxStacks)
	covered := bear.DemoralizingRoar.Cost.GetCurrentCost()
	if before <= 0 || covered != 0 {
		t.Fatalf("Demoralizing Roar costs %g before and %g with Clearcasting", before, covered)
	}
	if !bear.DemoralizingRoar.Cast(sim, bear.CurrentTarget) {
		t.Fatal("Demoralizing Roar did not cast")
	}
	t.Logf("cost before=%g, with Clearcasting=%g, charge after cast=%v", before, covered, bear.ClearcastingAura.IsActive())
	if bear.ClearcastingAura.IsActive() {
		t.Fatal("a covered offensive cast did not spend its charge")
	}
}
