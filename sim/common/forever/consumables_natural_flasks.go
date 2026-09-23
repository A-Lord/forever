package forever

import (
	"fmt"
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Flasks of Natural Accuracy 274273, Aggression 274274, Precision 274275 and Swiftness 274276. The
// first effect of each spell is its Stamina. The second, an A_DUMMY, is the bonus the spell's text
// grants in Mount Hyjal, Hyjal Summit and the Barrow Deeps, where the sim takes every fight to be; the
// auras after it name what the bonus is and state 0 themselves.
var NaturalFlasks = []int32{274273, 274274, 274275, 274276}

func init() {
	for _, itemID := range NaturalFlasks {
		core.NewConsumableEffect(itemID, func(agent core.Agent) {
			character := agent.GetCharacter()
			consumable := core.GetConsumableByID(itemID)
			row := spelldata.MustFind(consumable.ItemEffects[0].BuffId)

			config := spelldata.AuraConfig(row, spelldata.Label(consumable.Name))
			config.ActionID = core.ActionID{ItemID: itemID}
			config.BuildPhase = core.CharacterBuildPhaseConsumes
			spelldata.Permanent()(&config)
			aura := character.GetOrRegisterAura(config)

			spelldata.ParseEffects(character, aura, row, spelldata.Effects(1))
			attachNaturalFlaskZoneBonus(aura, row, row.EffectN(2).Average(character.Level))
		})
	}
}

// The zone bonus as the auras after the dummy name it: hit, crit, a lower chance to be dodged or
// parried, or melee, ranged and casting speed, each by the dummy's percent.
func attachNaturalFlaskZoneBonus(aura *core.Aura, row *spelldata.Spell, percent float64) {
	var bonus stats.Stats
	attackSpeed, castSpeed := 1.0, 1.0
	for _, e := range row.Effects[2:] {
		switch e.Aura {
		case dbcenums.A_MOD_HIT_CHANCE:
			bonus[stats.PhysicalHitPercent] = percent
		case dbcenums.A_MOD_SPELL_HIT_CHANCE:
			bonus[stats.SpellHitPercent] = percent
		case dbcenums.A_MOD_CRIT_PCT:
			bonus[stats.PhysicalCritPercent] = percent
			bonus[stats.SpellCritPercent] = percent
		case dbcenums.A_MOD_EXPERTISE:
			// Four points of expertise to the percent.
			bonus[stats.ExpertiseRating] = percent * 4 * core.ExpertisePerQuarterPercentReduction
		case dbcenums.A_MOD_MELEE_RANGED_HASTE_2:
			attackSpeed = 1 + percent/100
		case dbcenums.A_MOD_CASTING_SPEED_NOT_STACK:
			castSpeed = 1 + percent/100
		default:
			panic(fmt.Sprintf("%d %s: no zone bonus for %v", row.ID, row.Name, e.Aura))
		}
	}

	speeds := func(aura *core.Aura, sim *core.Simulation, exponent float64) {
		if attackSpeed != 1 {
			aura.Unit.MultiplyAttackSpeed(sim, math.Pow(attackSpeed, exponent))
		}
		if castSpeed != 1 {
			aura.Unit.MultiplyCastSpeed(sim, math.Pow(castSpeed, exponent))
		}
	}
	aura.ApplyOnGain(func(aura *core.Aura, sim *core.Simulation) {
		aura.Unit.AddStatsDynamic(sim, bonus)
		speeds(aura, sim, 1)
	}).ApplyOnExpire(func(aura *core.Aura, sim *core.Simulation) {
		aura.Unit.AddStatsDynamic(sim, bonus.Invert())
		speeds(aura, sim, -1)
	})
}
