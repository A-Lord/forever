package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var HolyStrikeRankMap = spellData.HolyStrike

// Holy Strike
// https://www.wowhead.com/forever/spell=10333
//
// An instant strike that causes 40% weapon damage plus an additional 93 as Holy damage.
//
// The row states the flat part as its normalized-weapon-damage effect and the percentage as a
// second effect; the tooltip reads them as 40% of weapon damage plus the flat number, and so
// does this.
//
// The spell power coefficient sits on the flat effect, but the paladin's own spell power goes
// through the weapon percent, as it does for the Seal of Command proc: Rank 3's 0.429 at 32% is
// 0.137 per point of spell power. The Holy damage the target takes extra, Judgement of the Crusader
// and the damage-against-mob-type gear, is added after the percent at the full coefficient.
func (paladin *Paladin) registerHolyStrike(_ int32, rank *spelldata.Spell) {
	flat := rank.DamageEffect()
	weaponPercent := rank.EffectN(2).Percent()
	coefficient := flat.Coeff()

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: rank.ID},
		SpellSchool:    rank.SpellSchool(),
		DefenseType:    rank.DefenseTypeCore(),
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyStrike,
		Rank:           rank.RankNumber(),
		MaxRange:       core.MaxMeleeRange,

		ManaCost: manaCost(rank),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: rank.GCD(),
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.holyStrikeTimer),
				Duration: cooldown(rank),
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		// No BonusCoefficient: the spell power goes through the weapon percent, so the strike adds
		// it by hand below rather than letting CalcDamage add it on top.

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			weaponDamage := spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			baseDamage := flat.Roll(sim, core.CharacterLevel) + (weaponDamage+coefficient*spell.SpellDamage(target))*weaponPercent

			targetBonus := target.PseudoStats.SchoolBonusSpellDamage[stats.SchoolIndexHoly] +
				spell.Unit.AttackTables[target.UnitIndex].MobTypeBonusStats[target.MobType][stats.SpellDamage]
			baseDamage += coefficient * targetBonus

			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}
