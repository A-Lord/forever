package balance

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/common"
	_ "github.com/wowsims/forever/sim/common" // imported to get caster sets included. (we use spellfire here)
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
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
