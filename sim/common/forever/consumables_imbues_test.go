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

// Scroll of Imbue Baleflame 277487 puts enchant 8706 on a staff, whose equip spell 1306572 states
// A_MOD_TARGET_RESISTANCE -10 on fire alone, and Scroll of Imbue Balefrost 277496 puts 8714 on it,
// whose 1306580 states -15 on frost alone. Each pierces its own school's resistance and neither
// adds SpellPiercing, which would reach every school.
func TestBaleImbuesPierceOneSchool(t *testing.T) {
	const staff = 873
	for _, c := range []struct {
		imbue       int32
		fire, frost float64
	}{
		{0, 0, 0},
		{277487, 10, 0},
		{277496, 0, 15},
	} {
		_, character := warriorWithConsumes(t, &proto.ConsumesSpec{MhImbueId: c.imbue}, staff, 0, proto.MobType_MobTypeUnknown)

		piercing := character.PseudoStats.SchoolSpellPiercing
		if piercing[stats.SchoolIndexFire] != c.fire || piercing[stats.SchoolIndexFrost] != c.frost {
			t.Errorf("imbue %d: fire and frost piercing %v and %v, want %v and %v", c.imbue,
				piercing[stats.SchoolIndexFire], piercing[stats.SchoolIndexFrost], c.fire, c.frost)
		}
		if got := character.GetStat(stats.SpellPiercing); got != 0 {
			t.Errorf("imbue %d: SpellPiercing %v, want 0", c.imbue, got)
		}
	}
}
