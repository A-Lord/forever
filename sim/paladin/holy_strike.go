package paladin

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
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
func (paladin *Paladin) registerHolyStrike(row shared.SpellData) {
	weaponPercent := effectAt(row, 1).Value / 100

	paladin.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: row.SpellID},
		SpellSchool:    row.SpellSchool,
		DefenseType:    row.DefenseType,
		ProcMask:       core.ProcMaskMeleeMHSpecial,
		Flags:          core.SpellFlagMeleeMetrics | core.SpellFlagAPL,
		ClassSpellMask: SpellMaskHolyStrike,
		Rank:           row.Rank,
		MaxRange:       core.MaxMeleeRange,

		ManaCost: manaCost(row),
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: row.GCD,
			},
			CD: core.Cooldown{
				Timer:    paladin.sharedTimer(&paladin.holyStrikeTimer),
				Duration: row.Cooldown,
			},
		},

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		BonusCoefficient: row.Direct.BonusCoefficient(),

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := row.Direct.Damage(sim) + weaponPercent*spell.Unit.MHNormalizedWeaponDamage(sim, spell.MeleeAttackPower(target))
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeSpecialHitAndCrit)
		},
	})
}
