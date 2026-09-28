package hunter

import (
	"strconv"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var summonHawkRank = spellData.SummonHawk.Highest()
var summonHawkBaseDamage = summonHawkRank.DamageEffect().Average(core.CharacterLevel)

// The creature the dive leaves behind, up for 18 s.
var hawkSummon = spellData.SummonHawkTriggered.ByID(1293248)

// The tooltip's ${$s1+($rap*($s4/100))}: effect 4 is the share of ranged attack power.
var summonHawkAPShare = summonHawkRank.EffectN(4).Percent()

// TODO: In-game testing required. The assault is a server script the rows do not state: read as one
// attack every 3 s for as long as the hawk is up, each dealing what the dive does.
const hawkAttackInterval = 3 * time.Second

// Effect 3 is the number of hawks up at once; a hawk past it replaces the one with the least time
// left.
func (hunter *Hunter) registerSummonHawk() {
	if !hunter.Talents.SummonHawk {
		return
	}

	hawks := make([]*core.Spell, int(summonHawkRank.EffectN(3).BaseValue()))
	for i := range hawks {
		// The hawk's own attacks: they are not the hunter's ranged hits, so they match only listeners
		// that state no mask.
		config := spelldata.SpellConfig(&hunter.Unit, summonHawkRank,
			spelldata.Melee(core.ProcMaskEmpty), spelldata.Proc(), spelldata.Tag(int32(i+1)))
		config.Dot = core.DotConfig{
			Aura:          spelldata.AuraConfig(hawkSummon, spelldata.Label("Summon Hawk "+strconv.Itoa(i+1))),
			TickLength:    hawkAttackInterval,
			NumberOfTicks: int32(hawkSummon.Duration() / hawkAttackInterval),
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				baseDamage := summonHawkBaseDamage + summonHawkAPShare*dot.Spell.RangedAttackPower(target)
				dot.Spell.CalcAndDealDamage(sim, target, baseDamage, dot.Spell.OutcomeRangedCritOnly)
			},
		}
		hawks[i] = hunter.RegisterSpell(config)
	}
	attacking := make([]*core.Dot, len(hawks))

	config := spelldata.SpellConfig(&hunter.Unit, summonHawkRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := summonHawkBaseDamage + summonHawkAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedCritOnly)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)

			slot := 0
			for i, dot := range attacking {
				if dot == nil || !dot.IsActive() {
					slot = i
					break
				}
				if dot.RemainingDuration(sim) < attacking[slot].RemainingDuration(sim) {
					slot = i
				}
			}
			if attacking[slot] != nil {
				attacking[slot].Deactivate(sim)
			}
			attacking[slot] = hawks[slot].Dot(target)
			attacking[slot].Apply(sim)
		})
	}

	hunter.RegisterSpell(config)
}
