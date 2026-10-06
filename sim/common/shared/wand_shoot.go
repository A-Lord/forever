package shared

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

var shootRow = spelldata.MustFind(5019)

// Shoot is a channel whose ticks are the wand's shots, one per wand speed, until something takes a
// tick's slot. A shot rolls the wand's damage in its school on the spell table: no attack power, no
// spell power.
func RegisterShoot(character *core.Character) *core.Spell {
	weapon := character.WeaponFromRanged()
	swingSpeed := core.DurationFromSeconds(max(weapon.SwingSpeed, 1))

	config := spelldata.SpellConfig(&character.Unit, shootRow, spelldata.Magic(core.ProcMaskRangedAuto))
	config.Flags |= core.SpellFlagChanneled | core.SpellFlagAPL
	config.SpellSchool = weapon.GetSpellSchool()
	config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
		return !character.Moving
	}
	config.Dot = core.DotConfig{
		Aura:                core.Aura{Label: "Shoot"},
		NumberOfTicks:       1 << 20,
		TickLength:          swingSpeed,
		AffectedByCastSpeed: true,
		Autorepeat:          true,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			spell := dot.Spell
			baseDamage := weapon.CalculateWeaponDamage(sim, 0)
			result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
			spell.Unit.Wand.RecordShot(sim)
			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				spell.DealDamage(sim, result)
			})
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.Dot(target).Apply(sim)
	}

	shoot := character.RegisterSpell(config)
	var wandSpeed time.Duration
	if character.GetRangedWeapon() != nil {
		wandSpeed = swingSpeed
	}
	character.EnableWand(shoot, wandSpeed)
	return shoot
}
