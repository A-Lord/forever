//go:build with_db

package feralcat

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/wowsims/forever/assets/database"
	assetsdb "github.com/wowsims/forever/assets/database"
	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
	"github.com/wowsims/forever/sim/druid"
	"github.com/wowsims/forever/sim/druid/feralbear"
	"google.golang.org/protobuf/encoding/protojson"
	goproto "google.golang.org/protobuf/proto"
)

func init() {
	RegisterFeralCatDruid()
	common.RegisterAllEffects()
}

func TestFeralCat(t *testing.T) {
	generators := core.FullCharacterTestSuiteGenerator([]core.CharacterSuiteConfig{starterConfig()})
	generators = append(generators, &core.SingleDpsTestGenerator{Name: "StarterCasts", Request: starterRequest(t)})
	core.RunTestSuite(t, t.Name(), generators)
}

func starterConfig() core.CharacterSuiteConfig {
	reference := proto.Stat_StatAttackPower
	weigh := []proto.Stat{proto.Stat_StatStrength, proto.Stat_StatAgility, proto.Stat_StatStamina}
	config := core.GetTestBuildFromJSON(
		proto.Class_ClassDruid, "../../../ui/specs/druid/feralcat/builds", "starter",
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
	config.OtherRotations = []core.RotationCombo{core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default")}
	return config
}

func BenchmarkSimulate(b *testing.B) {
	core.CharacterBenchmark(b, starterConfig())
}

func TestGenesisIncreasesCatBleedDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0500000000000000-00000000000000000000-0000000000000000", nil)
	for _, bleed := range []struct {
		name       string
		base, with float64
	}{
		{"Rake", base.Rake.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rake.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
		{"Rip", base.Rip.Dot(base.CurrentTarget).PeriodicDamageMultiplier, talented.Rip.Dot(talented.CurrentTarget).PeriodicDamageMultiplier},
	} {
		if got, want := bleed.with/bleed.base, 1.05; math.Abs(got-want) > 1e-9 {
			t.Errorf("%s periodic damage multiplier ratio = %v, want %v", bleed.name, got, want)
		}
	}
}

func TestNaturesMajestyRaisesCatMeleeAndSpellCrit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000200000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Majesty %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}

func TestNaturesReachRaisesCatMeleeAndSpellHit(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000020000000000-00000000000000000000-0000000000000000", nil)
	want := 4.0
	for _, stat := range []stats.Stat{stats.PhysicalHitPercent, stats.SpellHitPercent} {
		if got := talented.GetStat(stat) - base.GetStat(stat); math.Abs(got-want) > 1e-9 {
			t.Errorf("Nature's Reach %s = %v, want %v", stat.StatName(), got, want)
		}
	}
}

func TestNakedDruidAtSixtyAcrossSupportedRaces(t *testing.T) {
	for _, tc := range []struct {
		race                                                  proto.Race
		strength, agility, stamina, intellect, spirit, health float64
	}{
		{proto.Race_RaceNightElf, 62, 65, 69, 100, 110, 1993},
		{proto.Race_RaceTauren, 70, 55, 72, 95, 112, 2023},
		{proto.Race_RaceHighOrderSkyborne, 64, 61, 69, 101, 110, 1993},
		{proto.Race_RaceWindshaperSkyborne, 64, 61, 69, 101, 110, 1993},
	} {
		t.Run(tc.race.String(), func(t *testing.T) {
			// Given
			player := &proto.Player{Class: proto.Class_ClassDruid, Race: tc.race, Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{}}}, Equipment: &proto.EquipmentSpec{}}
			character := core.NewCharacter(&core.Party{Raid: &core.Raid{}}, 0, player)
			naked := druid.New(&character, druid.Humanoid, druid.SelfBuffs{}, "")
			naked.FinalizeStatDeps()
			// When
			sheet := naked.ApplyStatDependencies(naked.GetStats())
			// Then
			for _, want := range []struct {
				stat  stats.Stat
				value float64
			}{
				{stats.Strength, tc.strength}, {stats.Agility, tc.agility}, {stats.Stamina, tc.stamina},
				{stats.Intellect, tc.intellect}, {stats.Spirit, tc.spirit}, {stats.Health, tc.health},
				{stats.AttackPower, -20 + 2*tc.strength},
			} {
				if got := sheet[want.stat]; math.Abs(got-want.value) > 1e-9 {
					t.Errorf("%s = %v, want %v", want.stat.StatName(), got, want.value)
				}
			}
			t.Logf("humanoid strength=%v agility=%v stamina=%v intellect=%v spirit=%v health=%v AP=%v", sheet[stats.Strength], sheet[stats.Agility], sheet[stats.Stamina], sheet[stats.Intellect], sheet[stats.Spirit], sheet[stats.Health], sheet[stats.AttackPower])
		})
	}
}

func TestDruidDodgeFromAgilityAcrossForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{{"humanoid", druid.Humanoid}, {"cat", druid.Cat}, {"bear", druid.Bear}, {"dire bear", druid.DireBear}} {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 0})
			if tc.form != druid.Cat {
				cat.ClearForm(sim)
			}
			fillResource(sim, cat)
			if tc.form == druid.Bear && !cat.BearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("bear form did not cast")
			}
			if tc.form == druid.DireBear && !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("dire bear form did not cast")
			}
			before := cat.GetDodgeFromRating()
			// When
			cat.AddStatDynamic(sim, stats.Agility, 100)
			// Then
			if got := cat.GetDodgeFromRating() - before; math.Abs(got-0.05) > 1e-9 {
				t.Errorf("100 agility gives %v dodge, want 0.05", got)
			}
			if got := cat.PseudoStats.BaseDodgeChance; math.Abs(got-0.019) > 1e-9 {
				t.Errorf("base dodge %v, want 0.019", got)
			}
			t.Logf("form=%v dodge=%v→%v base=%v", tc.form, before, cat.GetDodgeFromRating(), cat.PseudoStats.BaseDodgeChance)
		})
	}
}

func TestThickHideEquippedDefenseFollowsFormsAndStatChanges(t *testing.T) {
	for _, form := range []druid.DruidForm{druid.Cat, druid.Bear, druid.DireBear} {
		t.Run(fmt.Sprintf("form_%d", form), func(t *testing.T) {
			worn := map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotFinger1: 12985}
			sim, cat := newCatWearing(t, "-000003", nil, worn)
			plainSim, plain := newCatWearing(t, "", nil, worn)
			fillResource(sim, cat)
			fillResource(plainSim, plain)
			if form == druid.Bear {
				cat.BearForm.Cast(sim, cat.CurrentTarget)
				plain.BearForm.Cast(plainSim, plain.CurrentTarget)
			}
			if form == druid.DireBear {
				cat.DireBearForm.Cast(sim, cat.CurrentTarget)
				plain.DireBearForm.Cast(plainSim, plain.CurrentTarget)
			}
			multiplier := 1.0
			if form == druid.Bear {
				multiplier = 2.8
			}
			if form == druid.DireBear {
				multiplier = 4.6
			}
			if got, want := cat.GetStat(stats.Armor)-plain.GetStat(stats.Armor), (60*3+4*2)*multiplier; math.Abs(got-want) > 1e-9 {
				t.Errorf("Thick Hide armor with four equipped Defense = %g, want %g", got, want)
			}
			before := cat.GetStat(stats.Armor)
			cat.AddStatDynamic(sim, stats.DefenseRating, 4)
			want := 8.0 * multiplier
			if got := cat.GetStat(stats.Armor) - before; math.Abs(got-want) > 1e-9 {
				t.Errorf("armor from four Defense = %g, want %g", got, want)
			}
			cat.ClearForm(sim)
			before = cat.GetStat(stats.Armor)
			cat.AddStatDynamic(sim, stats.DefenseRating, -4)
			if got := cat.GetStat(stats.Armor) - before; got != 0 {
				t.Errorf("humanoid Thick Hide armor = %g, want 0", got)
			}
		})
	}
}

func TestCatBarkskinCastsAndAppliesItsDamageReduction(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	row := spelldata.MustFind(22812)
	spell := cat.GetSpell(core.ActionID{SpellID: row.ID})
	if spell == nil || !spell.Cast(sim, &cat.Unit) {
		t.Fatal("Cat could not cast Barkskin")
	}
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; math.Abs(got-0.8) > 1e-9 {
		t.Errorf("Barkskin damage multiplier = %g, want 0.8", got)
	}
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire]; got != 1 {
		t.Errorf("Barkskin fire damage multiplier = %g, want 1", got)
	}
	cat.GetAura(row.Name).Deactivate(sim)
	if got := cat.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexPhysical]; got != 1 {
		t.Errorf("expired Barkskin damage multiplier = %g, want 1", got)
	}
}

func TestGenesisThreePieceIncludesDefenseAndArmor(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHead: 21353, proto.ItemSlot_ItemSlotShoulder: 21354, proto.ItemSlot_ItemSlotChest: 21357,
	})
	aura := cat.GetAura("Genesis Raiment 3P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Genesis three-piece is inactive")
	}
	armor := cat.GetStat(stats.Armor)
	if got := cat.GetStat(stats.DefenseRating); got != 15 {
		t.Errorf("Genesis Defense = %g, want 15", got)
	}
	aura.Deactivate(sim)
	if got := cat.GetStat(stats.DefenseRating); got != 0 {
		t.Errorf("expired Genesis Defense = %g, want 0", got)
	}
	if got := armor - cat.GetStat(stats.Armor); got != 150 {
		t.Errorf("Genesis armor = %g, want 150", got)
	}
}

// Feral charge, shifting power, primal bite and berserk, plus furor and natural shapeshifter so the
// form spells can both be paid for. Prowl is castable only before the pull, so each opener is its own
// request: one prowls into Pounce, the other into Ravage. Both then leave cat form and come back so
// both form spells are casts.
const coverageTalents = "0000000000000000-00000000101100000001-0500300000000000"

func TestFeralCatCoverage(t *testing.T) {
	core.RunTestSuite(t, t.Name(), []core.TestGenerator{castsGen{
		{"Casts", catCoverageRequest(9827)},
		{"RavageCasts", catCoverageRequest(9867)},
	}})
}

type namedRequest struct {
	name string
	rsr  *proto.RaidSimRequest
}

type castsGen []namedRequest

func (g castsGen) NumTests() int { return len(g) }

func (g castsGen) GetTest(i int) (string, *proto.ComputeStatsRequest, *proto.StatWeightsRequest, *proto.RaidSimRequest) {
	return g[i].name, nil, nil, g[i].rsr
}

func catCoverageRequest(opener int32) *proto.RaidSimRequest {
	// The opener while prowling, and nothing else while prowling so an auto-replace does not spend the
	// stealth. The form spells sit under the builders.
	prowling := auraUp(9913)
	return coverageRequest(false, &proto.APLRotation{
		Type: proto.APLRotation_TypeAPL,
		PrepullActions: []*proto.APLPrepullAction{{
			Action:    spellAction(9913),
			DoAtValue: &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: "-1s"}}},
		}},
		PriorityList: []*proto.APLListItem{
			castWhen(opener, prowling),
			castWhen(417141, not(prowling)),
			castWhen(1322605, not(prowling)),
			// Energy ticks jump over the gaps between these costs, so each builder and finisher
			// owns a slice of the fight. Claw also builds combo points for the finishers.
			castWhen(9830, and(not(prowling), timeBetween("0.2", "0.4"))),
			castWhen(9904, and(not(prowling), timeBetween("0.4", "0.6"))),
			castWhen(22829, and(not(prowling), timeBetween("0.6", "0.8"))),
			castWhen(9896, and(not(prowling), timeAtLeast("0.8"))),
			castWhen(9850, and(not(prowling), or(timeBelow("0.2"), timeAtLeast("0.6")))),
			// A late shift into dire bear spends the mana needed to shift back, so the form
			// spells only run while claw still owns the fight.
			castWhen(9634, and(not(prowling), timeBelow("0.2"))),
			castWhen(768, and(not(prowling), timeBelow("0.2"))),
			castWhen(9892, not(prowling)),
			castWhen(9821, not(prowling)),
			castWhen(9907, and(not(prowling), not(auraOnTarget(9907)))),
		},
	})
}

func coverageRequest(bear bool, rotation *proto.APLRotation) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	player := &proto.Player{
		Name: "Feral", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
		TalentsString: coverageTalents,
		Equipment:     &proto.EquipmentSpec{Items: items},
		Consumables:   &proto.ConsumesSpec{},
		Buffs:         &proto.IndividualBuffs{},
		Rotation:      rotation,
	}
	if bear {
		player.Spec = &proto.Player_FeralBearDruid{FeralBearDruid: &proto.FeralBearDruid{
			Options: &proto.FeralBearDruid_Options{StartingRage: 100},
		}}
	} else {
		player.Spec = &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
			Options: &proto.FeralCatDruid_Options{},
		}}
	}
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter:  &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 63}}, Duration: 180},
	}
}

func cast(id int32) *proto.APLListItem {
	return castWhen(id, nil)
}

func castWhen(id int32, condition *proto.APLValue) *proto.APLListItem {
	action := spellAction(id)
	action.Condition = condition
	return &proto.APLListItem{Action: action}
}

func auraUp(id int32) *proto.APLValue {
	return auraOn(id, nil)
}

func auraOnTarget(id int32) *proto.APLValue {
	return auraOn(id, &proto.UnitReference{Type: proto.UnitReference_CurrentTarget})
}

func auraOn(id int32, source *proto.UnitReference) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_AuraIsActive{AuraIsActive: &proto.APLValueAuraIsActive{
		SourceUnit: source,
		AuraId:     &proto.ActionID{RawId: &proto.ActionID_SpellId{SpellId: id}},
	}}}
}

func not(value *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Not{Not: &proto.APLValueNot{Val: value}}}
}

func and(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_And{And: &proto.APLValueAnd{Vals: values}}}
}

func or(values ...*proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Or{Or: &proto.APLValueOr{Vals: values}}}
}

func timeBelow(fraction string) *proto.APLValue {
	return cmp(proto.APLValueCompare_OpLt, timePercent(), konst(fraction))
}

func timeAtLeast(fraction string) *proto.APLValue {
	return cmp(proto.APLValueCompare_OpGe, timePercent(), konst(fraction))
}

func timeBetween(low, high string) *proto.APLValue {
	return and(timeAtLeast(low), timeBelow(high))
}

func timePercent() *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_CurrentTimePercent{CurrentTimePercent: &proto.APLValueCurrentTimePercent{}}}
}

func konst(val string) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Const{Const: &proto.APLValueConst{Val: val}}}
}

func cmp(op proto.APLValueCompare_ComparisonOperator, lhs, rhs *proto.APLValue) *proto.APLValue {
	return &proto.APLValue{Value: &proto.APLValue_Cmp{Cmp: &proto.APLValueCompare{Op: op, Lhs: lhs, Rhs: rhs}}}
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

func TestCatBuildsUseNativeRaidTargetArmor(t *testing.T) {
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			request := buildRequest(t, name)
			target := request.Encounter.Targets[0]
			defaultTarget := core.FreshDefaultTargetConfig()
			if target.Level != core.CharacterLevel+3 || target.Stats[stats.Armor] != defaultTarget.Stats[stats.Armor] || target.Stats[stats.Health] != defaultTarget.Stats[stats.Health] {
				t.Fatalf("Cat build overrides native raid target level, armor, or health: %v", target)
			}
			armored := core.RunRaidSim(request)
			unarmoredRequest := goproto.Clone(request).(*proto.RaidSimRequest)
			unarmoredRequest.Encounter.Targets[0].Stats[stats.Armor] = 0
			unarmored := core.RunRaidSim(unarmoredRequest)
			if armored.Error != nil || unarmored.Error != nil {
				t.Fatalf("armor comparison failed: %v / %v", armored.Error, unarmored.Error)
			}
			if armored.RaidMetrics.Dps.Avg >= unarmored.RaidMetrics.Dps.Avg {
				t.Fatalf("native armor did not reduce Cat damage: armored %.3f, unarmored %.3f", armored.RaidMetrics.Dps.Avg, unarmored.RaidMetrics.Dps.Avg)
			}
		})
	}
}

func TestMaintainFaerieFireUsesSharedDebuffRefresh(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{FinishingMove: proto.FeralCatDruid_Rotation_Rip, RipMinComboPoints: 5, MaintainFaerieFire: true})
	for _, external := range []bool{false, true} {
		request := catPowershiftRequest(powershiftTalents, rotation, false)
		request.Raid.Debuffs = &proto.Debuffs{FaerieFire: external}
		player := runCat(t, request)
		casts := castsOf(player.Actions, 9907)
		if external && casts != 0 {
			t.Errorf("Cat replaced an active external Faerie Fire %d times", casts)
		}
		if !external && casts == 0 {
			t.Error("Cat did not supply the missing Faerie Fire")
		}
	}
}

func TestRevelationIgnoresZeroDamageFaerieFire(t *testing.T) {
	// Given a Cat wearing Revelation and a fixed seed.
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: 866, Enchant: 8217}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{{
			Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
			Spec:      &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{Options: &proto.FeralCatDruid_Options{}}},
			Equipment: &proto.EquipmentSpec{Items: items}, Buffs: &proto.IndividualBuffs{}, Rotation: &proto.APLRotation{},
		}}, Buffs: &proto.PartyBuffs{}}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)
	buff := cat.GetAuraByID(core.ActionID{SpellID: 1323400})
	if buff == nil {
		t.Fatal("Revelation buff is not registered")
	}

	// When zero-damage Faerie Fire applications are delivered through the real proc listener.
	for casts := 0; casts < 300 && !buff.IsActive(); casts++ {
		cat.FaerieFire.SkipCastAndApplyEffects(sim, cat.CurrentTarget)
		sim.Step()
	}

	// Then no Revelation charge is granted.
	t.Logf("zero-damage Faerie Fire activated Revelation=%v at %v", buff.IsActive(), sim.CurrentTime)
	if buff.IsActive() {
		t.Fatal("zero-damage Faerie Fire activated Revelation")
	}
}

func TestRaidDefaultRunsSelectedCatRotation(t *testing.T) {
	player := runCat(t, buildRequest(t, "raid"))
	for _, id := range []int32{1322605, 9830, 9896, 9904, 417141, 9907} {
		if castsOf(player.Actions, id) == 0 {
			t.Errorf("raid default never casts spell %d", id)
		}
	}
	for _, id := range []int32{768, 9634, 22829} {
		if got := castsOf(player.Actions, id); got != 0 {
			t.Errorf("raid default casts spell %d %d times", id, got)
		}
	}
}

func defaultCatRotation() *proto.APLRotation {
	return catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
		FaerieFireForOmen:  true,
	})
}

func TestRotationLeavesCatFormOnlyForEngineeringExplosives(t *testing.T) {
	for _, c := range []struct {
		name       string
		itemID     int32
		leavesForm bool
	}{{"Goblin Sapper Charge", 10646, true}, {"EZ-Thro Dark Bomb", 260817, false}} {
		request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
		cat := request.Raid.Parties[0].Players[0]
		cat.Profession1 = proto.Profession_Engineering
		cat.Consumables = &proto.ConsumesSpec{ExtraIds: []int32{c.itemID}}
		player := runCat(t, request)

		thrown := itemCastsOf(player.Actions, c.itemID)
		if thrown == 0 {
			t.Errorf("%s was never thrown", c.name)
		}
		want := int32(0)
		if c.leavesForm {
			want = thrown
		}
		if got := castsOf(player.Actions, 768); got != want {
			t.Errorf("%s: cat form casts = %d after %d throws, want %d", c.name, got, thrown, want)
		}
	}
}

func TestBerserkWaitsForADecisionItsCritCovers(t *testing.T) {
	request := catPowershiftRequest(powershiftTalents, defaultCatRotation(), false)
	player := request.Raid.Parties[0].Players[0]
	player.Profession1 = proto.Profession_Engineering
	player.Consumables = &proto.ConsumesSpec{ExtraIds: []int32{10646}}
	sim := core.NewSim(request, simsignals.CreateSignals())
	sim.Reset()
	cat := sim.Raid.Parties[0].Players[0].(*FeralDruid)

	berserk := cat.GetMajorCooldown(cat.Berserk.ActionID)
	if berserk == nil {
		t.Fatal("Berserk is not a major cooldown")
	}
	ready := func() bool { return berserk.ShouldActivate(sim, cat.GetCharacter()) }
	if ready() {
		t.Error("Berserk is ready while the sapper waits to be thrown")
	}

	sapper := cat.GetSpell(core.ActionID{ItemID: 10646})
	if !sapper.Cast(sim, cat.CurrentTarget) {
		t.Fatal("the sapper was not thrown")
	}
	if !cat.InForm(druid.Cat) && !cat.CatForm.Cast(sim, nil) {
		t.Fatal("Cat Form was not cast after the throw")
	}
	if ready() {
		t.Error("Berserk is ready inside Cat Form's GCD")
	}

	for !cat.GCD.IsReady(sim) {
		sim.Step()
	}
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !ready() {
		t.Error("Berserk is not ready with the GCD free, the sapper thrown and a full bar")
	}

	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if ready() {
		t.Error("Berserk is ready on a bar that cannot pay for the builder")
	}

	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if cat.Rip.CurDot().IsActive() || !cat.ripNow(sim, cat.CurrentTarget) {
		t.Fatal("the decision would not cast Rip")
	}
	if ready() {
		t.Error("Berserk is ready when the decision casts Rip")
	}
}

func itemCastsOf(actions []*proto.ActionMetrics, itemID int32) int32 {
	var n int32
	for _, action := range actions {
		if action.GetId().GetItemId() != itemID || action.GetId().GetTag() != 0 {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func TestShiftingPowerWaitsLongerWithClearcastingUp(t *testing.T) {
	sim, cat := newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	cat.AddEnergy(sim, 45, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.wantShift(sim, cat.CurrentTarget) {
		t.Fatal("no shift at 45 energy without Clearcasting")
	}
	cat.ClearcastingAura.Activate(sim)
	if cat.wantShift(sim, cat.CurrentTarget) {
		t.Error("a shift at 45 energy with Clearcasting up, which caps the bar during the free cast")
	}
}

func TestFaerieFireFishesForOmenOnlyWhenItCanProc(t *testing.T) {
	sim, cat := newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	cat.ShiftingPower.CD.Use(sim)
	target := cat.CurrentTarget
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing with the option off")
	}

	cat.setupRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		FaerieFireForOmen:  true,
	})
	if !cat.wantOmenFishing(sim, target) {
		t.Fatal("no fishing with the option on, an empty bar and Omen ready")
	}

	energy := cat.NewEnergyMetrics(core.ActionID{SpellID: 1})
	cat.AddEnergy(sim, cat.builder().Cost.GetCurrentCost()-1, energy)
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing when the builder is affordable before the GCD ends")
	}
	emptyEnergy(sim, cat)

	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	cat.AddEnergy(sim, cat.Rip.Cost.GetCurrentCost()-1, energy)
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing when Rip is affordable before the GCD ends")
	}
	emptyEnergy(sim, cat)
	cat.SpendComboPoints(sim, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))

	cat.ShiftingPower.CD.Reset()
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing when Shifting Power is ready before the GCD ends")
	}
	cat.ShiftingPower.CD.Use(sim)
	cat.ClearcastingAura.Activate(sim)
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing with Clearcasting already up")
	}
	cat.ClearcastingAura.Deactivate(sim)
	cat.OmenOfClarityAura.Icd.Use(sim)
	if cat.wantOmenFishing(sim, target) {
		t.Error("fishing while Omen of Clarity is on its ICD")
	}
}

func prepareFeralDamageCase(sim *core.Simulation, cat *FeralDruid) {
	target := cat.CurrentTarget
	target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
	cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
	cat.AddStatDynamic(sim, stats.ExpertisePercent, 100)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	fillResource(sim, cat)
}

func TestRipSnapshotsComboPointsOneThroughFive(t *testing.T) {
	for comboPoints := int32(1); comboPoints <= 5; comboPoints++ {
		t.Run(string(rune('0'+comboPoints)), func(t *testing.T) {
			// Given a cat with a fixed number of combo points.
			sim, cat := newCat(t, "", nil)
			prepareFeralDamageCase(sim, cat)
			cat.AddComboPoints(sim, comboPoints, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))

			// When Rip is applied.
			dot := cat.Rip.Dot(cat.CurrentTarget)
			dot.Apply(sim)

			// Then the row's 15 base and 25.5 per point are captured at application.
			want := 15 + 25.5*float64(comboPoints)
			if dot.SnapshotBaseDamage != want {
				t.Errorf("%d-point Rip snapshot = %v, want %v", comboPoints, dot.SnapshotBaseDamage, want)
			}
		})
	}
}

func TestRipReadsAttackPowerAtEachTickAndCapsItsPlaceholderAtFourPoints(t *testing.T) {
	tickDamage := func(attackPower float64) float64 {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		cat.AddStatDynamic(sim, stats.AttackPower, attackPower)
		dot.TickOnce(sim)
		return cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}

	// Given two identical five-point Rips, with attack power added after application.
	withoutAP := tickDamage(0)
	withAP := tickDamage(100)

	// When each Rip ticks, then the assumed 0.01 coefficient uses current attack power and four points.
	if got, want := withAP-withoutAP, 4.0; math.Abs(got-want) > 1e-9 {
		t.Errorf("five-point Rip gained %v tick damage from 100 attack power, want %v", got, want)
	}
}

func TestRakeUsesAttackPowerForTicksOnly(t *testing.T) {
	damage := func(attackPowerBeforeCast float64, attackPowerAfterCast float64) (float64, float64) {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddStatDynamic(sim, stats.AttackPower, attackPowerBeforeCast)
		if !cat.Rake.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Rake did not cast")
		}
		metric := &cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex]
		direct := metric.TotalDamage
		cat.AddStatDynamic(sim, stats.AttackPower, attackPowerAfterCast)
		cat.Rake.Dot(cat.CurrentTarget).TickOnce(sim)
		return direct, metric.TotalDamage - direct
	}

	baseDirect, baseTick := damage(0, 0)
	preCastDirect, preCastTick := damage(1000, 0)
	postCastDirect, postCastTick := damage(0, 1000)

	if preCastDirect != baseDirect || postCastDirect != baseDirect {
		t.Errorf("Rake direct damage with attack power before/after cast = %v/%v, want unchanged %v", preCastDirect, postCastDirect, baseDirect)
	}
	// Rank 1 observations fit 0.055 attack power per tick. Higher ranks keep that coefficient as an
	// assumption, and whether Rake should snapshot attack power remains open.
	for _, result := range []struct {
		timing string
		tick   float64
	}{
		{"before cast", preCastTick},
		{"after application", postCastTick},
	} {
		if got, want := result.tick-baseTick, 55.0; math.Abs(got-want) > 1e-9 {
			t.Errorf("Rake tick gained %v damage from 1000 attack power added %s, want %v", got, result.timing, want)
		}
	}
}

func TestGenesisReachesTheTriggeredPounceBleed(t *testing.T) {
	tickDamage := func(talents string) float64 {
		sim, cat := newCat(t, talents, nil)
		prepareFeralDamageCase(sim, cat)
		dot := cat.PounceBleed.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		dot.TickOnce(sim)
		return cat.PounceBleed.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage
	}

	// Given Pounce's triggered bleed with and without Genesis 5/5.
	base := tickDamage("")
	genesis := tickDamage("0500000000000000-00000000000000000000-0000000000000000")

	// When the bleed ticks, then the talent's parent-Pounce mask raises it by 5%.
	if got, want := genesis/base, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("Genesis Pounce bleed multiplier = %v, want %v", got, want)
	}
}

func TestPounceTalentsModifyOnlyTheirOwnEffects(t *testing.T) {
	const brutalImpactTalents = "0000000000000000-00002000000000000000-0000000000000000"

	castPounce := func(talents string) (float64, time.Duration, time.Duration, time.Duration) {
		sim, cat := newCat(t, talents, nil)
		if talents == brutalImpactTalents && cat.Talents.BrutalImpact != 2 {
			t.Fatalf("Brutal Impact rank = %d, want 2", cat.Talents.BrutalImpact)
		}
		prepareFeralDamageCase(sim, cat)
		sim.CurrentTime = -time.Second
		if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Prowl did not cast before the pull")
		}
		sim.CurrentTime = 0
		fillResource(sim, cat)
		if !cat.Pounce.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Pounce did not cast from Prowl")
		}
		bleed := cat.PounceBleed.Dot(cat.CurrentTarget)
		bleed.TickOnce(sim)
		stun := cat.CurrentTarget.GetAura("Pounce Stun")
		if stun == nil || !stun.IsActive() {
			t.Fatal("Pounce stun did not activate")
		}
		return cat.PounceBleed.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage,
			stun.RemainingDuration(sim), bleed.RemainingDuration(sim), bleed.BaseDurationFlat
	}

	baseDamage, baseStun, baseBleed, _ := castPounce("")
	genesisDamage, _, _, _ := castPounce("0500000000000000-00000000000000000000-0000000000000000")
	_, brutalStun, brutalBleed, brutalBleedExtension := castPounce(brutalImpactTalents)

	if got, want := genesisDamage/baseDamage, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("Genesis Pounce bleed multiplier = %v, want %v", got, want)
	}
	if got, want := brutalStun-baseStun, time.Second; got != want {
		t.Errorf("Brutal Impact Pounce stun extension = %v, want %v", got, want)
	}
	if brutalBleed != baseBleed {
		t.Errorf("Pounce bleed duration with Brutal Impact = %v, want unchanged %v", brutalBleed, baseBleed)
	}
	if brutalBleedExtension != 0 {
		t.Errorf("Pounce bleed duration extension with Brutal Impact = %v, want 0", brutalBleedExtension)
	}
}

func TestRendAndTearReadsBleedsPerTarget(t *testing.T) {
	biteDamage := func(targetIndex int) float64 {
		sim, cat := newCatWearingTargets(t, "0000000000000000-00000000000000000050-0000000000000000", nil, nil, 2)
		target := sim.Environment.GetTargetUnitByIndex(int32(targetIndex))
		target.AddStatDynamic(sim, stats.Armor, -target.GetStat(stats.Armor))
		cat.AddStatDynamic(sim, stats.PhysicalHitPercent, 100)
		cat.AddStatDynamic(sim, stats.ExpertisePercent, 100)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		fillResource(sim, cat)
		cat.Rake.Dot(sim.Environment.GetTargetUnitByIndex(0)).Apply(sim)
		cat.AddComboPoints(sim, 1, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		if !cat.FerociousBite.Cast(sim, target) {
			t.Fatal("Ferocious Bite did not cast")
		}
		return cat.FerociousBite.SpellMetrics[target.UnitIndex].TotalDamage
	}

	// Given two targets where only the first has a tagged Druid bleed.
	bleeding := biteDamage(0)
	unbled := biteDamage(1)

	// When Rend and Tear prices Bite, then its 10% bonus applies only to the bleeding target.
	if got, want := bleeding/unbled, 1.1; math.Abs(got-want) > 1e-9 {
		t.Errorf("Rend and Tear multiplier = %v, want %v", got, want)
	}
}

func TestFerociousBiteScalesWithComboPointsAndSpendsAllExtraEnergy(t *testing.T) {
	biteDamage := func(comboPoints int32, attackPower float64, clearcasting bool) (float64, float64, bool) {
		sim, cat := newCat(t, "", nil)
		prepareFeralDamageCase(sim, cat)
		cat.AddComboPoints(sim, comboPoints, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		cat.AddStatDynamic(sim, stats.AttackPower, attackPower)
		if clearcasting {
			cat.ClearcastingAura.Activate(sim)
			cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)
		}
		if !cat.FerociousBite.Cast(sim, cat.CurrentTarget) {
			t.Fatal("Ferocious Bite did not cast")
		}
		energy := cat.CurrentEnergy()
		for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
		}
		return cat.FerociousBite.SpellMetrics[cat.CurrentTarget.UnitIndex].TotalDamage,
			energy, cat.ClearcastingAura.IsActive()
	}

	for comboPoints := int32(1); comboPoints <= 5; comboPoints++ {
		// Given equal Bites whose only damage input difference is 100 attack power.
		withoutAP, energy, _ := biteDamage(comboPoints, 0, false)
		withAP, _, _ := biteDamage(comboPoints, 100, false)

		// When they land, then the assumed 0.03 coefficient scales once per combo point.
		if got, want := withAP-withoutAP, 3*float64(comboPoints); math.Abs(got-want) > 1e-9 {
			t.Errorf("%d-point Bite gained %v damage from 100 attack power, want %v", comboPoints, got, want)
		}
		if energy != 0 {
			t.Errorf("%d-point Bite left %v energy, want 0", comboPoints, energy)
		}
	}

	// Given Clearcasting at full energy, the base cost is free and all 100 energy remains extra.
	regular, _, _ := biteDamage(5, 0, false)
	clear, energy, active := biteDamage(5, 0, true)

	// When Bite lands, then the extra 35 energy adds 87.5 damage and the cast consumes the charge.
	if got, want := clear-regular, 87.5; math.Abs(got-want) > 1e-9 {
		t.Errorf("Clearcasting Bite bonus = %v, want %v", got, want)
	}
	if energy != 0 || active {
		t.Errorf("Clearcasting Bite left energy=%v active=%v, want 0 and false", energy, active)
	}
}

func TestFeralFormsUseFeralSkillAndRestoreEquippedSkill(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 866})
	cat.ClearForm(sim)
	cat.PseudoStats.WeaponSkills[proto.WeaponSkill_WeaponSkillStaves] = -20
	cat.SetFeralCombat(false, 0)
	table := cat.AttackTables[cat.CurrentTarget.UnitIndex]
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("humanoid staff miss = %v, want 0.09", got)
	}

	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Cat Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	sim.CurrentTime += 2 * time.Second
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	if got := table.MH.BaseMissChance; got != 0.05 {
		t.Fatalf("Bear Form miss = %v, want 0.05", got)
	}

	cat.ClearForm(sim)
	if got := table.MH.BaseMissChance; got != 0.09 {
		t.Fatalf("restored humanoid staff miss = %v, want 0.09", got)
	}
}

func TestFeralAttackPowerConvertsOneForOneOnlyInFeralForms(t *testing.T) {
	// Given a cat with 77 Feral Attack Power added after initialization.
	sim, cat := newCat(t, "", nil)
	catAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)

	// When the druid moves through Cat, humanoid, Bear, and humanoid forms.
	if got := cat.GetStat(stats.AttackPower) - catAP; got != 77 {
		t.Errorf("Cat Form gained %v attack power, want 77", got)
	}
	cat.ClearForm(sim)
	humanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != humanoidAP {
		t.Errorf("removing Feral Attack Power in humanoid form changed attack power from %v to %v", humanoidAP, got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	bearAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := bearAP - cat.GetStat(stats.AttackPower); got != 77 {
		t.Errorf("Bear Form gained %v attack power, want 77", got)
	}
	cat.AddStatDynamic(sim, stats.FeralAttackPower, 77)
	cat.ClearForm(sim)

	// Then the form-only dependency is removed again on exit.
	secondHumanoidAP := cat.GetStat(stats.AttackPower)
	cat.AddStatDynamic(sim, stats.FeralAttackPower, -77)
	if got := cat.GetStat(stats.AttackPower); got != secondHumanoidAP {
		t.Errorf("removing Feral Attack Power after Bear changed humanoid attack power from %v to %v", secondHumanoidAP, got)
	}
}

func TestFurorKeepsPullEnergyOnFirstEntry(t *testing.T) {
	// Given a newly reset cat with Furor 5/5.
	_, cat := newCat(t, furorTalents, nil)

	// When its starting Cat Form activates, then it keeps the energy bar's pull value.
	if got := cat.CurrentEnergy(); got != 100 {
		t.Errorf("energy on first Cat Form entry = %v, want 100", got)
	}
}

func TestFurorRegeneratesStoredEnergyWhileFormIsCancelled(t *testing.T) {
	// Given a Furor cat that cancels form at 40 energy.
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.ClearForm(sim)
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)

	// When it returns after three seconds in humanoid form.
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Form did not cast after cancellation")
	}

	// Then the supported assumption restores 40 stored plus 30 regenerated energy.
	if got := cat.CurrentEnergy(); got != 70 {
		t.Errorf("energy after cancelling Cat Form for three seconds = %v, want 70", got)
	}
}

func TestNaturalReactionOnlyRestoresRageInBearForm(t *testing.T) {
	// Given Natural Reaction 5/5 and a dodged incoming attack while in Cat Form.
	sim, cat := newCat(t, "0000000000000000-00000000000000000500-0000000000000000", nil)
	naturalReaction := cat.GetAura("Natural Reaction")
	if naturalReaction == nil || naturalReaction.OnSpellHitTaken == nil {
		t.Fatal("Natural Reaction listener is missing")
	}
	cat.AutoAttacks.CancelAutoSwing(sim)
	cat.CurrentTarget.AutoAttacks.CancelAutoSwing(sim)
	dodge := &core.SpellResult{Target: &cat.Unit, Outcome: core.OutcomeDodge}
	naturalReaction.OnSpellHitTaken(naturalReaction, sim, cat.AutoAttacks.MHAuto(), dodge)
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if got := cat.CurrentRage(); got != 0 {
		t.Errorf("Natural Reaction restored %v rage in Cat Form, want 0", got)
	}

	// When the same dodge occurs in Bear Form, then the triggered row restores 5 rage.
	fillResource(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Bear Form did not cast")
	}
	cat.SpendRage(sim, cat.CurrentRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
	naturalReaction.OnSpellHitTaken(naturalReaction, sim, cat.AutoAttacks.MHAuto(), dodge)
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if got := cat.CurrentRage(); got != 5 {
		t.Errorf("Natural Reaction restored %v rage in Bear Form, want 5", got)
	}
}

func TestOmenUsesTwoPPMWithTenSecondICDOnMeleeHits(t *testing.T) {
	// Given Omen's permanent listener and a landed main-hand auto result.
	sim, cat := newCat(t, "0000000000000000-00000000000000000000-0000000000000001", nil)
	omen := cat.GetAura("Omen of Clarity")
	if omen == nil || omen.Icd == nil {
		t.Fatalf("Omen listener is incomplete: aura=%v ICD=%v", omen, omen != nil && omen.Icd != nil)
	}
	if omen.Icd.Duration != 10*time.Second {
		t.Fatalf("Omen ICD = %v, want 10s", omen.Icd.Duration)
	}
	cat.AutoAttacks.CancelAutoSwing(sim)
	result := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeHit, Damage: 1}
	before := omen.Icd.ReadyAt()
	wrongMask := cat.RegisterSpell(core.SpellConfig{ActionID: core.ActionID{SpellID: 2}, ProcMask: core.ProcMaskRangedAuto})

	// When a ranged hit reaches the listener directly, then its decoded mask rejects it without using the ICD.
	for range 200 {
		omen.OnSpellHitDealt(omen, sim, wrongMask.Spell, result)
	}
	if got := omen.Icd.ReadyAt(); got != before {
		t.Fatalf("a ranged hit moved Omen's ICD from %v to %v", before, got)
	}

	// When enough eligible one-second paw hits roll 2 PPM on the paw's speed, one activates Clearcasting.
	for range 500 {
		omen.OnSpellHitDealt(omen, sim, cat.AutoAttacks.MHAuto(), result)
	}
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if !cat.ClearcastingAura.IsActive() {
		t.Fatal("500 eligible paw hits produced no Clearcasting proc")
	}
	if got := omen.Icd.ReadyAt(); got != 10*time.Second {
		t.Errorf("Omen ICD ready at %v, want 10s", got)
	}

	// Then an eligible hit inside the ICD cannot reactivate a consumed Clearcasting aura.
	cat.ClearcastingAura.Deactivate(sim)
	for range 500 {
		omen.OnSpellHitDealt(omen, sim, cat.AutoAttacks.MHAuto(), result)
	}
	for end := sim.CurrentTime + core.SpellBatchWindow; sim.CurrentTime <= end && !sim.Step(); {
	}
	if cat.ClearcastingAura.IsActive() {
		t.Error("Omen proc bypassed its 10-second ICD")
	}
}

func TestOmenSpecialHitsUseFormSwingSpeed(t *testing.T) {
	const omenTalents = "0000000000000000-00000000000000000000-0000000000000001"
	const slowWeapon = 1607

	countProcs := func(t *testing.T, form druid.DruidForm) (float64, int) {
		t.Helper()
		sim, cat := newCatWearing(t, omenTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: slowWeapon})
		cat.AutoAttacks.CancelAutoSwing(sim)
		omen := cat.GetAura("Omen of Clarity")
		if omen == nil || omen.OnSpellHitDealt == nil || omen.Icd == nil {
			t.Fatal("Omen listener is incomplete")
		}

		special := cat.Claw.Spell
		switch form {
		case druid.Bear:
			fillResource(sim, cat)
			if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("Bear Form did not cast")
			}
			special = cat.Swipe.Spell
		case druid.Humanoid:
			cat.ClearForm(sim)
		}
		result := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeHit, Damage: 1}
		procs := 0
		for range 500 {
			omen.OnSpellHitDealt(omen, sim, special, result)
			if !omen.Icd.IsReady(sim) {
				procs++
				omen.Icd.Reset()
			}
		}
		if procs == 0 {
			t.Fatal("500 eligible special hits produced no Omen proc")
		}
		return cat.AutoAttacks.MH().SwingSpeed, procs
	}

	catSpeed, catProcs := countProcs(t, druid.Cat)
	bearSpeed, bearProcs := countProcs(t, druid.Bear)
	weaponSpeed, weaponProcs := countProcs(t, druid.Humanoid)
	if catSpeed != 1 {
		t.Errorf("Cat paw speed = %g, want 1", catSpeed)
	}
	if bearSpeed != 2.5 {
		t.Errorf("Bear paw speed = %g, want 2.5", bearSpeed)
	}
	if weaponSpeed != 3.8 {
		t.Errorf("humanoid weapon speed = %g, want 3.8", weaponSpeed)
	}
	t.Logf("Omen procs from 500 real special hits: Cat=%d Bear=%d humanoid weapon=%d", catProcs, bearProcs, weaponProcs)
	if !(catProcs < bearProcs && bearProcs < weaponProcs) {
		t.Fatalf("special Omen rolls did not follow Cat/Bear paw/weapon pace: Cat=%d Bear=%d weapon=%d", catProcs, bearProcs, weaponProcs)
	}
}

func TestClearcastingIsSpentByACoveredCatAbility(t *testing.T) {
	// Given Clearcasting at full energy.
	sim, cat := newCat(t, "", nil)
	prepareFeralDamageCase(sim, cat)
	cat.AutoAttacks.CancelAutoSwing(sim)
	cat.ClearcastingAura.Activate(sim)
	cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)

	// When Rake casts, then its covered class flags make the cast free and consume the charge.
	if !cat.Rake.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Rake did not cast")
	}
	for sim.CurrentTime <= core.SpellBatchWindow && !sim.Step() {
	}
	if got := cat.CurrentEnergy(); got != 100 {
		t.Errorf("Clearcasting Rake left %v energy, want 100", got)
	}
	if cat.ClearcastingAura.IsActive() {
		t.Error("Rake did not consume Clearcasting")
	}

	// Given another charge, when the covered ability misses, then the charge is retained.
	cat.ClearcastingAura.Activate(sim)
	cat.ClearcastingAura.SetStacks(sim, cat.ClearcastingAura.MaxStacks)
	miss := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeMiss}
	cat.ClearcastingAura.OnSpellHitDealt(cat.ClearcastingAura, sim, cat.Rake.Spell, miss)
	if !cat.ClearcastingAura.IsActive() || cat.ClearcastingAura.GetStacks() != 1 {
		t.Error("a missed Rake consumed Clearcasting")
	}

	// When an auto attack lands, then the ability-only class mask also keeps the charge.
	landed := &core.SpellResult{Target: cat.CurrentTarget, Outcome: core.OutcomeHit, Damage: 1}
	cat.ClearcastingAura.OnSpellHitDealt(cat.ClearcastingAura, sim, cat.AutoAttacks.MHAuto(), landed)
	if !cat.ClearcastingAura.IsActive() {
		t.Error("a paw auto attack consumed Clearcasting")
	}
}

func TestFurorLeavingAtZeroTimeWithZeroEnergy(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear cast failed")
	}
	sim.CurrentTime = 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat cast failed")
	}
	if cat.CurrentEnergy() != 0 {
		t.Fatalf("zero energy stored at t=0 restored %g", cat.CurrentEnergy())
	}
	t.Logf("energy after zero-time bear return=%g", cat.CurrentEnergy())
}

func TestRaidGearFixtureUsesEquippableDruidItems(t *testing.T) {
	data, err := os.ReadFile("../../../ui/specs/druid/feralcat/builds/raid.build.json")
	if err != nil {
		t.Fatal(err)
	}
	settings := &proto.IndividualSimSettings{}
	if err := protojson.Unmarshal(data, settings); err != nil {
		t.Fatal(err)
	}
	gear := settings.GetPlayer().GetEquipment()
	if gear == nil {
		t.Fatal("raid build has no equipment")
	}
	if len(gear.Items) != int(proto.ItemSlot_ItemSlotRanged)+1 {
		t.Fatalf("gear has %d slots, want %d", len(gear.Items), int(proto.ItemSlot_ItemSlotRanged)+1)
	}
	for slot, spec := range gear.Items {
		if spec.Id == 0 {
			continue
		}
		item := core.GetItemByID(spec.Id)
		if item == nil {
			t.Errorf("slot %d item %d is absent from the generated database", slot, spec.Id)
			continue
		}
		if !item.UsableBy(proto.Class_ClassDruid) {
			t.Errorf("slot %d item %d is not usable by Druids", slot, spec.Id)
		}
	}
	worn := make(map[proto.ItemSlot]int32)
	for _, id := range []int32{8345, 272427} {
		found := false
		for slot, item := range gear.Items {
			found = found || item.Id == id
			if item.Id != 0 {
				worn[proto.ItemSlot(slot)] = item.Id
			}
		}
		if !found {
			t.Errorf("supported Cat item %d is absent", id)
		}
	}
	if core.GetItemByID(19287) == nil {
		t.Fatal("Darkmoon Card: Heroism is absent from the generated database")
	}
	for _, item := range gear.Items {
		if item.Id == 19287 {
			t.Error("unsupported Darkmoon Card: Heroism is equipped")
		}
	}
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, worn)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("raid fixture could not cast Shifting Power")
	}
	if got := cat.CurrentEnergy(); got != 60 {
		t.Errorf("raid fixture Shifting Power energy = %v, want 60", got)
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 15*time.Second {
		t.Errorf("raid fixture Shifting Power ready at %v, want 15s", got)
	}
}

func TestFeralStaticSetBonuses(t *testing.T) {
	wildheart := []int32{16716, 16715, 16714, 16720, 16706, 16718}
	feralheart := []int32{22106, 22107, 22108, 22109, 22110, 22111, 22112, 22113}
	for _, bonus := range []struct {
		set    string
		pieces []int32
		stat   stats.Stat
		want   float64
	}{
		{"Wildheart Raiment", wildheart[:2], stats.FireResistance, 8},
		{"Wildheart Raiment", wildheart[:3], stats.AttackPower, 30},
		{"Wildheart Raiment", wildheart[:3], stats.SpellDamage, 18},
		{"Wildheart Raiment", wildheart, stats.MP5, 8},
		{"Feralheart Raiment", feralheart[:2], stats.FireResistance, 8},
		{"Feralheart Raiment", feralheart[:6], stats.AttackPower, 26},
		{"Feralheart Raiment", feralheart[:6], stats.SpellDamage, 15},
		{"Feralheart Raiment", feralheart, stats.Armor, 200},
	} {
		for _, form := range []druid.DruidForm{druid.Cat, druid.DireBear} {
			t.Run(fmt.Sprintf("%s_%d_%s_form_%d", bonus.set, len(bonus.pieces), bonus.stat.StatName(), form), func(t *testing.T) {
				// Given a real character wearing the threshold's source-listed items.
				worn := map[proto.ItemSlot]int32{}
				for _, id := range bonus.pieces {
					item := core.GetItemByID(id)
					if item == nil {
						t.Fatalf("set item %d is absent", id)
					}
					worn[core.ItemTypeToSlot(item.Type)] = id
				}
				sim, cat := newCatWearing(t, "", nil, worn)
				if form == druid.DireBear {
					fillResource(sim, cat)
					if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
						t.Fatal("dire bear form did not cast")
					}
				}
				aura := cat.GetAura(fmt.Sprintf("%s %dP", bonus.set, len(bonus.pieces)))
				if aura == nil || !aura.IsActive() {
					t.Fatalf("%s %d-piece bonus is inactive", bonus.set, len(bonus.pieces))
				}
				before := cat.GetStat(bonus.stat)

				// When the set status aura expires, as it does on an item swap below the threshold.
				aura.Deactivate(sim)

				// Then the source bonus is removed once, independently of the equipped item stats.
				got := before - cat.GetStat(bonus.stat)
				if math.Abs(got-bonus.want) > 1e-9 {
					t.Errorf("removed %s = %v, want %v", bonus.stat.StatName(), got, bonus.want)
				}
				t.Logf("form %d removed %s = %v", form, bonus.stat.StatName(), got)
			})
		}
	}
}

func TestGrovekeeperFerocityFivePieceRegistration(t *testing.T) {
	if druid.ItemSetGrovekeeperFerocity.ID != 2115 {
		t.Errorf("Grovekeeper Ferocity set ID = %d, want 2115", druid.ItemSetGrovekeeperFerocity.ID)
	}
	if druid.ItemSetGrovekeeperFerocity.Bonuses[5] == nil {
		t.Fatal("Grovekeeper Ferocity five-piece bonus is unregistered")
	}
	effect := spelldata.MustFind(1301247).EffectN(1)
	if got := effect.TimeValue(); got != -time.Second {
		t.Errorf("Grovekeeper Ferocity cooldown modifier = %v, want -1s", got)
	}
}

func TestGrovekeeperFerocityTwoPieceRaisesCatAttackSpeed(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotHands: 273889,
		proto.ItemSlot_ItemSlotLegs:  273888,
	})
	aura := cat.GetAura("Grovekeeper Ferocity 2P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Grovekeeper Ferocity two-piece bonus is inactive")
	}
	if got := cat.PseudoStats.AttackSpeedMultiplier; math.Abs(got-1.01) > 1e-9 {
		t.Errorf("Cat melee speed multiplier = %v, want 1.01", got)
	}
	aura.Deactivate(sim)
	if got := cat.PseudoStats.AttackSpeedMultiplier; got != 1 {
		t.Errorf("Cat melee speed multiplier after expiry = %v, want 1", got)
	}
}

func TestGrovekeeperFerocityFourPieceAddsAttackPowerAgainstDemons(t *testing.T) {
	sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{
		proto.ItemSlot_ItemSlotShoulder: 273890,
		proto.ItemSlot_ItemSlotHands:    273889,
		proto.ItemSlot_ItemSlotLegs:     273888,
		proto.ItemSlot_ItemSlotFeet:     273891,
	})
	aura := cat.GetAura("Grovekeeper Ferocity 4P")
	if aura == nil || !aura.IsActive() {
		t.Fatal("Grovekeeper Ferocity four-piece bonus is inactive")
	}
	aura.Deactivate(sim)
	cat.CurrentTarget.MobType = proto.MobType_MobTypeDemon
	aura.Activate(sim)
	table := cat.AttackTables[cat.CurrentTarget.UnitIndex]
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon][stats.AttackPower]; got != 36 {
		t.Errorf("attack power against Demons = %v, want 36", got)
	}
	aura.Deactivate(sim)
	if got := table.MobTypeBonusStats[proto.MobType_MobTypeDemon][stats.AttackPower]; got != 0 {
		t.Errorf("attack power against Demons after expiry = %v, want 0", got)
	}
}

// 16 balance, 20 feral, 16 restoration. Furor is the second restoration talent.
const furorTalents = "0000000000000000-00000000000000000000-0500000000000000"

// Shifting Power is the 11th feral talent and Improved Shifting Power the 15th. Natural Shapeshifter is
// the 5th restoration talent.
const (
	shiftingPowerTalents         = "0000000000000000-00000000001000000000-0000000000000000"
	improvedShiftingPowerTalents = "0000000000000000-00000000001000200000-0000000000000000"
	shapeshifterTalents          = "0000000000000000-00000000001000000000-0000300000000000"
)

func newCat(t *testing.T, talents string, consumes *proto.ConsumesSpec) (*core.Simulation, *FeralDruid) {
	t.Helper()
	return newCatWearing(t, talents, consumes, nil)
}

func newCatWearing(t *testing.T, talents string, consumes *proto.ConsumesSpec, worn map[proto.ItemSlot]int32) (*core.Simulation, *FeralDruid) {
	return newCatWearingTargets(t, talents, consumes, worn, 1)
}

func newCatWearingTargets(t *testing.T, talents string, consumes *proto.ConsumesSpec, worn map[proto.ItemSlot]int32, targetCount int) (*core.Simulation, *FeralDruid) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	for slot, id := range worn {
		items[slot].Id = id
	}
	if consumes == nil {
		consumes = &proto.ConsumesSpec{}
	}
	targets := make([]*proto.Target, targetCount)
	for i := range targets {
		targets[i] = &proto.Target{Name: "target", Level: 60}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString: talents,
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{},
				}},
				Equipment:   &proto.EquipmentSpec{Items: items},
				Consumables: consumes,
				Buffs:       &proto.IndividualBuffs{},
				Rotation:    &proto.APLRotation{},
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: targets, Duration: 60},
	}, simsignals.CreateSignals())
	sim.Reset()
	return sim, sim.Raid.Parties[0].Players[0].(*FeralDruid)
}

func fillResource(sim *core.Simulation, cat *FeralDruid) {
	cat.AddEnergy(sim, cat.MaximumEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.AddRage(sim, cat.MaximumRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
}

// Ferocity 5/5 and Shredding Attacks 3/3: tooltips 16934 ("5 Rage or Energy") and 16966 ("Shred by 18",
// "Lacerate by 3").
func TestFeralCostTalentsLowerEnergyAndRageInTheirOwnUnits(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000000000000000-50000030000000000000-0000000000000000", nil)
	for _, c := range []struct {
		name       string
		base, with *druid.DruidSpell
		less       float64
	}{
		{"Shred", base.Shred, talented.Shred, 18},
		{"Rake", base.Rake, talented.Rake, 5},
		{"Claw", base.Claw, talented.Claw, 5},
		{"Lacerate", base.Lacerate, talented.Lacerate, 3},
		{"Swipe", base.Swipe, talented.Swipe, 5},
		{"Primal Bite", base.PrimalBite, talented.PrimalBite, 5},
	} {
		if got := c.base.Cost.GetCurrentCost() - c.with.Cost.GetCurrentCost(); got != c.less {
			t.Errorf("%s costs %v less with the talents, want %v", c.name, got, c.less)
		}
	}
}

func TestFurorRestoresEnergyAfterLeavingCat(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	if !cat.InForm(druid.Cat) {
		t.Fatal("the cat did not start in cat form")
	}
	fillResource(sim, cat)
	cat.SpendEnergy(sim, 60, cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	if cat.InForm(druid.Cat) {
		t.Fatal("dire bear form left the cat in cat form")
	}
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	// 40 stored at 5/5. The 3 seconds were spent in Dire Bear Form, which earns nothing.
	if got := cat.CurrentEnergy(); got != 40 {
		t.Errorf("energy back in cat form = %v, want 40", got)
	}
}

func TestWithoutFurorCatFormComesBackEmpty(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	sim.CurrentTime += 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat form did not cast on the way back")
	}
	if got := cat.CurrentEnergy(); got != 0 {
		t.Errorf("energy back in cat form without Furor = %v, want 0", got)
	}
}

func TestProwlHoldsAutoAttacksUntilTheOpener(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	sim.CurrentTime = -time.Second
	if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
		t.Fatal("prowl did not cast before the pull")
	}
	mh, _, _ := cat.AutoAttacks.SwingAt()
	if mh != core.NeverExpires-time.Second {
		t.Errorf("main-hand swing while prowling at %v, want held", mh)
	}
	sim.CurrentTime = 0
	fillResource(sim, cat)
	if cat.Prowl.CanCast(sim, cat.CurrentTarget) {
		t.Error("prowl could be cast in combat")
	}
	if !cat.Pounce.Cast(sim, cat.CurrentTarget) {
		t.Fatal("pounce did not cast from prowl")
	}
	if cat.ProwlAura.IsActive() {
		t.Error("pounce left the druid prowling")
	}
	if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != 0 {
		t.Errorf("main-hand swing after the opener at %v, want 0", mh)
	}
}

func TestCatQueuesMaulOntoTheBearSwing(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	fillResource(sim, cat)
	if !cat.Maul.Cast(sim, cat.CurrentTarget) {
		t.Fatal("maul did not queue")
	}
	for sim.CurrentTime < 100*time.Millisecond {
		if sim.Step() {
			break
		}
	}
	if got := cat.AutoAttacks.MaybeReplaceMHSwing(sim, nil); got == nil || got.ActionID.SpellID != 9881 {
		t.Fatalf("the queued swing was %v, want maul 9881", got)
	}
}

func TestCowerAndFaerieFireEndProwl(t *testing.T) {
	for _, spell := range []struct {
		name string
		cast func(*core.Simulation, *FeralDruid) bool
	}{
		{"cower", func(sim *core.Simulation, cat *FeralDruid) bool {
			return cat.Cower.Cast(sim, cat.CurrentTarget)
		}},
		{"faerie fire", func(sim *core.Simulation, cat *FeralDruid) bool {
			return cat.FaerieFire.Cast(sim, cat.CurrentTarget)
		}},
	} {
		t.Run(spell.name, func(t *testing.T) {
			sim, cat := newCat(t, furorTalents, nil)
			sim.CurrentTime = -time.Second
			if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
				t.Fatal("prowl did not cast before the pull")
			}
			sim.CurrentTime = 3 * time.Second
			fillResource(sim, cat)
			if !spell.cast(sim, cat) {
				t.Fatalf("%s did not cast from prowl", spell.name)
			}
			if cat.ProwlAura.IsActive() {
				t.Errorf("%s left the druid prowling", spell.name)
			}
			if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != 3*time.Second {
				t.Errorf("main-hand swing after %s at %v, want 3s", spell.name, mh)
			}
		})
	}
}

func TestFaerieFireRefreshesWhileUp(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	target := cat.CurrentTarget
	aura := cat.FaerieFireAuras.Get(target)
	aura.Activate(sim)
	sim.CurrentTime += 35 * time.Second
	if !cat.FaerieFire.CanCast(sim, target) {
		t.Fatal("faerie fire could not be recast with 5 s of its debuff left")
	}
}

func emptyEnergy(sim *core.Simulation, cat *FeralDruid) {
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
}

func TestShiftingPowerTradesManaForEnergy(t *testing.T) {
	sim, cat := newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	mana := cat.CurrentMana()
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast in cat form")
	}
	if got := cat.CurrentEnergy(); got != 40 {
		t.Errorf("energy after shifting power = %v, want 40", got)
	}
	// 55% of base mana, the shapeshift's own price.
	if got, want := mana-cat.CurrentMana(), float64(int32(55*cat.BaseMana)/100); got != want {
		t.Errorf("mana spent = %v, want %v", got, want)
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 16*time.Second {
		t.Errorf("shifting power ready at %v, want 16s", got)
	}
}

func TestShiftingPowerNeedsTheTalentAndCatForm(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	fillResource(sim, cat)
	if cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) {
		t.Error("shifting power could be cast without the talent")
	}
	sim, cat = newCat(t, shiftingPowerTalents, nil)
	fillResource(sim, cat)
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("dire bear form did not cast")
	}
	fillResource(sim, cat)
	if cat.ShiftingPower.CanCast(sim, cat.CurrentTarget) {
		t.Error("shifting power could be cast in dire bear form")
	}
}

func TestImprovedShiftingPowerTakesTheCooldownToEightSeconds(t *testing.T) {
	sim, cat := newCat(t, improvedShiftingPowerTalents, nil)
	fillResource(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 8*time.Second {
		t.Errorf("shifting power ready at %v, want 8s", got)
	}
}

func TestNaturalShapeshifterCheapensShiftingPower(t *testing.T) {
	_, cat := newCat(t, shiftingPowerTalents, nil)
	full := cat.ShiftingPower.Cost.GetCurrentCost()
	_, cat = newCat(t, shapeshifterTalents, nil)
	// 3/3 is 30% off, the same as the form spells.
	if got, want := cat.ShiftingPower.Cost.GetCurrentCost(), full*0.7; math.Abs(got-want) > 1 {
		t.Errorf("shifting power cost with natural shapeshifter = %v, want %v", got, want)
	}
}

func TestWolfsheadAddsShiftingPowerEnergyAndEnrageRage(t *testing.T) {
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotHead: 8345})
	fillResource(sim, cat)
	emptyEnergy(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	if got := cat.CurrentEnergy(); got != 60 {
		t.Errorf("energy after shifting power with wolfshead = %v, want 60", got)
	}
	sim.CurrentTime += 2 * time.Second
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear form did not cast")
	}
	cat.SpendRage(sim, cat.CurrentRage(), cat.NewRageMetrics(core.ActionID{SpellID: 1}))
	if !cat.Enrage.Cast(sim, cat.CurrentTarget) {
		t.Fatal("enrage did not cast")
	}
	// The row's 10 instant rage, raised by half.
	if got := cat.CurrentRage(); got != 15 {
		t.Errorf("rage after enrage with wolfshead = %v, want 15", got)
	}
}

func TestHowlingIdolShortensShiftingPower(t *testing.T) {
	sim, cat := newCatWearing(t, shiftingPowerTalents, nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotRanged: 272427})
	fillResource(sim, cat)
	if !cat.ShiftingPower.Cast(sim, cat.CurrentTarget) {
		t.Fatal("shifting power did not cast")
	}
	if got := cat.ShiftingPower.ReadyAt(); got != 15*time.Second {
		t.Errorf("shifting power ready at %v with the idol, want 15s", got)
	}
}

func TestFurorGrantsRageOnEnteringBear(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.AddMana(sim, cat.MaxMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear form did not cast")
	}
	if !cat.InForm(druid.Bear) {
		t.Fatal("bear form did not set the bear form")
	}
	// 5/5 is a guaranteed 10 rage. The bar starts empty.
	if got := cat.CurrentRage(); got != 10 {
		t.Errorf("rage on entering bear = %v, want 10", got)
	}
}

func TestClawAwardsComboPointsAndRipSpendsThem(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	var landed bool
	for range 15 {
		if cat.Claw.Cast(sim, cat.CurrentTarget) && cat.ComboPoints() > 0 {
			landed = true
			break
		}
		fillResource(sim, cat)
		sim.CurrentTime += 2 * time.Second
	}
	if !landed {
		t.Fatal("claw never awarded a combo point")
	}
	points := cat.ComboPoints()
	fillResource(sim, cat)
	sim.CurrentTime += 2 * time.Second
	if !cat.Rip.Cast(sim, cat.CurrentTarget) {
		t.Fatal("rip did not cast")
	}
	if got := cat.ComboPoints(); got != 0 {
		t.Errorf("combo points after rip = %d, want 0 (spent %d)", got, points)
	}
	if !cat.Rip.Dot(cat.CurrentTarget).IsActive() {
		t.Error("rip left no bleed")
	}
}

func TestShredRequiresATargetThatCanBeShredded(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	fillResource(sim, cat)
	cat.PseudoStats.InFrontOfTarget = true
	if cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could be cast while in front of the target")
	}
	cat.PseudoStats.InFrontOfTarget = false
	if !cat.Shred.CanCast(sim, cat.CurrentTarget) {
		t.Error("shred could not be cast from behind")
	}
}

func TestConsumablesFollowTheNotShapeshiftedRow(t *testing.T) {
	interesting := map[proto.ConsumableType]bool{
		proto.ConsumableType_ConsumableTypePotion:    true,
		proto.ConsumableType_ConsumableTypeExplosive: true,
		proto.ConsumableType_ConsumableTypeDevice:    true,
		proto.ConsumableType_ConsumableTypeConjured:  true,
	}
	var breakingID, safeID int32
	var breakingSpell, safeSpell int32
	for _, consumable := range assetsdb.Load().Consumables {
		id := consumable.Id
		item := core.GetConsumableByID(id)
		if !interesting[item.Type] || len(item.OnUseEffects()) == 0 {
			continue
		}
		row := spelldata.Find(item.OnUse().GetBuffId())
		if row == spelldata.Nil {
			continue
		}
		breaks := row.NotShapeshifted()
		if breaks {
			t.Logf("breaks form: %d %s spell %d", id, item.Name, row.ID)
		}
		// The cast below uses a breaker and a potion that need no profession. The log above is every row.
		if item.RequiredProfession != 0 {
			continue
		}
		if breaks && breakingID == 0 {
			breakingID, breakingSpell = id, row.ID
		}
		if !breaks && safeID == 0 && item.Type == proto.ConsumableType_ConsumableTypePotion && id == 13444 {
			safeID, safeSpell = id, row.ID
		}
	}
	if breakingID == 0 || safeID == 0 {
		t.Fatalf("missing a consumable pair, breaking %d safe %d", breakingID, safeID)
	}

	sim, cat := newCat(t, furorTalents, &proto.ConsumesSpec{PotId: safeID, ExtraIds: []int32{breakingID}})
	breaking := cat.GetSpell(core.ActionID{ItemID: breakingID})
	safe := cat.GetSpell(core.ActionID{ItemID: safeID})
	if breaking == nil || safe == nil {
		t.Fatalf("consumable spells did not register, item %d spell %d and item %d spell %d", breakingID, breakingSpell, safeID, safeSpell)
	}
	if !breaking.CastRequirement.NotShapeshifted || safe.CastRequirement.NotShapeshifted {
		t.Fatalf("cast requirement breaking=%v safe=%v", breaking.CastRequirement.NotShapeshifted, safe.CastRequirement.NotShapeshifted)
	}

	sawBreaker := false
	for _, mcd := range cat.GetMajorCooldowns() {
		if mcd.Spell != breaking {
			continue
		}
		sawBreaker = true
		if mcd.TryActivate(sim, cat.GetCharacter()) || !cat.InForm(druid.Cat) {
			t.Error("a cat automatically used a consumable that leaves cat form")
		}
	}
	if !sawBreaker {
		t.Fatal("the form-breaking consumable was not a major cooldown")
	}

	castConsumable := func(form druid.DruidForm, spell *core.Spell, wantShift bool) {
		t.Helper()
		fillResource(sim, cat)
		waitReady(sim, cat)
		if !spell.Cast(sim, cat.CurrentTarget) {
			t.Fatalf("%s did not cast in %v", spell.ActionID, form)
		}
		waitReady(sim, cat)
		shifted := cat.InForm(druid.Cat) || cat.InForm(druid.Bear)
		if wantShift && shifted {
			t.Errorf("%s left the druid in a form", spell.ActionID)
		}
		if !wantShift && !cat.InForm(form) {
			t.Errorf("%s left %v", spell.ActionID, form)
		}
	}

	castConsumable(druid.Cat, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to cat form")
	}
	resetCooldown(safe)
	castConsumable(druid.Cat, safe, false)

	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not enter bear form")
	}
	resetCooldown(breaking)
	castConsumable(druid.Bear, breaking, true)
	fillResource(sim, cat)
	waitReady(sim, cat)
	if !cat.BearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("could not return to bear form")
	}
	resetCooldown(safe)
	castConsumable(druid.Bear, safe, false)
}

func waitReady(sim *core.Simulation, cat *FeralDruid) {
	deadline := sim.CurrentTime + 10*time.Second
	for sim.CurrentTime < deadline {
		if cat.GCD.IsReady(sim) && cat.Hardcast.Expires <= sim.CurrentTime {
			return
		}
		if sim.Step() {
			return
		}
	}
}

func resetCooldown(spell *core.Spell) {
	if spell.CD.Timer != nil {
		spell.CD.Reset()
	}
	if spell.SharedCD.Timer != nil {
		spell.SharedCD.Reset()
	}
}

func TestRotationWithoutShiftingPowerDoesNotReportManaExhaustion(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	cat.SpendMana(sim, cat.CurrentMana(), cat.NewManaMetrics(core.ActionID{SpellID: 1}))
	cat.doRotation(sim)
	if cat.IsOOM() {
		t.Fatal("rotation without Shifting Power reported mana exhaustion")
	}
}

func TestFeralPreset(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rotation := core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", "default").Rotation
			request := coverageRequest(false, rotation)
			player := request.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""

			result := core.RunRaidSim(request)

			if result.Error != nil {
				t.Fatal(result.Error.Message)
			}
			metrics := result.RaidMetrics.Parties[0].Players[0]
			if dps := metrics.Dps.Avg; dps <= 0 || math.IsNaN(dps) || math.IsInf(dps, 0) {
				t.Fatalf("minimal cat produces invalid damage rate %v", dps)
			}
			casts := make(map[int32]int32)
			for _, action := range metrics.Actions {
				for _, target := range action.Targets {
					casts[action.Id.GetSpellId()] += target.Casts
				}
			}
			for _, id := range []int32{9850, 9896} {
				if casts[id] == 0 {
					t.Errorf("minimal cat never casts spell %d", id)
				}
			}
			t.Logf("Claw casts=%d, Rip casts=%d, DPS=%g", casts[9850], casts[9896], metrics.Dps.Avg)
		})
	}
}

func TestReflectionRestoresCastingManaRegen(t *testing.T) {
	for rank, fraction := range []float64{0, 0.17, 0.33, 0.50} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given a cat with no mana buffs or equipment mana regeneration.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-00000%d0000000000", rank)
			_, cat := newCat(t, talents, nil)

			// When the mana bar computes regeneration during casting.
			got := cat.ManaRegenPerSecondWhileCasting()

			// Then the game files' rank curve preserves this share of spirit regeneration.
			want := cat.SpiritManaRegenPerSecond() * fraction
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("casting mana regeneration = %v, want %v at fraction %v", got, want, fraction)
			}
			t.Logf("rank %d casting mana regeneration = %v, spirit fraction = %v", rank, got, cat.PseudoStats.SpiritRegenRateCasting)
		})
	}
}

func TestSubtletyReducesNatureThreatWithoutReducingPhysicalThreat(t *testing.T) {
	for rank, multiplier := range []float64{1, 0.9, 0.8, 0.7} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given a cat with registered Nature and physical abilities.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-000%d000000000000", rank)
			sim, cat := newCat(t, talents, nil)
			arcane := cat.RegisterSpell(core.SpellConfig{
				ActionID:         core.ActionID{SpellID: 1},
				SpellSchool:      core.SpellSchoolArcane,
				ProcMask:         core.ProcMaskSpellDamage,
				DamageMultiplier: 1,
				ThreatMultiplier: 1,
				FlatThreatBonus:  20,
			})
			for _, ability := range []struct {
				name  string
				spell *core.Spell
				want  float64
			}{
				{"nature", cat.FaerieFire.Spell, 71 * multiplier},
				{"arcane with flat threat", arcane.Spell, 85.2 * multiplier},
				{"physical", cat.Claw.Spell, 71},
			} {
				t.Run(ability.name, func(t *testing.T) {
					// When the registered spell converts a landed damage amount to threat.
					got := ability.spell.ThreatFromDamage(sim, core.OutcomeHit, 100, cat.AttackTables[cat.CurrentTarget.UnitIndex])

					// Then Subtlety affects only the row's school mask.
					if math.Abs(got-ability.want) > 1e-9 {
						t.Errorf("threat = %v, want %v", got, ability.want)
					}
					t.Logf("rank %d %s threat = %v", rank, ability.name, got)
				})
			}
		})
	}
}

func TestLivingSpiritIncreasesTotalSpirit(t *testing.T) {
	_, base := newCat(t, "", nil)
	for rank, multiplier := range []float64{1, 1.05, 1.10, 1.15} {
		t.Run(fmt.Sprintf("rank_%d", rank), func(t *testing.T) {
			// Given the same cat with a rank of Living Spirit.
			talents := fmt.Sprintf("0000000000000000-00000000000000000000-000000000000%d000", rank)
			_, cat := newCat(t, talents, nil)

			// When the character's stat dependencies resolve.
			got := cat.GetStat(stats.Spirit)

			// Then total Spirit includes the rank's percentage.
			want := math.Floor(base.GetStat(stats.Spirit) * multiplier)
			if math.Abs(got-want) > 1e-9 {
				t.Errorf("Spirit = %v, want %v at multiplier %v", got, want, multiplier)
			}
			t.Logf("rank %d Spirit = %v (untalented %v)", rank, got, base.GetStat(stats.Spirit))
		})
	}
}

func TestNaturalistIncreasesCatDamage(t *testing.T) {
	_, base := newCat(t, "", nil)
	_, talented := newCat(t, "0000000000000000-00000000000000000000-0050000000000000", nil)
	if got, want := talented.PseudoStats.DamageDealtMultiplier/base.PseudoStats.DamageDealtMultiplier, 1.05; math.Abs(got-want) > 1e-9 {
		t.Errorf("damage dealt multiplier ratio = %v, want %v", got, want)
	}
}

func TestBiteDodgeRefundsBaseCostAndKeepsComboPoints(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	target.PseudoStats.BaseDodgeChance = 1
	cat.AddComboPoints(sim, 5, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
	if !cat.FerociousBite.Cast(sim, target) {
		t.Fatal("bite did not cast")
	}
	if cat.FerociousBite.SpellMetrics[target.UnitIndex].Dodges != 1 {
		t.Fatal("bite was not dodged")
	}
	if cat.CurrentEnergy() != 93 || cat.ComboPoints() != 5 {
		t.Fatalf("dodged bite leaves energy=%g CP=%d, want 93 and 5", cat.CurrentEnergy(), cat.ComboPoints())
	}
	t.Logf("dodged bite: energy=%g CP=%d", cat.CurrentEnergy(), cat.ComboPoints())
}

func TestRakeReadsAttackPowerAtEachTick(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
	dot := cat.Rake.Dot(cat.CurrentTarget)
	dot.Apply(sim)
	metric := &cat.Rake.SpellMetrics[cat.CurrentTarget.UnitIndex]
	before := metric.TotalDamage
	dot.TickOnce(sim)
	low := metric.TotalDamage - before
	cat.AddStatDynamic(sim, stats.AttackPower, 100)
	before = metric.TotalDamage
	dot.TickOnce(sim)
	high := metric.TotalDamage - before
	if got, want := high-low, 5.5; math.Abs(got-want) > 1e-9 {
		t.Fatalf("Rake tick gained %g damage from 100 attack power added after application, want %g", got, want)
	}
}

func TestRakeInitialDamageUsesNativeArmor(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	target := cat.CurrentTarget
	hit := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	target.AddStatDynamic(sim, stats.Armor, 10000)
	armored := cat.Rake.CalcDamage(sim, target, 100, cat.Rake.OutcomeAlwaysHit).Damage
	if hit <= 0 || armored <= 0 || armored >= hit {
		t.Fatalf("Rake initial hit %g -> %g after armor", hit, armored)
	}
	t.Logf("Rake initial hit before/after 10000 armor: %g/%g", hit, armored)
}

func TestReviewedFurorLeavesAtZeroTimeWithZeroEnergy(t *testing.T) {
	sim, cat := newCat(t, furorTalents, nil)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("bear cast failed")
	}
	sim.CurrentTime = 3 * time.Second
	fillResource(sim, cat)
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("cat cast failed")
	}
	if cat.CurrentEnergy() != 0 {
		t.Fatalf("zero energy stored at t=0 restored %g", cat.CurrentEnergy())
	}
}

func TestRipAttackPowerCapsAtFourComboPoints(t *testing.T) {
	for cp := int32(1); cp <= 5; cp++ {
		sim, cat := newCat(t, "", nil)
		cat.AddStatDynamic(sim, stats.PhysicalCritPercent, -100)
		cat.AddComboPoints(sim, cp, cat.NewComboPointMetrics(core.ActionID{SpellID: 1}))
		dot := cat.Rip.Dot(cat.CurrentTarget)
		dot.Apply(sim)
		metric := &cat.Rip.SpellMetrics[cat.CurrentTarget.UnitIndex]
		before := metric.TotalDamage
		dot.TickOnce(sim)
		low := metric.TotalDamage - before
		cat.AddStatDynamic(sim, stats.AttackPower, 100)
		before = metric.TotalDamage
		dot.TickOnce(sim)
		gain := metric.TotalDamage - before - low
		if math.Abs(gain-float64(min(cp, 4))) > 1e-5 {
			t.Fatalf("cp=%d tick gain=%g", cp, gain)
		}
		t.Logf("Rip %d CP: 100 AP gains %g per tick / %g total", cp, gain, gain*6)
	}
}

func TestCatShippedPresetRuns(t *testing.T) {
	req := buildRequest(t, "raid")
	player := runCat(t, req)
	for _, id := range []int32{9830, 9896, 9904, 1322605, 417141} {
		if castsOf(player.Actions, id) == 0 {
			t.Fatalf("shipped preset never casts %d", id)
		}
	}
	for _, id := range []int32{22812} {
		if n := castsOf(player.Actions, id); n != 0 {
			t.Fatalf("shipped preset casts %d %d times", id, n)
		}
	}
	if n := castsOf(player.Actions, 9907); n == 0 {
		t.Fatal("shipped preset never fishes with Faerie Fire")
	}
}

func TestCatStealthPresetsRun(t *testing.T) {
	for name, opener := range map[string]int32{"ravage": 9867, "pounce": 9827} {
		t.Run(name, func(t *testing.T) {
			req := catPowershiftRequest(powershiftTalents, core.GetAplRotation("../../../ui/specs/druid/feralcat/apls", name).Rotation, false)
			req.Raid.Parties[0].Players[0].Equipment = core.GetGearSet("../../../ui/specs/druid/feralcat/gear_sets", "default").GearSet
			player := runCat(t, req)
			if castsOf(player.Actions, opener) != 1 || castsOf(player.Actions, 9830) == 0 {
				t.Fatalf("%s must open once then continue with Shred", name)
			}
		})
	}
}

func reviewProwl(t *testing.T, sim *core.Simulation, cat *FeralDruid) {
	t.Helper()
	sim.CurrentTime = -time.Second
	if !cat.Prowl.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Prowl failed before combat")
	}
	sim.CurrentTime = time.Second
	fillResource(sim, cat)
}

func TestReviewProwlAndFeralSwiftnessSpeed(t *testing.T) {
	_, base := newCat(t, "", nil)
	baseSpeed := base.GetMovementSpeed()
	sim, cat := newCat(t, "0000000000000000-00200000000000000000-0000000000000000", nil)
	check := func(want float64) {
		t.Helper()
		if got := cat.GetMovementSpeed(); math.Abs(got-want) > 1e-9 {
			t.Fatalf("movement speed %g, want %g", got, want)
		}
	}
	check(baseSpeed * 1.3)
	reviewProwl(t, sim, cat)
	check(baseSpeed * 1.3 * 0.7)
	if !cat.Dash.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("Dash failed or broke Prowl")
	}
	sim.CurrentTime += 2 * time.Second
	if !cat.Claw.Cast(sim, cat.CurrentTarget) || cat.ProwlAura.IsActive() {
		t.Fatal("Claw failed or retained Prowl")
	}
	check(baseSpeed * 1.3)
	cat.ClearForm(sim)
	check(baseSpeed)
	sim.CurrentTime += 2 * time.Second
	if !cat.CatForm.Cast(sim, cat.CurrentTarget) {
		t.Fatal("return to Cat Form failed")
	}
	check(baseSpeed * 1.3)
	t.Logf("cat=%g yd/s, prowling cat=%g yd/s, humanoid=%g yd/s; re-entry restored cat speed", baseSpeed*1.3, baseSpeed*1.3*0.7, baseSpeed)
}

func TestReviewCatChargeBreaksProwlWithoutDebuff(t *testing.T) {
	sim, cat := newCat(t, "0000000000000000-00000000100000000000-0000000000000000", nil)
	reviewProwl(t, sim, cat)
	cat.DistanceFromTarget = 10
	if !cat.FeralCharge.Cast(sim, cat.CurrentTarget) {
		t.Fatal("Cat Charge failed")
	}
	if cat.ProwlAura.IsActive() {
		t.Fatal("Cat Charge retained Prowl")
	}
	for _, array := range cat.FeralCharge.RelatedAuraArrays {
		if array.Get(cat.CurrentTarget).IsActive() {
			t.Fatal("Cat Charge applied a bear debuff")
		}
	}
	if mh, _, _ := cat.AutoAttacks.SwingAt(); mh != sim.CurrentTime {
		t.Fatalf("melee resumes at %v, want %v", mh, sim.CurrentTime)
	}
	t.Log("Cat Charge ended Prowl, resumed melee, and applied no root or daze")
}

func TestReviewProwlIgnoresFailedCastsAndTriggeredSpells(t *testing.T) {
	sim, cat := newCat(t, "", nil)
	reviewProwl(t, sim, cat)
	cat.SpendEnergy(sim, cat.CurrentEnergy(), cat.NewEnergyMetrics(core.ActionID{SpellID: 1}))
	if cat.Claw.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("unaffordable Claw cast or broke Prowl")
	}
	if !cat.PounceBleed.Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("triggered Pounce bleed failed or broke Prowl")
	}
	if !cat.AutoAttacks.MHAuto().Cast(sim, cat.CurrentTarget) || !cat.ProwlAura.IsActive() {
		t.Fatal("automatic swing failed or was treated as a Prowl-breaking cast")
	}
	t.Log("failed cast and triggered bleed retained Prowl")
}

func TestReviewConsumableBreaksProwl(t *testing.T) {
	sim, cat := newCat(t, "", &proto.ConsumesSpec{PotId: 13444})
	reviewProwl(t, sim, cat)
	potion := cat.GetSpell(core.ActionID{ItemID: 13444})
	if potion == nil || !potion.Cast(sim, &cat.Unit) {
		t.Fatal("Major Mana Potion failed")
	}
	if cat.ProwlAura.IsActive() {
		t.Fatal("Major Mana Potion retained Prowl")
	}
	t.Log("item-ID consumable ended Prowl")
}

// Same tree as the cat UI default. Shifting Power is the 11th feral talent.
const powershiftTalents = "05002-55000032021032212051-055"

func rakeRipRotation() *proto.APLRotation {
	return catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	})
}

func TestRakeRipRotationStaysInCatForm(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, rakeRipRotation(), false))
	for _, c := range []struct {
		name string
		id   int32
	}{{"Shifting Power", 1322605}, {"Shred", 9830}, {"Rip", 9896}, {"Rake", 9904}, {"Berserk", 417141}} {
		if castsOf(player.Actions, c.id) == 0 {
			t.Errorf("%s was never cast", c.name)
		}
	}
	for _, c := range []struct {
		name string
		id   int32
	}{{"Cat Form", 768}, {"Ferocious Bite", 22829}, {"Faerie Fire", 9907}} {
		if got := castsOf(player.Actions, c.id); got != 0 {
			t.Errorf("%s casts = %d, the default rotation leaves it out", c.name, got)
		}
	}
}

func TestMaintainRakeKeepsRakeUp(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints:  4,
		BiteMinComboPoints: 5,
		RakeTrick:          true,
	})
	player := runCat(t, catPowershiftRequest(powershiftTalents, rotation, false))
	// Rake lasts 9 s, so a 90 s fight that keeps it up casts it about ten times.
	if got := castsOf(player.Actions, 9904); got < 8 {
		t.Errorf("rake casts = %d in 90 s, want it kept up", got)
	}
}

func TestRotationWithoutShiftingPowerTalent(t *testing.T) {
	player := runCat(t, catPowershiftRequest(withFeralDigit(powershiftTalents, 10, '0'), rakeRipRotation(), false))
	if got := castsOf(player.Actions, 1322605); got != 0 {
		t.Fatalf("shifting power casts = %d, the talent is not taken", got)
	}
	if castsOf(player.Actions, 9896) == 0 {
		t.Fatal("rip was never cast")
	}
}

func TestRotationClawsWhenShredCannotLand(t *testing.T) {
	player := runCat(t, catPowershiftRequest(powershiftTalents, rakeRipRotation(), true))
	if got := castsOf(player.Actions, 9830); got != 0 {
		t.Errorf("shred casts = %d against a target that cannot be shredded", got)
	}
	if castsOf(player.Actions, 9850) == 0 || castsOf(player.Actions, 9896) == 0 {
		t.Error("claw did not build combo points into rip")
	}
}

func TestRipweaveCastsRipBesideTheBiteFinisher(t *testing.T) {
	rotation := catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:      proto.FeralCatDruid_Rotation_Bite,
		Ripweave:           true,
		RipMinComboPoints:  5,
		BiteMinComboPoints: 5,
	})
	player := runCat(t, catPowershiftRequest(powershiftTalents, rotation, false))
	if castsOf(player.Actions, 9896) == 0 || castsOf(player.Actions, 22829) == 0 {
		t.Errorf("rip %d and bite %d casts, want both", castsOf(player.Actions, 9896), castsOf(player.Actions, 22829))
	}
}

func runCat(t *testing.T, request *proto.RaidSimRequest) *proto.UnitMetrics {
	t.Helper()
	result := core.RunSim(request, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatalf("sim error: %s", result.Error.Message)
	}
	player := result.RaidMetrics.Parties[0].Players[0]
	if player.Dps.Avg <= 0 {
		t.Fatalf("dps = %v, want a fight that deals damage", player.Dps.Avg)
	}
	return player
}

func catPowershiftRequest(talents string, rotation *proto.APLRotation, cannotShred bool) *proto.RaidSimRequest {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotMainHand+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand].Id = 866
	return &proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1, Iterations: 1, IsTest: true},
		Raid: &proto.Raid{Parties: []*proto.Party{{
			Players: []*proto.Player{{
				Name: "Cat", Class: proto.Class_ClassDruid, Race: proto.Race_RaceNightElf,
				TalentsString:  talents,
				ReactionTimeMs: 100,
				Equipment:      &proto.EquipmentSpec{Items: items},
				Consumables:    &proto.ConsumesSpec{},
				Buffs:          &proto.IndividualBuffs{},
				Spec: &proto.Player_FeralCatDruid{FeralCatDruid: &proto.FeralCatDruid{
					Options: &proto.FeralCatDruid_Options{CannotShredTarget: cannotShred},
				}},
				Rotation: rotation,
			}},
			Buffs: &proto.PartyBuffs{},
		}}},
		Encounter: &proto.Encounter{Targets: []*proto.Target{{Name: "target", Level: 60}}, Duration: 90},
	}
}

func TestManaPotionAndThistleTeaAreUsed(t *testing.T) {
	const iterations int32 = 100
	req := sweepRequest(t, powershiftTalents, catRotation(&proto.FeralCatDruid_Rotation{
		FinishingMove:     proto.FeralCatDruid_Rotation_Rip,
		RipMinComboPoints: 5,
		RakeTrick:         true,
	}), false, iterations)
	req.Raid.Parties[0].Players[0].Consumables = &proto.ConsumesSpec{PotId: 13444, ConjuredId: 7676}
	result := core.RunSim(req, nil, simsignals.CreateSignals())
	if result.Error != nil {
		t.Fatal(result.Error.Message)
	}
	actions := result.RaidMetrics.Parties[0].Players[0].Actions
	per := func(id int32) float64 { return float64(usesOfID(actions, id)) / float64(result.IterationsDone) }
	if tea, pot := per(7676), per(13444); tea < 0.8 || pot < 0.8 {
		t.Fatalf("thistle tea %.2f, mana potion %.2f per fight, want both near 1", tea, pot)
	}
}

func TestAvailableThistleTeaRequiresSelection(t *testing.T) {
	req := buildRequest(t, "raid")
	req.Raid.Parties[0].Players[0].Consumables.ConjuredItems = []int32{7676}
	player := runCat(t, req)
	if n := usesOfID(player.Actions, 7676); n != 0 {
		t.Fatalf("unselected Thistle Tea casts = %d, want 0", n)
	}
}

func usesOfID(actions []*proto.ActionMetrics, id int32) int32 {
	var n int32
	for _, action := range actions {
		got := action.GetId()
		if got.GetSpellId() != id && got.GetItemId() != id {
			continue
		}
		for _, target := range action.Targets {
			n += target.Casts
		}
	}
	return n
}

func castsOf(actions []*proto.ActionMetrics, spellID int32) int32 {
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

type presetTalentLocation struct {
	Row int `json:"rowIdx"`
	Col int `json:"colIdx"`
}

type presetTalent struct {
	FieldName string                `json:"fieldName"`
	Location  presetTalentLocation  `json:"location"`
	Prereq    *presetTalentLocation `json:"prereqLocation"`
	MaxPoints int                   `json:"maxPoints"`
}

type presetTalentTree struct {
	Talents []presetTalent `json:"talents"`
}

func TestFeralStarterInputs(t *testing.T) {
	data, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []presetTalentTree
	if err := json.Unmarshal(data, &trees); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct{ name, path string }{
		{"feralcat", "../../../ui/specs/druid/feralcat/builds/starter.build.json"},
		{"feralbear", "../../../ui/specs/druid/feralbear/builds/starter.build.json"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			data, err := os.ReadFile(fixture.path)
			if err != nil {
				t.Fatal(err)
			}
			settings := &proto.IndividualSimSettings{}
			if err := protojson.Unmarshal(data, settings); err != nil {
				t.Fatal(err)
			}
			player := settings.Player
			if player.Class != proto.Class_ClassDruid || player.Race != proto.Race_RaceNightElf {
				t.Fatal("starter requires a night elf druid")
			}
			filter := core.ItemFilter{
				Class: proto.Class_ClassDruid, ArmorType: proto.ArmorType_ArmorTypeLeather,
				WeaponTypes: []proto.WeaponType{proto.WeaponType_WeaponTypeStaff},
			}
			equipment := core.ProtoToEquipment(player.Equipment)
			for slot, itemSpec := range player.Equipment.Items {
				if itemSpec.Id == 0 {
					continue
				}
				item := core.GetItemByID(itemSpec.Id)
				if item == nil || !filter.Matches(*item, true) {
					t.Errorf("illegal starter item %d in slot %d", itemSpec.Id, slot)
					continue
				}
				if equipment[slot].ID != itemSpec.Id {
					t.Errorf("item %d does not equip in slot %d", itemSpec.Id, slot)
				}
				if item.SetName != "" {
					t.Errorf("starter item %d selects an unmodelled set bonus", itemSpec.Id)
				}
				if itemSpec.Enchant != 0 || itemSpec.RandomSuffix != 0 {
					t.Errorf("starter item %d assumes extra customization", itemSpec.Id)
				}
			}
			validatePresetTalents(t, trees, player.TalentsString)
		})
	}
}

func validatePresetTalents(t testing.TB, trees []presetTalentTree, talents string) {
	t.Helper()
	parts := strings.Split(talents, "-")
	if len(parts) != len(trees) {
		t.Fatal("talents must identify all three talent trees")
	}
	total := 0
	for treeIdx, tree := range trees {
		points := make([]int, len(tree.Talents))
		if len(parts[treeIdx]) > len(points) {
			t.Fatal("talent string extends beyond the tree")
		}
		for i, digit := range parts[treeIdx] {
			points[i] = int(digit - '0')
			total += points[i]
		}
		for i, talent := range tree.Talents {
			if points[i] < 0 || points[i] > talent.MaxPoints {
				t.Errorf("%s has an invalid rank", talent.FieldName)
			}
			if points[i] == 0 {
				continue
			}
			below := 0
			for j, parent := range tree.Talents {
				if parent.Location.Row < talent.Location.Row {
					below += points[j]
				}
				if talent.Prereq != nil && parent.Location == *talent.Prereq && points[j] != parent.MaxPoints {
					t.Errorf("%s lacks its prerequisite", talent.FieldName)
				}
			}
			if below < talent.Location.Row*5 {
				t.Errorf("%s lacks lower-tier points", talent.FieldName)
			}
		}
	}
	if total != core.CharacterLevel-9 {
		t.Errorf("build spends %d points, want %d", total, core.CharacterLevel-9)
	}
}

func TestRaidAndBearWeaveTalentBuildsAreLegal(t *testing.T) {
	data, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []presetTalentTree
	if err := json.Unmarshal(data, &trees); err != nil {
		t.Fatal(err)
	}
	for _, talents := range []string{buildRequest(t, "raid").Raid.Parties[0].Players[0].TalentsString, "05001-55000032021132212051-055"} {
		validatePresetTalents(t, trees, talents)
	}
}

func TestCatBuildGearMatchesFaction(t *testing.T) {
	items := make(map[int32]*proto.UIItem)
	for _, item := range database.Load().Items {
		items[item.Id] = item
	}
	for _, name := range []string{"starter", "raid"} {
		t.Run(name, func(t *testing.T) {
			player := buildRequest(t, name).Raid.Parties[0].Players[0]
			for _, equipped := range player.Equipment.Items {
				if equipped.Id == 0 {
					continue
				}
				item := items[equipped.Id]
				if item == nil {
					t.Fatalf("equipped item %d is absent from the embedded database", equipped.Id)
				}
				if player.Race == proto.Race_RaceNightElf && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY {
					t.Errorf("Night Elf build equips Horde-only item %d %s", item.Id, item.Name)
				}
				if player.Race == proto.Race_RaceTauren && item.FactionRestriction == proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY {
					t.Errorf("Tauren build equips Alliance-only item %d %s", item.Id, item.Name)
				}
			}
		})
	}
}

func starterRequest(t testing.TB) *proto.RaidSimRequest {
	return buildRequest(t, "starter")
}

func buildRequest(t testing.TB, name string) *proto.RaidSimRequest {
	t.Helper()
	data, err := os.ReadFile("../../../ui/specs/druid/feralcat/builds/" + name + ".build.json")
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

func TestFeralCatStarterCasts(t *testing.T) {
	request := starterRequest(t)
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			request.Raid.Parties[0].Players[0].Race = race
			stats := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid})
			for _, action := range stats.RaidStats.Parties[0].Players[0].RotationStats.PriorityList {
				for _, validation := range action.Validations {
					if validation.LogLevel != proto.LogLevel_Information {
						t.Fatalf("rotation validation: %v", validation)
					}
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
			for _, id := range []int32{9830, 9896, 9904, 1322605} {
				if casts[id] == 0 {
					t.Errorf("starter rotation never casts spell %d", id)
				}
			}
			t.Logf("casts=%v DPS=%g DTPS=%g", casts, player.Dps.Avg, player.Dtps.Avg)
		})
	}
}

func TestComputedNakedFeralStatsAcrossSupportedRaces(t *testing.T) {
	feralbear.RegisterFeralBearDruid()
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren, proto.Race_RaceHighOrderSkyborne, proto.Race_RaceWindshaperSkyborne} {
		if !slices.Contains(core.ClassRaceCapabilities[proto.Class_ClassDruid], race) {
			t.Fatalf("Druid race %v is unsupported", race)
		}
		for _, bear := range []bool{false, true} {
			// Given
			request := coverageRequest(bear, &proto.APLRotation{})
			player := request.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""
			player.Equipment = &proto.EquipmentSpec{}
			// When
			result := core.ComputeStats(&proto.ComputeStatsRequest{Raid: request.Raid, Encounter: request.Encounter})
			// Then
			if result.ErrorResult != "" {
				t.Fatal(result.ErrorResult)
			}
			sheet := result.RaidStats.Parties[0].Players[0].FinalStats
			strength, agility, stamina := sheet.Stats[stats.Strength], sheet.Stats[stats.Agility], sheet.Stats[stats.Stamina]
			wantAP := -20 + 2*strength + 120 + agility
			wantHealth := 1483 + 20 + (stamina-20)*10
			if bear {
				wantAP = -20 + 2*strength + 180
				wantHealth += 1240
			}
			if race == proto.Race_RaceTauren {
				wantHealth *= 1.05
			}
			if got := sheet.Stats[stats.AttackPower]; got != wantAP {
				t.Errorf("%v bear=%v AP=%v, want %v", race, bear, got, wantAP)
			}
			if got := sheet.Stats[stats.Health]; math.Abs(got-wantHealth) > 1e-8 {
				t.Errorf("%v bear=%v health=%v, want %v", race, bear, got, wantHealth)
			}
			t.Logf("race=%v bear=%v strength=%v agility=%v stamina=%v health=%v AP=%v mana=%v dodge=%v", race, bear, strength, agility, stamina, sheet.Stats[stats.Health], sheet.Stats[stats.AttackPower], sheet.Stats[stats.Mana], sheet.PseudoStats[proto.PseudoStat_PseudoStatDodgePercent])
		}
	}
}

func TestEquippedAndUnarmedFeralCombatTablesMatch(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{{"cat", druid.Cat}, {"humanoid", druid.Humanoid}, {"bear", druid.Bear}, {"dire bear", druid.DireBear}} {
		t.Run(tc.name, func(t *testing.T) {
			var first *core.AttackTableHand
			for _, weapon := range []int32{0, 866} {
				// Given
				sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: weapon})
				if tc.form != druid.Cat {
					cat.ClearForm(sim)
				}
				fillResource(sim, cat)
				// When
				if tc.form == druid.Bear && !cat.BearForm.Cast(sim, cat.CurrentTarget) {
					t.Fatal("bear form did not cast")
				}
				if tc.form == druid.DireBear && !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
					t.Fatal("dire bear form did not cast")
				}
				// Then
				table := &cat.AttackTables[cat.CurrentTarget.UnitIndex].MH
				if first == nil {
					first = table
				} else {
					if table.BaseMissChance != first.BaseMissChance || table.BaseDodgeChance != first.BaseDodgeChance || table.BaseGlanceChance != first.BaseGlanceChance {
						t.Errorf("weapon %v changed the level-based combat table", weapon)
					}
				}
				t.Logf("form=%v weapon=%v miss=%v dodge=%v glance=%v", tc.form, weapon, table.BaseMissChance, table.BaseDodgeChance, table.BaseGlanceChance)
			}
		})
	}
}

func TestStrengthAddsTwoAttackPowerAcrossForms(t *testing.T) {
	for _, tc := range []struct {
		name string
		form druid.DruidForm
	}{
		{"cat", druid.Cat},
		{"humanoid", druid.Humanoid},
		{"bear", druid.Bear},
		{"dire bear", druid.DireBear},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			sim, cat := newCatWearing(t, "", nil, map[proto.ItemSlot]int32{proto.ItemSlot_ItemSlotMainHand: 0})
			if tc.form != druid.Cat {
				cat.ClearForm(sim)
			}
			fillResource(sim, cat)
			if tc.form == druid.Bear && !cat.BearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("bear form did not cast")
			}
			if tc.form == druid.DireBear && !cat.DireBearForm.Cast(sim, cat.CurrentTarget) {
				t.Fatal("dire bear form did not cast")
			}
			if !cat.InForm(tc.form) {
				t.Fatalf("expected form %v", tc.form)
			}
			before := cat.GetStat(stats.AttackPower)

			// When
			cat.AddStatDynamic(sim, stats.Strength, 10)

			// Then
			if got := cat.GetStat(stats.AttackPower) - before; got != 20 {
				t.Errorf("10 strength adds %v attack power, want 20", got)
			}
			cat.AddStatDynamic(sim, stats.Strength, -10)
			if got := cat.GetStat(stats.AttackPower); got != before {
				t.Errorf("removing strength leaves attack power %v, want %v", got, before)
			}
		})
	}
}

func TestStrengthBonusAddsTwoAttackPowerToComputedCatStats(t *testing.T) {
	for _, race := range []proto.Race{proto.Race_RaceNightElf, proto.Race_RaceTauren} {
		t.Run(race.String(), func(t *testing.T) {
			rsr := coverageRequest(false, &proto.APLRotation{})
			player := rsr.Raid.Parties[0].Players[0]
			player.Race = race
			player.TalentsString = ""
			player.Equipment.Items[proto.ItemSlot_ItemSlotMainHand].Id = 0
			request := &proto.ComputeStatsRequest{Raid: rsr.Raid, Encounter: rsr.Encounter}
			before := core.ComputeStats(request)
			if before.ErrorResult != "" {
				t.Fatal(before.ErrorResult)
			}
			bonus := make([]float64, len(before.RaidStats.Parties[0].Players[0].FinalStats.Stats))
			bonus[stats.Strength] = 10
			player.BonusStats = &proto.UnitStats{Stats: bonus}
			after := core.ComputeStats(request)
			if after.ErrorResult != "" {
				t.Fatal(after.ErrorResult)
			}
			base := before.RaidStats.Parties[0].Players[0].FinalStats.Stats
			buffed := after.RaidStats.Parties[0].Players[0].FinalStats.Stats
			if got := buffed[stats.AttackPower] - base[stats.AttackPower]; got != 20 {
				t.Errorf("10 bonus strength adds %v computed attack power, want 20", got)
			}
			t.Logf("strength %v→%v; attack power %v→%v", base[stats.Strength], buffed[stats.Strength], base[stats.AttackPower], buffed[stats.AttackPower])
		})
	}
}
