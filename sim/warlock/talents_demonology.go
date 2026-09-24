package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

func (warlock *Warlock) registerDemonologyTalents() {
	// Tier 1
	warlock.applyImprovedHealthFunnel()
	warlock.applyImprovedImp()
	warlock.applyDemonicEmbrace()
	warlock.applyUnholyPower()

	// Tier 2
	// Demonic Aegis implemented in armors.go
	warlock.applyImprovedVoidwalker()
	warlock.applyFelVitality()
	warlock.applyDemonicEnergies()

	// Tier 3
	warlock.applyImprovedSayaad()
	warlock.applyDemonicSacrifice()
	warlock.applyMasterSummoner()

	// Tier 4
	warlock.applyDecimation()
	warlock.applyFelDomination()
	warlock.applyDemonicBrand()

	// Tier 5
	warlock.applyImprovedFelhunter()
	warlock.applySoulLink()
	warlock.applyDemonicKnowledge()

	// Tier 6
	warlock.applyMasterDemonologist()

	// Tier 7
	warlock.applyDemonicPact()
}

// TODO: To be implemented. Port the TBC Improved Imp implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyImprovedImp() {
	if warlock.Talents.ImprovedImp == 0 || warlock.Options.SacrificeSummon {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ImprovedImp == 0 || warlock.Options.SacrificeSummon {
	// 	return
	// }
	//
	// warlock.Imp.AddStaticMod(core.SpellModConfig{
	// 	Kind: core.SpellMod_DamageDone_Flat,
	// 	// SPELLMOD_ALL_EFFECTS carries the same ladder and also covers Blood Pact, which is
	// 	// buffed elsewhere; this mod is the Firebolt damage half.
	// 	FloatValue: spellData.ImprovedImp.Effect(dbcenums.A_ADD_PCT_MODIFIER, int32(dbcenums.SPELLMOD_DAMAGE)).FractionAt(warlock.Talents.ImprovedImp),
	// 	ClassMask:  WarlockSpellImpFireBolt,
	// })
}

// TODO: To be implemented. Port the TBC Demonic Embrace implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.DemonicEmbrace == 0 {
	// 	return
	// }
	//
	// warlock.MultiplyStat(stats.Stamina, 1.0+(0.03)*float64(warlock.Talents.DemonicEmbrace))
	// warlock.MultiplyStat(stats.Spirit, 1.0-(0.01)*float64(warlock.Talents.DemonicEmbrace))
}

// TODO: To be implemented. Port the TBC Improved Sayaad implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyImprovedSayaad() {
	if warlock.Talents.ImprovedSayaad == 0 || warlock.Options.SacrificeSummon {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.ImprovedSayaad == 0 || warlock.Options.SacrificeSummon {
	// 	return
	// }
	//
	// //This might not actually increase the damage, find a source to prove this
	// warlock.Succubus.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: 0.1 * float64(warlock.Talents.ImprovedSayaad),
	// 	ClassMask:  WarlockSpellSuccubusLashOfPain,
	// })
}

// TODO: To be implemented. Port the TBC Unholy Power implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 || warlock.Options.SacrificeSummon {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.UnholyPower == 0 || warlock.Options.SacrificeSummon {
	// 	return
	// }
	//
	// for _, pet := range warlock.Pets {
	// 	if pet != &warlock.Imp.Pet {
	// 		pet.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexPhysical] *= spellData.UnholyPower.MultiplierAt(warlock.Talents.UnholyPower)
	// 	}
	// }
	//
	// warlock.Imp.AddStaticMod(core.SpellModConfig{
	// 	Kind:       core.SpellMod_DamageDone_Flat,
	// 	FloatValue: spellData.UnholyPower.FractionAt(warlock.Talents.UnholyPower),
	// 	ClassMask:  WarlockSpellImpFireBolt,
	// })
}

// TODO: To be implemented. Port the TBC Demonic Sacrifice implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyDemonicSacrifice() {
	if !warlock.Talents.DemonicSacrifice || warlock.Options.SacrificeSummon == false {
		return
	}

	// The TBC implementation, kept for the port:
	// if !warlock.Talents.DemonicSacrifice || warlock.Options.SacrificeSummon == false {
	// 	return
	// }
	//
	// switch warlock.Options.Summon {
	// case proto.WarlockOptions_Succubus:
	// 	warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexShadow] *= 1.15
	// case proto.WarlockOptions_Imp:
	// 	warlock.PseudoStats.SchoolDamageDealtMultiplier[stats.SchoolIndexFire] *= 1.15
	// case proto.WarlockOptions_Felhunter:
	// 	warlock.applyDemonicSacrificeManaRegen(core.ActionID{SpellID: 18792}, 0.03)
	// }
}

// TODO: To be implemented. Port the TBC Demonic Sacrifice Mana Regen implementation below; not yet verified against the Forever client.
// Demonic Sacrifice restores a percentage of maximum mana every 4 seconds, which is
// independent of the sim's regular 2 second mana ticks.
func (warlock *Warlock) applyDemonicSacrificeManaRegen(actionID core.ActionID, manaPercent float64) {

	// The TBC implementation, kept for the port:
	// manaMetrics := warlock.NewManaMetrics(actionID)
	//
	// warlock.RegisterResetEffect(func(sim *core.Simulation) {
	// 	core.StartPeriodicAction(sim, core.PeriodicActionOptions{
	// 		Period:   time.Second * 4,
	// 		Priority: core.ActionPriorityRegen,
	// 		OnAction: func(sim *core.Simulation) {
	// 			warlock.AddMana(sim, warlock.MaxMana()*manaPercent, manaMetrics)
	// 		},
	// 	})
	// })
}

// TODO: To be implemented. Port the TBC Master Demonologist implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 || warlock.Options.SacrificeSummon == true {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.MasterDemonologist == 0 || warlock.Options.SacrificeSummon == true {
	// 	return
	// }
	// points := float64(warlock.Talents.MasterDemonologist)
	//
	// switch warlock.Options.Summon {
	//
	// case proto.WarlockOptions_Imp:
	// 	warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23825 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
	// 	warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.ThreatMultiplier, 1.0-0.04*points)
	// 	for _, pet := range warlock.Pets {
	// 		if pet == &warlock.Imp.Pet {
	// 			pet.PseudoStats.ThreatMultiplier *= 1.0 - 0.04*points
	// 		}
	// 	}
	// case proto.WarlockOptions_Succubus:
	// 	warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23832 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
	// 	warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.DamageDealtMultiplier, 1.0+0.02*points)
	// 	for _, pet := range warlock.Pets {
	// 		if pet == &warlock.Succubus.Pet {
	// 			pet.PseudoStats.DamageDealtMultiplier *= 1.0 + 0.02*points
	// 		}
	// 	}
	// case proto.WarlockOptions_Voidwalker:
	// 	warlock.PseudoStats.BonusPhysicalDamageTaken *= 1.0 - 0.02*points
	// 	warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23840 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
	// 	warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.BonusPhysicalDamageTaken, 1.0-0.02*points)
	// 	for _, pet := range warlock.Pets {
	// 		if pet == &warlock.Voidwalker.Pet {
	// 			pet.PseudoStats.BonusPhysicalDamageTaken *= 1.0 - 0.02*points
	// 		}
	// 	}
	// case proto.WarlockOptions_Felhunter:
	// 	resistsBonus := 0.20 * points * 70
	// 	warlock.MasterDemonologistAura = warlock.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23836 + int32(points))}, stats.Stats{}, core.NeverExpires).Aura
	// 	warlock.MasterDemonologistAura.AttachStatsBuff(stats.Stats{
	// 		stats.ArcaneResistance: resistsBonus,
	// 		stats.FireResistance:   resistsBonus,
	// 		stats.FrostResistance:  resistsBonus,
	// 		stats.NatureResistance: resistsBonus,
	// 		stats.ShadowResistance: resistsBonus,
	// 	})
	// 	for _, pet := range warlock.Pets {
	// 		if pet == &warlock.Felhunter.Pet {
	// 			pet.NewTemporaryStatsAura("Master Demonologist", core.ActionID{SpellID: (23836 + int32(points))}, stats.Stats{
	// 				stats.ArcaneResistance: resistsBonus,
	// 				stats.FireResistance:   resistsBonus,
	// 				stats.FrostResistance:  resistsBonus,
	// 				stats.NatureResistance: resistsBonus,
	// 				stats.ShadowResistance: resistsBonus,
	// 			}, core.NeverExpires)
	// 		}
	// 	}
	// }
	//
}

// TODO: To be implemented. Port the TBC Soul Link implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink {
		return
	}

	// The TBC implementation, kept for the port:
	// if !warlock.Talents.SoulLink {
	// 	return
	// }
	//
	// // TODO Add if/while pet is alive
	// warlock.PseudoStats.DamageTakenMultiplier *= 0.80
	// warlock.PseudoStats.DamageDealtMultiplier *= 1.05
	//
	// for _, pet := range warlock.Pets {
	// 	pet.PseudoStats.DamageDealtMultiplier *= 1.05
	// }
}

// TODO: To be implemented. Port the TBC Demonic Knowledge implementation below; not yet verified against the Forever client.
func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 {
		return
	}

	// The TBC implementation, kept for the port:
	// if warlock.Talents.DemonicKnowledge == 0 {
	// 	return
	// }
	//
	// warlock.DemonicKnowledgeAura = warlock.RegisterAura(core.Aura{
	// 	Label:    "Demonic Knowledge",
	// 	Duration: core.NeverExpires,
	// })
}

func (warlock *Warlock) updateDemonicKnowledge(sim *core.Simulation) {
	if warlock.DemonicKnowledgeBonus != 0 {
		warlock.AddStatDynamic(sim, stats.SpellDamage, -warlock.DemonicKnowledgeBonus)
	}

	if warlock.ActivePet == nil {
		warlock.DemonicKnowledgeBonus = 0
		return
	}

	coeff := spellData.DemonicKnowledge.FractionAt(warlock.Talents.DemonicKnowledge)
	bonus := coeff * (warlock.ActivePet.GetStat(stats.Stamina) + warlock.ActivePet.GetStat(stats.Intellect))

	warlock.DemonicKnowledgeBonus = bonus
	warlock.AddStatDynamic(sim, stats.SpellDamage, bonus)
}

// applyImprovedHealthFunnel implements Improved Health Funnel, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyImprovedHealthFunnel() {
	if warlock.Talents.ImprovedHealthFunnel == 0 {
		return
	}
}

// applyImprovedVoidwalker implements Improved Voidwalker, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyImprovedVoidwalker() {
	if warlock.Talents.ImprovedVoidwalker == 0 {
		return
	}
}

// applyFelVitality implements Fel Vitality, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyFelVitality() {
	if warlock.Talents.FelVitality == 0 {
		return
	}
}

// applyDemonicEnergies implements Demonic Energies, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyDemonicEnergies() {
	if warlock.Talents.DemonicEnergies == 0 {
		return
	}
}

// applyMasterSummoner implements Master Summoner, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyMasterSummoner() {
	if warlock.Talents.MasterSummoner == 0 {
		return
	}
}

// applyDecimation implements Decimation, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyDecimation() {
	if warlock.Talents.Decimation == 0 {
		return
	}
}

// applyFelDomination implements Fel Domination, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyFelDomination() {
	if !warlock.Talents.FelDomination {
		return
	}
}

// applyDemonicBrand implements Demonic Brand, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyDemonicBrand() {
	if warlock.Talents.DemonicBrand == 0 {
		return
	}
}

// applyImprovedFelhunter implements Improved Felhunter, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyImprovedFelhunter() {
	if warlock.Talents.ImprovedFelhunter == 0 {
		return
	}
}

// applyDemonicPact implements Demonic Pact, new in Forever.
//
// TODO: To be implemented. Needs the Forever tooltip and a spellData ladder before
// the effect can be modelled; there is no TBC equivalent to port.
func (warlock *Warlock) applyDemonicPact() {
	if !warlock.Talents.DemonicPact {
		return
	}
}
