//go:build with_db

package forever_test

import (
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Graverobber's Shovel 273817's equip spell 1319047 grants 12 attack power and 12 ranged attack
// power, which the item's stats carry, and 12 of each against Undead (creature type mask 32).
func TestGraverobbersShovelGrantsAttackPowerAgainstUndead(t *testing.T) {
	const shovel = 273817
	_, character := warriorWithConsumes(t, &proto.ConsumesSpec{}, shovel, 0)

	if got, want := bonusAgainst(character, proto.MobType_MobTypeUndead), (stats.Stats{stats.AttackPower: 12, stats.RangedAttackPower: 12}); got != want {
		t.Errorf("the bonus against Undead is %v, want %v", got, want)
	}
	if got := bonusAgainst(character, proto.MobType_MobTypeDemon); got != (stats.Stats{}) {
		t.Errorf("the bonus against Demons is %v, want none", got)
	}
}
