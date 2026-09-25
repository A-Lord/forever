//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/common/forever"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func init() {
	forever.RegisterAllTemporaryEnchantConsumables()
}

// Elemental Sharpening Stone 18262 puts enchant 2506 on the weapon, whose equip spell 22755 states
// A_MOD_WEAPON_CRIT_PERCENT 2 on the weapons its SpellEquippedItems mask 42483 names, all of them
// melee. db.json carries it as 2% melee crit and a ranged crit total of 0, so a melee attack gains
// 2% crit and a ranged one nothing.
func TestElementalSharpeningStoneGrantsMeleeCrit(t *testing.T) {
	const sword = 727
	_, bare := warriorWithConsumes(t, &proto.ConsumesSpec{}, sword, 0, proto.MobType_MobTypeUnknown)
	_, stoned := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: 18262}, sword, 0, proto.MobType_MobTypeUnknown)

	gained := stoned.GetStats().Subtract(bare.GetStats())
	melee := gained[stats.PhysicalCritPercent]
	ranged := gained[stats.PhysicalCritPercent] + gained[stats.RangedCritPercent]
	if melee != 2 || ranged != 0 {
		t.Errorf("the stone adds %v%% melee and %v%% ranged crit, want 2 and 0", melee, ranged)
	}
}
