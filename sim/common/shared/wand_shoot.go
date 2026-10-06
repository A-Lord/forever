package shared

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shootRow = spelldata.MustFind(5019)

// Shoot is a channel whose ticks are the wand's shots, one per wand speed, until something takes a
// tick's slot. A shot rolls the wand's damage in its school on the spell table: no attack power, no
// spell power. It crits on the ranged crit chance, for 1.5 times like a spell.
func RegisterShoot(character *core.Character) *core.Spell {
	weapon := character.WeaponFromRanged()
	wandSpeed := func() time.Duration {
		if wand := character.GetRangedWeapon(); wand != nil {
			return core.DurationFromSeconds(max(wand.SwingSpeed, 1))
		}
		return 0
	}

	config := spelldata.SpellConfig(&character.Unit, shootRow, spelldata.Magic(core.ProcMaskRangedAuto))
	config.Flags |= core.SpellFlagChanneled | core.SpellFlagAPL | core.SpellFlagNoPartialResist
	config.SpellSchool = weapon.GetSpellSchool()
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !character.Moving
	}
	config.Dot = core.DotConfig{
		Aura:                core.Aura{Label: "Shoot"},
		TickLength:          core.DurationFromSeconds(max(weapon.SwingSpeed, 1)),
		AffectedByCastSpeed: true,
		Autorepeat:          true,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			spell := dot.Spell
			baseDamage := weapon.CalculateWeaponDamage(sim, 0)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndRangedCrit)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weapon = character.WeaponFromRanged()
		spell.SpellSchool = weapon.GetSpellSchool()
		spell.SchoolIndex = spell.SpellSchool.MultiplierSchool().SchoolIndex()
		dot := spell.Dot(target)
		dot.BaseTickLength = core.DurationFromSeconds(max(weapon.SwingSpeed, 1))
		dot.Apply(sim)
	}

	shoot := character.RegisterSpell(config)
	character.EnableWand(shoot, wandSpeed)
	return shoot
}
