package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
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

// --- Tier 1 ---

// TODO: Not implemented, as it does not affect sims for now: Health Funnel is not modelled.
func (warlock *Warlock) applyImprovedHealthFunnel() {
	if warlock.Talents.ImprovedHealthFunnel == 0 {
		return
	}
}

// Effect 2 raises Firebolt, on the Imp. Effect 1 (Fire Shield) and the dummy effect 3 do not affect sims.
func (warlock *Warlock) applyImprovedImp() {
	if warlock.Talents.ImprovedImp == 0 || warlock.Imp == nil {
		return
	}

	spelldata.ParseStatic(&warlock.Imp.Character, spellData.ImprovedImp.Rank(warlock.Talents.ImprovedImp), spelldata.Effects(2))
}

// Effect 1 raises total Stamina, as the tooltip says.
// TODO: Effect 2's -1% Spirit is not in the tooltip and is left out until it is confirmed in game.
func (warlock *Warlock) applyDemonicEmbrace() {
	if warlock.Talents.DemonicEmbrace == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.DemonicEmbrace.Rank(warlock.Talents.DemonicEmbrace), spelldata.Effects(1))
}

// All damage the demons deal. The row's modifier names no spell the sim registers, so it is the pet's own
// damage multiplier; the Firebolt tooltip's $?s18769 reads this talent too.
func (warlock *Warlock) applyUnholyPower() {
	if warlock.Talents.UnholyPower == 0 {
		return
	}

	unholyPowerMult := 1 + spellData.UnholyPower.Rank(warlock.Talents.UnholyPower).EffectN(1).Percent()
	for _, pet := range warlock.demons() {
		pet.PseudoStats.DamageDealtMultiplier *= unholyPowerMult
	}
}

// --- Tier 2 ---

// TODO: Not implemented, as it does not affect sims for now: it raises the Voidwalker's threat and defensive
// abilities, which are not modelled.
func (warlock *Warlock) applyImprovedVoidwalker() {
	if warlock.Talents.ImprovedVoidwalker == 0 {
		return
	}
}

// The demons' maximum health and mana (effect 1), and the warlock's maximum mana (effect 2). The parse skips
// both, so they are applied here.
func (warlock *Warlock) applyFelVitality() {
	if warlock.Talents.FelVitality == 0 {
		return
	}

	rank := spellData.FelVitality.Rank(warlock.Talents.FelVitality)
	petMult, warlockMult := 1+rank.EffectN(1).Percent(), 1+rank.EffectN(2).Percent()

	for _, pet := range warlock.demons() {
		pet.MultiplyStat(stats.Health, petMult)
		pet.MultiplyStat(stats.Mana, petMult)
	}
	warlock.MultiplyStat(stats.Mana, warlockMult)
}

// The demon gains a share of the mana Life Tap returns (lifetap.go). Healing the demon for a share of the
// warlock's spell damage does not affect sims.
func (warlock *Warlock) applyDemonicEnergies() {
	if warlock.Talents.DemonicEnergies == 0 {
		return
	}

	warlock.lifeTapPetShare = spellData.DemonicEnergies.Rank(warlock.Talents.DemonicEnergies).EffectN(2).Percent()
}

// --- Tier 3 ---

// Raises Lash of Pain, on the Sayaad. Effect 2 (Seduction's duration) does not affect sims.
func (warlock *Warlock) applyImprovedSayaad() {
	if warlock.Talents.ImprovedSayaad == 0 || warlock.Succubus == nil {
		return
	}

	spelldata.ParseStatic(&warlock.Succubus.Character, spellData.ImprovedSayaad.Rank(warlock.Talents.ImprovedSayaad))
}

// The sacrificed demon (settled in resolveSacrifice) grants its effect for the fight: the Imp 15% more
// Shadow damage (18789), the Sayaad 15% more Fire damage (18791), the Voidwalker mana every 4 sec (18792).
// The Felhunter's health every 4 sec (18790) does not affect sims.
// TODO: Looking for a way to nicely correlate demons/buffs for frontend display in warlock settings.
func (warlock *Warlock) applyDemonicSacrifice() {
	var effectRow *spelldata.Spell
	switch warlock.Options.SacrificedDemon {
	case proto.WarlockOptions_Imp:
		effectRow = spellData.DemonicSacrificeTriggered.ByID(18789)
	case proto.WarlockOptions_Succubus:
		effectRow = spellData.DemonicSacrificeTriggered.ByID(18791)
	case proto.WarlockOptions_Voidwalker:
		effectRow = spellData.DemonicSacrificeTriggered.ByID(18792)
	default:
		return
	}

	aura := warlock.RegisterAura(spelldata.AuraConfig(effectRow))
	spelldata.ParseEffects(&warlock.Character, aura, effectRow)
	core.MakePermanent(aura)

	// Fel Energy's A_OBS_MOD_POWER: a share of maximum mana every period, which the parse skips.
	if warlock.Options.SacrificedDemon == proto.WarlockOptions_Voidwalker {
		restore := effectRow.EffectN(1)
		manaMetrics := warlock.NewManaMetrics(core.ActionID{SpellID: effectRow.ID})
		warlock.RegisterResetEffect(func(sim *core.Simulation) {
			core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period: restore.Period(),
				OnAction: func(sim *core.Simulation) {
					warlock.AddMana(sim, warlock.MaxMana()*restore.Percent(), manaMetrics)
				},
			})
		})
	}
}

// TODO: Not implemented, as it does not affect sims for now: demons are summoned before the fight.
func (warlock *Warlock) applyMasterSummoner() {
	if warlock.Talents.MasterSummoner == 0 {
		return
	}
}

// --- Tier 4 ---

// Effect 2 shortens Soul Fire's cooldown. Below 35% health (effect 3), Shadow Bolt and Searing Pain deal more
// damage (effect 4), and casting either grants Decimation (440873): Soul Fire casts faster for 10 sec. Its
// Soul Shard saving is not modelled.
func (warlock *Warlock) applyDecimation() {
	if warlock.Talents.Decimation == 0 {
		return
	}

	rank := spellData.Decimation.Rank(warlock.Talents.Decimation)
	spelldata.ParseStatic(&warlock.Character, rank, spelldata.Effects(2))

	// The execute bonus, up from the moment the fight reaches 35%
	executeBonus := rank.EffectN(4)
	executeAura := warlock.RegisterAura(core.Aura{
		Label:    "Decimation - Execute",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: core.NeverExpires,
	})
	spelldata.ParseEffects(&warlock.Character, executeAura, rank, spelldata.Effects(4))

	// TODO: The row states the threshold as effect 3 (35); the sim's execute phases are fixed, and 35 is one.
	warlock.RegisterResetEffect(func(sim *core.Simulation) {
		sim.RegisterExecutePhaseCallback(func(sim *core.Simulation, isExecute int32) {
			if isExecute == 35 {
				executeAura.Activate(sim)
			}
		})
	})

	// The buff's row states the 2/2 value, -40%; the talent's effect 1 prices the rank (-20% at 1/2), so the
	// buff takes the row's duration and Soul Fire's flags, and the rank's value.
	decimationRow := spellData.DecimationTriggered.ByID(440873)
	decimationAura := warlock.RegisterAura(spelldata.AuraConfig(decimationRow)).AttachSpellMod(core.SpellModConfig{
		Kind:       core.SpellMod_CastTime_Pct,
		ClassFlags: decimationRow.EffectN(1).ClassFlags,
		FloatValue: rank.EffectN(1).Percent(),
	})

	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:       "Decimation - Trigger",
		Callback:   core.CallbackOnCastComplete,
		ClassFlags: executeBonus.ClassFlags,
		ExtraCondition: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) bool {
			return sim.IsExecutePhase35()
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			decimationAura.Activate(sim)
		},
	})
}

// TODO: Not implemented, as it does not affect sims for now: demons are summoned before the fight.
func (warlock *Warlock) applyFelDomination() {
	if !warlock.Talents.FelDomination {
		return
	}
}

// Searing Pain brands the target (1293696) for 10 sec; the demon's next attacks against it each consume a
// charge and deal the brand's damage (pets.go). Effect 1, Searing Pain's lower threat, is parsed.
// TODO: The ladder prices effect 2 (the charges) 2/4/6 by rank, but the tooltip reads 3 at 3/3, the row's
// own value, which is taken; confirm in-game.
func (warlock *Warlock) applyDemonicBrand() {
	if warlock.Talents.DemonicBrand == 0 {
		return
	}

	spelldata.ParseStatic(&warlock.Character, spellData.DemonicBrand.Rank(warlock.Talents.DemonicBrand), spelldata.Effects(1))
	charges := int32(spelldata.MustFind(1293695).EffectN(2).BaseValue())

	brandRow := spellData.DemonicBrandTriggered.ByID(1293696)
	warlock.DemonicBrandAuras = warlock.NewEnemyAuraArray(func(target *core.Unit) *core.Aura {
		auraConfig := spelldata.AuraConfig(brandRow)
		auraConfig.Label += " - " + warlock.Label
		auraConfig.MaxStacks = charges
		return target.GetOrRegisterAura(auraConfig)
	})

	warlock.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Demonic Brand - Trigger",
		Callback: core.CallbackOnSpellHitDealt,
		Outcome:  core.OutcomeLanded,
		ExtraCondition: func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spellData.SearingPain.Contains(spell)
		},
		Handler: func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			brand := warlock.DemonicBrandAuras.Get(result.Target)
			brand.Activate(sim)
			brand.SetStacks(sim, charges)
		},
	})
}

// --- Tier 5 ---

// TODO: Not implemented, as it does not affect sims for now: it improves the Felhunter's utility abilities,
// which are not modelled.
func (warlock *Warlock) applyImprovedFelhunter() {
	if warlock.Talents.ImprovedFelhunter == 0 {
		return
	}
}

// While a demon is out, it and the warlock deal more damage (25228, effect 1). The share of damage
// the demon takes instead does not affect sims.
func (warlock *Warlock) applySoulLink() {
	if !warlock.Talents.SoulLink || warlock.ActivePet == nil {
		return
	}

	soulLinkMult := 1 + spellData.SoulLinkTriggered.ByID(25228).EffectN(1).Percent()

	warlock.addPetActiveAura(warlock.RegisterAura(core.Aura{
		Label:    "Soul Link",
		ActionID: core.ActionID{SpellID: spellData.SoulLink.ByID(19028).ID},
		Duration: core.NeverExpires,
	}).AttachMultiplicativePseudoStatBuff(&warlock.PseudoStats.DamageDealtMultiplier, soulLinkMult))

	for _, pet := range warlock.demons() {
		pet.PseudoStats.DamageDealtMultiplier *= soulLinkMult
	}
}

// While a demon is out, the warlock and the demon gain spell damage of up to the rank's share of the
// warlock's level. The row's value is a dummy only the tooltip reads.
func (warlock *Warlock) applyDemonicKnowledge() {
	if warlock.Talents.DemonicKnowledge == 0 || warlock.ActivePet == nil {
		return
	}

	rank := spellData.DemonicKnowledge.Rank(warlock.Talents.DemonicKnowledge)
	spellDamage := float64(warlock.Level) * rank.EffectN(1).Percent()

	warlock.DemonicKnowledgeAura = warlock.RegisterAura(core.Aura{
		Label:    "Demonic Knowledge",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: core.NeverExpires,
	}).AttachStatBuff(stats.SpellDamage, spellDamage)
	warlock.addPetActiveAura(warlock.DemonicKnowledgeAura)

	for _, pet := range warlock.demons() {
		pet.AddStat(stats.SpellDamage, spellDamage)
	}
}

// --- Tier 6 ---

// Grants the warlock and the demon out an effect by demon, its four dummies read by the tooltip: the Imp
// more Fire damage (effect 1), the Voidwalker less Physical damage taken (effect 2), the Sayaad more Shadow
// damage (effect 3), the Felhunter less magic damage taken (effect 4).
func (warlock *Warlock) applyMasterDemonologist() {
	if warlock.Talents.MasterDemonologist == 0 || warlock.ActivePet == nil {
		return
	}

	rank := spellData.MasterDemonologist.Rank(warlock.Talents.MasterDemonologist)

	var schools []stats.SchoolIndex
	var amount float64
	dealt := true
	switch warlock.Options.Summon {
	case proto.WarlockOptions_Imp:
		schools, amount = []stats.SchoolIndex{stats.SchoolIndexFire}, rank.EffectN(1).Percent()
	case proto.WarlockOptions_Voidwalker:
		schools, amount, dealt = []stats.SchoolIndex{stats.SchoolIndexPhysical}, rank.EffectN(2).Percent(), false
	case proto.WarlockOptions_Succubus:
		schools, amount = []stats.SchoolIndex{stats.SchoolIndexShadow}, rank.EffectN(3).Percent()
	case proto.WarlockOptions_Felhunter:
		schools, amount, dealt = []stats.SchoolIndex{stats.SchoolIndexArcane, stats.SchoolIndexFire, stats.SchoolIndexFrost,
			stats.SchoolIndexHoly, stats.SchoolIndexNature, stats.SchoolIndexShadow}, rank.EffectN(4).Percent(), false
	default:
		return
	}

	multipliers := func(unit *core.Unit) []*float64 {
		fields := make([]*float64, len(schools))
		for i, school := range schools {
			if dealt {
				fields[i] = &unit.PseudoStats.SchoolDamageDealtMultiplier[school]
			} else {
				fields[i] = &unit.PseudoStats.SchoolDamageTakenMultiplier[school]
			}
		}
		return fields
	}

	warlock.MasterDemonologistAura = warlock.RegisterAura(core.Aura{
		Label:    "Master Demonologist",
		ActionID: core.ActionID{SpellID: rank.ID},
		Duration: core.NeverExpires,
	})
	for _, field := range multipliers(&warlock.Unit) {
		warlock.MasterDemonologistAura.AttachMultiplicativePseudoStatBuff(field, 1+amount)
	}
	warlock.addPetActiveAura(warlock.MasterDemonologistAura)

	for _, field := range multipliers(&warlock.ActivePet.Unit) {
		*field *= 1 + amount
	}
}

// --- Tier 7 ---

// TODO: Not implemented, as it does not affect sims for now: demons are not resummoned during the fight.
func (warlock *Warlock) applyDemonicPact() {
	if !warlock.Talents.DemonicPact {
		return
	}
}
