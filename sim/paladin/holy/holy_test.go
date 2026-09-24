package holy

import (
	"testing"

	"github.com/wowsims/forever/sim/common"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

func init() {
	RegisterHolyPaladin()
	common.RegisterAllEffects()
}

// Stats-only suite: this spec is a gear planner, it has no healing rotation.
// Pins the final character stats for each gear preset so the passives stay covered. The empty APL
// rotation and the fake prepull (no SkipRotation) make it exercise a full environment reset, the
// path the UI's stats request takes.
func TestHolyPaladin(t *testing.T) {
	t.Skip("class talents and abilities are stubbed pending their Forever implementations; " +
		"the golden numbers cannot be meaningful until then")
	var generators []core.TestGenerator
	for _, gearSet := range []string{"preraid", "p3"} {
		player := core.WithSpec(
			&proto.Player{
				Class:         proto.Class_ClassPaladin,
				Race:          proto.Race_RaceUndead,
				Equipment:     core.GetGearSet("../../../ui/specs/paladin/holy/gear_sets", gearSet).GearSet,
				Consumables:   FullConsumes,
				Buffs:         core.FullIndividualBuffs,
				TalentsString: StandardTalents,
				Profession1:   proto.Profession_Enchanting,
				Profession2:   proto.Profession_Jewelcrafting,
				Rotation:      &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
			},
			PlayerOptions,
		)
		generators = append(generators, &core.SingleCharacterStatsTestGenerator{
			Name: gearSet,
			Request: &proto.ComputeStatsRequest{
				Raid: core.SinglePlayerRaidProto(player, core.FullPartyBuffs, core.FullRaidBuffs, core.FullDebuffs),
			},
		})
	}
	core.RunTestSuite(t, t.Name(), generators)
}

// 45/11/5, wowhead's TBC raid build.
var StandardTalents = "05503121520132531051-500231-5"

var FullConsumes = &proto.ConsumesSpec{
	FlaskId: 22853, // Flask of Mighty Restoration
	FoodId:  27666, // Golden Fish Sticks
	PotId:   22832, // Super Mana Potion
}

var PlayerOptions = &proto.Player_HolyPaladin{
	HolyPaladin: &proto.HolyPaladin{
		Options: &proto.HolyPaladin_Options{
			ClassOptions: &proto.PaladinOptions{},
		},
	},
}
