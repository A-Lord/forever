//go:build with_db

package feralcat

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
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
