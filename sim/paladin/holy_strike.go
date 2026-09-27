package paladin

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var HolyStrikeRankMap = spellData.HolyStrike

// Holy Strike
// https://www.wowhead.com/forever/spell=10333
//
// An instant strike that causes 40% weapon damage plus an additional 93 as Holy damage.
//
// The row states the flat part as its normalized-weapon-damage effect and the percentage as a
// second effect; the tooltip reads them as 40% of weapon damage plus the flat number, and so
// does this. The flat part carries the spell power coefficient.
func (paladin *Paladin) registerHolyStrike(_ int32, rank *spelldata.Spell) {
	flat := rank.DamageEffect()
	weaponPercent := rank.EffectN(2).Percent()

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
		BonusCoefficient: flat.Coeff(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := flat.Roll(sim, core.CharacterLevel) + weaponPercent*spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}
