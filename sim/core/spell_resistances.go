package core

import (
	"math"

	"github.com/wowsims/forever/sim/core/stats"
)

func (result *SpellResult) applyResistances(sim *Simulation, spell *Spell, isPeriodic bool, attackTable *AttackTable) {
	resistanceMultiplier, outcome := spell.ResistanceMultiplier(sim, isPeriodic, attackTable)

	result.Damage *= resistanceMultiplier
	result.Outcome |= outcome

	result.ArmorAndResistanceMultiplier = resistanceMultiplier
}

// Modifies damage based on Armor or Magic resistances, depending on the damage type.
func (spell *Spell) ResistanceMultiplier(sim *Simulation, isPeriodic bool, attackTable *AttackTable) (float64, HitOutcome) {
	if spell.Flags.Matches(SpellFlagIgnoreResists) {
		return 1, OutcomeEmpty
	}

	if spell.SpellSchool.Matches(SpellSchoolPhysical) {
		// All physical dots (Bleeds) ignore armor.
		if isPeriodic && !spell.Flags.Matches(SpellFlagApplyArmorReduction) {
			return 1, OutcomeEmpty
		}

		// Physical resistance (armor).
		return attackTable.GetArmorDamageModifier(spell), OutcomeEmpty
	}

	// Magical resistance.
	if spell.Flags.Matches(SpellFlagBinary | SpellFlagNoPartialResist) {
		return 1, OutcomeEmpty
	}

	averageResist := attackTable.Defender.averageResist(spell, attackTable.Attacker, false, isPeriodic)
	if averageResist == 0 {
		return 1, OutcomeEmpty
	}

	tenths := rollResistTenths(sim, math.Abs(averageResist))
	if tenths == 0 {
		return 1, OutcomeEmpty
	}
	if averageResist < 0 {
		return 1 + 0.1*float64(tenths), OutcomeEmpty
	}
	return 1 - 0.1*float64(tenths), partialResistOutcome(tenths)
}

func (at *AttackTable) GetBinaryHitChance(spell *Spell) float64 {
	return at.Defender.binaryHitChance(spell, at.Attacker)
}

// The average share of a spell's damage the target resists, R / (R + 5L + 80) for the caster's level L,
// with the caster's Spell Piercing taken off R. Piercing beyond the resistance makes the share negative,
// and the roll pays it out as bonus damage. The target's own resistance stops at 0, so a debuff that
// lowers it below 0 adds nothing.
func (unit *Unit) averageResist(spell *Spell, attacker *Unit, binary bool, isPeriodic bool) float64 {
	if spell.SchoolIndex <= stats.SchoolIndexPhysical {
		return 0
	}

	resistance := 0.0
	if stat, ok := spell.SpellSchool.ResistanceStat(); ok {
		piercing := attacker.stats[stats.SpellPiercing] + attacker.PseudoStats.SchoolSpellPiercing[spell.SchoolIndex]
		resistance = max(0, unit.GetStat(stat)) - piercing
	}
	// TODO: 5L + 80 and 10L (as R / 10L) both fit the level 20-24 Spell Piercing logs; a level-30 log
	// decides, with 46% of hits binned at 15 piercing under 5L + 80 and 38% under 10L (#76).
	resistanceConstant := 5*float64(attacker.Level) + 80
	averageResist := resistance / (math.Abs(resistance) + resistanceConstant)

	// A DoT's tick, a hybrid's included, rolls its partial resist against a tenth of the resistance.
	if isPeriodic && !spell.Flags.Matches(SpellFlagChanneled|SpellFlagFullResistanceTicks) {
		averageResist /= 10
	}

	// Level-based resistance does not apply to binary spells.
	if !binary && unit.Type == EnemyUnit && unit.Level > attacker.Level {
		averageResist += 0.02 * float64(combatLevelDiff(attacker, unit))
	}

	return min(1, averageResist)
}

// A binary spell is resisted in full with the target's average resist as the chance.
func (unit *Unit) binaryHitChance(spell *Spell, attacker *Unit) float64 {
	return 1 - max(0, unit.averageResist(spell, attacker, true, false))
}

// The partial resist roll, in tenths of the damage, around the average share a. Up to a = 10% it gives
// 0, 1 or 2 tenths at 1 - 7.5a, 5a and 2.5a; from 90% the same mirrored from 10 tenths; in between each
// 10% step x within 20% of a at 0.5 - 2.5|x - a|.
func rollResistTenths(sim *Simulation, a float64) int {
	roll := sim.RandomFloat("Partial Resist")
	if a <= 0.1 {
		switch {
		case roll < 1-7.5*a:
			return 0
		case roll < 1-2.5*a:
			return 1
		default:
			return 2
		}
	}
	if a >= 0.9 {
		switch {
		case roll < 1-7.5*(1-a):
			return 10
		case roll < 1-2.5*(1-a):
			return 9
		default:
			return 8
		}
	}

	cumulative, last := 0.0, 0
	for tenths := 0; tenths <= 10; tenths++ {
		chance := 0.5 - 2.5*math.Abs(float64(tenths)/10-a)
		if chance <= 0 {
			continue
		}
		cumulative += chance
		last = tenths
		if roll < cumulative {
			return tenths
		}
	}
	return last
}

func (at *AttackTable) GetArmorDamageModifier(spell *Spell) float64 {
	if at.IgnoreArmor {
		return 1.0
	}

	ignoreArmorFactor := Clamp(at.ArmorIgnoreFactor, 0.0, 1.0)

	armorConstant := 400 + 85*float64(at.Attacker.Level)
	defenderArmor := at.Defender.Armor() - (at.Defender.Armor() * ignoreArmorFactor)
	// TODO: Forever possibly has different rules as there is Armor Piercing Flat and % values
	defenderArmor = max(defenderArmor-at.Attacker.stats[stats.ArmorPenetration], 0)
	// Damage reduction from armor is capped at 75%.
	return max(1-defenderArmor/(defenderArmor+armorConstant), 0.25)
}
