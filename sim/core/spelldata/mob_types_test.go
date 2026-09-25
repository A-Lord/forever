package spelldata

import (
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

func TestMobTypesReadsTheCreatureTypeMask(t *testing.T) {
	for mask, want := range map[int32][]proto.MobType{
		4:   {proto.MobType_MobTypeDemon},
		32:  {proto.MobType_MobTypeUndead},
		36:  {proto.MobType_MobTypeDemon, proto.MobType_MobTypeUndead},
		128: nil,
		256: {proto.MobType_MobTypeMechanical},
	} {
		if got := MobTypes(mask); !slices.Equal(got, want) {
			t.Errorf("MobTypes(%d) = %v, want %v", mask, got, want)
		}
	}
}

func TestAVersusAuraFollowsItsAuraOnTheAttackTables(t *testing.T) {
	withRows(t, parseRows())
	sim := &core.Simulation{}
	character := parseWarrior()
	undead := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	character.AttackTables = []*core.AttackTable{undead}

	row := oneEffectRow(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, 32, 99)
	aura := character.RegisterAura(AuraConfig(row))
	ParseEffects(character, aura, row)

	aura.Activate(sim)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.AttackPower]; got != 99 {
		t.Errorf("attack power against Undead while up: %v, want 99", got)
	}
	aura.Deactivate(sim)
	if got := undead.MobTypeBonusStats[proto.MobType_MobTypeUndead][stats.AttackPower]; got != 0 {
		t.Errorf("attack power against Undead after expiry: %v, want 0", got)
	}

	if parsed := ParseStatic(parseWarrior(), row); len(parsed.Applied) != 0 {
		t.Errorf("the static path attached %v, want nothing before the attack tables exist", appliedKinds(parsed))
	}
}

// Per stat, a bonus against one creature type bids apart from the same stat against another, and
// competes only with the same stat against the same mask.
func TestAVersusAuraBidsPerCreatureTypeMask(t *testing.T) {
	sim := &core.Simulation{}
	character := parseWarrior()
	undead := core.NewAttackTable(&character.Unit, &core.Unit{Type: core.EnemyUnit, MobType: proto.MobType_MobTypeUndead})
	character.AttackTables = []*core.AttackTable{undead}

	register := func(label string, mask int32, points float64) *core.Aura {
		row := buffRow(7070, auraEffect(dbcenums.A_MOD_MELEE_ATTACK_POWER_VERSUS, mask, points))
		aura := character.RegisterAura(AuraConfig(row, Label(label)))
		ParseEffects(character, aura, row, ExclusivePerStat("StatBuff"))
		return aura
	}
	againstUndead := register("Undead", 32, 99)
	againstDemons := register("Demons", 4, 185)
	weakerAgainstUndead := register("Weaker Undead", 32, 36)

	if a, b := categoryNames(againstUndead), categoryNames(againstDemons); slices.Equal(a, b) {
		t.Fatalf("masks 32 and 4 bid in the same category %v", a)
	}
	if a, b := categoryNames(againstUndead), categoryNames(weakerAgainstUndead); !slices.Equal(a, b) {
		t.Fatalf("two bonuses against mask 32 bid in %v and %v, want one category", a, b)
	}

	againstUndead.Activate(sim)
	againstDemons.Activate(sim)
	weakerAgainstUndead.Activate(sim)
	bonus := undead.MobTypeBonusStats
	if got := bonus[proto.MobType_MobTypeUndead][stats.AttackPower]; got != 99 {
		t.Errorf("attack power against Undead: %v, want the stronger 99 alone", got)
	}
	if got := bonus[proto.MobType_MobTypeDemon][stats.AttackPower]; got != 185 {
		t.Errorf("attack power against Demons: %v, want 185", got)
	}
}
