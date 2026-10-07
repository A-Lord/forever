package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Package-level state the commented-out implementations used:
// var waterShieldRank = spellData.WaterShield.ByID(33736)

func (shaman *Shaman) registerShieldsSpells() {
	shaman.registerWaterShieldSpell()
	shaman.registerLightningShieldSpell()
}

// The shield proc rate option stands for the hits a shaman takes when nothing attacks it in the sim.
func (shaman *Shaman) startShieldProcPeriodicAction(sim *core.Simulation) {
	if shaman.SelfBuffs.ShieldProcrate > 0 && shaman.ShieldSelfProc != nil {
		core.StartPeriodicAction(sim, core.PeriodicActionOptions{
			Period:   time.Duration(float64(time.Minute) / shaman.SelfBuffs.ShieldProcrate),
			Priority: core.ActionPriorityGCD,
			OnAction: func(sim *core.Simulation) {
				shaman.ShieldSelfProc(sim)
			},
		})
	}
}

// TODO: To be implemented. The ability exists: spell 408510 on the Restoration line. No rank subtext,
// so no generated table -- pin the id directly.
func (shaman *Shaman) registerWaterShieldSpell() {
	// The TBC implementation, kept for the port:
	// bonusManaReturned := 0.0
	// mp5 := 50.0
	// if shaman.CouldHaveSetBonus(ItemSetTidefuryRaiment, 4) {
	// 	bonusManaReturned = 56
	// }
	//
	// actionID := core.ActionID{SpellID: waterShieldRank.ID}
	// waterShieldManaMetrics := shaman.NewManaMetrics(actionID)
	//
	// shaman.WaterShieldAura = shaman.RegisterAura(core.Aura{
	// 	Label:     "Water Shield",
	// 	ActionID:  actionID,
	// 	Duration:  10 * time.Minute,
	// 	MaxStacks: 3,
	// }).AttachProcTrigger(core.ProcTrigger{
	// 	Name:           "Water Shield Trigger",
	// 	Callback:       core.CallbackOnSpellHitTaken,
	// 	ICD:            3500 * time.Millisecond,
	// 	ClassSpellMask: SpellMaskShieldSelfProc,
	// 	Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
	// 		shaman.WaterShieldAura.RemoveStack(sim)
	// 		shaman.AddMana(sim, waterShieldRank.DamageEffect().Average(core.CharacterLevel)+bonusManaReturned, waterShieldManaMetrics)
	// 	},
	// }).AttachStatBuff(stats.MP5, mp5)
	//
	// shaman.RegisterSpell(core.SpellConfig{
	// 	ActionID:    actionID,
	// 	SpellSchool: core.SpellSchoolNature,
	// 	DefenseType: core.DefenseTypeMagic,
	// 	Flags:       core.SpellFlagAPL | SpellFlagInstant,
	// 	Cast: core.CastConfig{
	// 		DefaultCast: core.Cast{
	// 			GCD: waterShieldRank.GCD(),
	// 		},
	// 	},
	// 	ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 		shaman.LightningShieldAura.Deactivate(sim)
	// 		shaman.WaterShieldAura.Activate(sim)
	// 		shaman.WaterShieldAura.SetStacks(sim, 3)
	// 	},
	// 	RelatedSelfBuff: shaman.WaterShieldAura,
	// })
}

var lightningShieldRank = spellData.LightningShield.Highest()

// The shield's effect 1 fires 26545, a dummy, and the orb its tooltip names deals the damage.
var lightningShieldOrb = lightningShieldRank.Refs()[0]

// 10432 holds three charges for 10 minutes, and a hit taken spends one, at most once every 3.5 s, to
// strike the attacker with the orb.
func (shaman *Shaman) registerLightningShieldSpell() {
	orbConfig := spelldata.SpellConfig(&shaman.Unit, lightningShieldOrb,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	orbConfig.ClassSpellMask = SpellMaskLightningShield
	damage := lightningShieldOrb.DamageEffect()
	orbConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		// The row cannot crit.
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHit)
	}
	orb := shaman.RegisterSpell(orbConfig)

	shaman.LightningShieldAura = shaman.RegisterAura(spelldata.AuraConfig(lightningShieldRank))

	// The cooldown is the trigger's own, shared with the shield proc rate option.
	icd := core.Cooldown{
		Timer:    shaman.NewTimer(),
		Duration: lightningShieldRank.ICD(),
	}
	spendCharge := func(sim *core.Simulation, attacker *core.Unit) {
		if !shaman.LightningShieldAura.IsActive() || !icd.IsReady(sim) {
			return
		}
		icd.Use(sim)
		shaman.LightningShieldAura.RemoveStack(sim)
		orb.Cast(sim, attacker)
	}
	shaman.ShieldSelfProc = func(sim *core.Simulation) {
		spendCharge(sim, shaman.CurrentTarget)
	}

	trigger := spelldata.ProcTrigger(&shaman.Character, lightningShieldRank, func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
		spendCharge(sim, spell.Unit)
	})
	trigger.Name = "Lightning Shield Trigger"
	trigger.ICD = 0
	shaman.LightningShieldAura.AttachProcTrigger(trigger)

	config := spelldata.SpellConfig(&shaman.Unit, lightningShieldRank, spelldata.Flags(core.SpellFlagAPL|SpellFlagShamanSpell|SpellFlagInstant))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = SpellMaskLightningShield
	config.RelatedSelfBuff = shaman.LightningShieldAura
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		if shaman.WaterShieldAura != nil {
			shaman.WaterShieldAura.Deactivate(sim)
		}
		shaman.LightningShieldAura.Activate(sim)
		shaman.LightningShieldAura.SetStacks(sim, shaman.LightningShieldAura.MaxStacks)
	}
	shaman.LightningShield = shaman.RegisterSpell(config)
}
