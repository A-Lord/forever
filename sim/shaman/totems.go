package shaman

import (
	"time"

	"github.com/wowsims/forever/sim/core"
)

var windfuryTotemRank = spellData.WindfuryTotem.Highest()
var strengthOfEarthTotemRank = spellData.StrengthOfEarthTotem.Highest()
var graceOfAirTotemRank = spellData.GraceOfAirTotem.ByID(25359)
var manaSpringTotemRank = spellData.ManaSpringTotem.Highest()

func (shaman *Shaman) newTotemSpellConfig(flatCost int32, spellID int32, spellMask int64, gcd time.Duration) core.SpellConfig {
	return core.SpellConfig{
		ActionID:       core.ActionID{SpellID: spellID},
		DefenseType:    core.DefenseTypeMagic,
		Flags:          core.SpellFlagAPL | SpellFlagInstant,
		ClassSpellMask: spellMask,

		ManaCost: core.ManaCostOptions{
			FlatCost: flatCost,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: gcd,
			},
		},
	}
}

// TODO: To be implemented. Port the TBC Windfury Totem Spell implementation below; not yet verified against the Forever client.
func (shaman *Shaman) registerWindfuryTotemSpell() {
	panic("To be implemented")

	// The TBC implementation, kept for the port:
	// duration := time.Second * 120
	// // TODO: Forever drops Improved Weapon Totems; untalented Windfury Totem AP only.
	// value := 445.0
	//
	// wfProcAura := shaman.NewTemporaryStatsAura("Windfury Totem Proc (Self)", core.ActionID{SpellID: 25584}, stats.Stats{stats.AttackPower: value}, time.Millisecond*1500)
	// wfProcAura.MaxStacks = 2
	// wfProcAura.AttachProcTrigger(core.ProcTrigger{
	// 	Name:     "Windfury Attack (Self)",
	// 	Callback: core.CallbackOnSpellHitDealt,
	// 	ProcMask: core.ProcMaskMeleeMHAuto | core.ProcMaskMeleeOHAuto,
	// 	// TriggerImmediately ommited for improved UI clarity (the timeline tick would be near invisible for MHAuto procs)
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		if wfProcAura.IsActive() && !spell.ProcMask.Matches(core.ProcMaskMeleeSpecial) {
	// 			wfProcAura.RemoveStack(sim)
	// 			if wfProcAura.GetStacks() == 0 {
	// 				wfProcAura.Deactivate(sim)
	// 			}
	// 		}
	// 	},
	// })
	//
	// config := shaman.newTotemSpellConfig(int32(windfuryTotemRank.Cost()), windfuryTotemRank.ID, SpellMaskBasicTotem, windfuryTotemRank.GCD())
	//
	// var windfurySpell *core.Spell
	// wfProcTrigger := shaman.MakeProcTriggerAura(core.ProcTrigger{
	// 	Name:               "Windfury Totem Trigger (Self)",
	// 	MetricsActionID:    core.ActionID{SpellID: 25580},
	// 	IsWeaponProc:       true,
	// 	ProcChance:         0.2,
	// 	Duration:           core.NeverExpires,
	// 	Outcome:            core.OutcomeLanded,
	// 	Callback:           core.CallbackOnSpellHitDealt,
	// 	ProcMask:           core.ProcMaskMeleeMHAuto,
	// 	ICD:                time.Millisecond * 1500,
	// 	TriggerImmediately: true,
	// 	Handler: func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
	// 		wfProcAura.Activate(sim)
	// 		if spell.ProcMask == core.ProcMaskMeleeMHAuto {
	// 			wfProcAura.SetStacks(sim, 1)
	// 		} else {
	// 			wfProcAura.SetStacks(sim, 2)
	// 		}
	// 		shaman.AutoAttacks.MaybeReplaceMHSwing(sim, windfurySpell).Cast(sim, result.Target)
	// 	},
	// })
	//
	// wfIntermediateAuraForExclusitivity := shaman.RegisterAura(core.Aura{
	// 	Label:    "Windfury Dummy Aura (self)",
	// 	Duration: time.Second * 10,
	// })
	//
	// wfPartyWeaponBuffTrackingAura := shaman.RegisterAura(core.Aura{
	// 	Label:    "Windfury Party Weapon Buff Tracking Aura",
	// 	Duration: time.Second * 10,
	// 	ActionID: core.ActionID{SpellID: 25580, Tag: 1},
	// })
	//
	// wfAura := shaman.RegisterAura(core.Aura{
	// 	Label:    "Windfury Totem (Self)",
	// 	ActionID: config.ActionID,
	// 	Duration: duration,
	// }).ApplyOnInit(func(aura *core.Aura, sim *core.Simulation) {
	// 	mhConfig := *shaman.AutoAttacks.MHConfig()
	// 	mhConfig.ActionID = mhConfig.ActionID.WithTag(25584)
	// 	windfurySpell = shaman.GetOrRegisterSpell(mhConfig)
	// }).AttachPeriodicAction(core.PeriodicActionOptions{
	// 	Period:          time.Second * 5,
	// 	TickImmediately: true,
	// 	Priority:        core.ActionPriorityAuto,
	// 	OnAction: func(sim *core.Simulation) {
	// 		wfPartyWeaponBuffTrackingAura.Activate(sim)
	// 		wfIntermediateAuraForExclusitivity.Activate(sim)
	// 	},
	// })
	//
	// wfIntermediateAuraForExclusitivity.NewExclusiveEffect(buffs.WindfuryTotemCategory, false, core.ExclusiveEffect{
	// 	Priority: value,
	// 	OnGain: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
	// 		wfProcTrigger.Activate(sim)
	// 	},
	// 	OnExpire: func(_ *core.ExclusiveEffect, sim *core.Simulation) {
	// 		wfProcTrigger.Deactivate(sim)
	// 		wfIntermediateAuraForExclusitivity.Deactivate(sim)
	// 	},
	// })
	//
	// config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 	if shaman.AirTotemAura != nil {
	// 		shaman.AirTotemAura.Deactivate(sim)
	// 	}
	// 	shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
	// 	shaman.AirTotemAura = wfAura
	// 	wfAura.Activate(sim)
	// }
	//
	// shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, slot proto.ItemSlot) {
	// 	wfIntermediateAuraForExclusitivity.Deactivate(sim)
	// })
	//
	// shaman.RegisterSpell(config)
}

// TODO: To be implemented. The body below builds the totem around the generated
// aura, whose value the client states; kept commented until this class's port is reviewed.
func (shaman *Shaman) registerStrengthOfEarthTotemSpell() {
	panic("To be implemented")

	// The implementation, kept for the port:
	// duration := time.Second * 120
	// config := shaman.newTotemSpellConfig(int32(strengthOfEarthTotemRank.Cost()), strengthOfEarthTotemRank.ID, SpellMaskBasicTotem, strengthOfEarthTotemRank.GCD())
	// // Enhancing Totems (16295) has no node in shaman tree 1082, so the aura is untalented.
	// buffAura := buffs.StrengthOfEarthTotemAura(&shaman.Unit, true, 0)
	// buffAura.Duration = duration
	// config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 	if shaman.EarthTotemAura != nil {
	// 		shaman.EarthTotemAura.Deactivate(sim)
	// 	}
	// 	shaman.TotemExpirations[EarthTotem] = sim.CurrentTime + duration
	// 	shaman.EarthTotemAura = buffAura
	// 	buffAura.Activate(sim)
	// }
	// shaman.RegisterSpell(config)
}

// TODO: To be implemented. The body below builds the totem around the generated
// aura, which is the client's 89 agility; Enhancing Totems (16295) has no node
// in shaman tree 1082, so there are no talent points to pass.
func (shaman *Shaman) registerGraceOfAirTotemSpell() {
	panic("To be implemented")

	// The body the port needs:
	// duration := time.Second * 120
	// config := shaman.newTotemSpellConfig(int32(graceOfAirTotemRank.Cost()), graceOfAirTotemRank.ID, SpellMaskBasicTotem, graceOfAirTotemRank.GCD())
	// buffAura := buffs.GraceOfAirTotemAura(&shaman.Unit, true, 0)
	// buffAura.Duration = duration
	// config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 	if shaman.AirTotemAura != nil {
	// 		shaman.AirTotemAura.Deactivate(sim)
	// 	}
	// 	shaman.TotemExpirations[AirTotem] = sim.CurrentTime + duration
	// 	shaman.AirTotemAura = buffAura
	// 	buffAura.Activate(sim)
	// }
	// shaman.RegisterSpell(config)
}

// TODO: To be implemented. The body below builds the totem around the generated
// aura, which Restorative Totems prices; kept commented until this class's port is reviewed.
func (shaman *Shaman) registerManaSpringTotemSpell() {
	panic("To be implemented")

	// The implementation, kept for the port:
	// duration := time.Second * 120
	// config := shaman.newTotemSpellConfig(int32(manaSpringTotemRank.Cost()), manaSpringTotemRank.ID, SpellMaskBasicTotem, manaSpringTotemRank.GCD())
	// buffAura := buffs.ManaSpringTotemAura(&shaman.Unit, true, shaman.Talents.RestorativeTotems)
	// buffAura.Duration = duration
	// config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
	// 	if shaman.WaterTotemAura != nil {
	// 		shaman.WaterTotemAura.Deactivate(sim)
	// 	}
	// 	shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + duration
	// 	shaman.WaterTotemAura = buffAura
	// 	buffAura.Activate(sim)
	// }
	// shaman.RegisterSpell(config)
}

// Commented out upstream of this fork by "only 1 totem per type" (53970b1d84), not by
// the Forever stubbing pass, and never called since. It is kept rather than reduced to
// the usual empty-body no-op because the body is the only record of how the totem was
// modelled; shaman.HealingStreamTotem stays nil until someone revisits totem slots.
/* func (shaman *Shaman) registerHealingStreamTotemSpell() {
	config := shaman.newTotemSpellConfig(3, 5394, SpellMaskBasicTotem, time.Second)
	hsHeal := shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{SpellID: 5394},
		SpellSchool:      core.SpellSchoolNature,
		ProcMask:         core.ProcMaskEmpty,
		Flags:            core.SpellFlagHelpful | core.SpellFlagNoOnCastComplete | SpellFlagInstant,
		DamageMultiplier: 1,
		CritMultiplier:   1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			healing := 28 + spell.HealingPower(target)*0.08272
			spell.CalcAndDealHealing(sim, target, healing, spell.OutcomeHealing)
		},
	})
	config.Hot = core.DotConfig{
		Aura: core.Aura{
			Label: "HealingStreamHot",
		},
		NumberOfTicks: 150,
		TickLength:    time.Second * 2,
		OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
			hsHeal.Cast(sim, target)
		},
	}
	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
		shaman.TotemExpirations[WaterTotem] = sim.CurrentTime + time.Second*300
		for _, agent := range shaman.Party.Players {
			spell.Hot(&agent.GetCharacter().Unit).Activate(sim)
		}
	}
	shaman.HealingStreamTotem = shaman.RegisterSpell(config)
} */
