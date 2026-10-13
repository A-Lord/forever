package shaman

import (
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// Windfury Weapon and Flametongue Weapon put the same passive on the weapon at every rank, and the
// rank decides only which spells the passive fires, so they are read off the top rank's refs.
var (
	windfuryWeaponPassive    = spelldata.MustFind(439431)
	windfuryWeaponAttack     = spellData.WindfuryWeapon.Highest().Refs()[0]
	flametongueWeaponPassive = spelldata.MustFind(436519)
	flametongueWeaponProc    = spellData.FlametongueWeapon.Highest().Refs()[0]
	flametongueWeaponAttack  = spellData.FlametongueWeapon.Highest().Refs()[1]
	frostbrandWeaponAttack   = spellData.FrostbrandWeapon.Highest().Refs()[0]
	rockbiterWeaponPassive   = spellData.RockbiterWeapon.Highest().Refs()[0]
)

// Each imbue registers only where the main hand, equipped or in the swap set, carries it. A shaman's
// imbue sits beside an oil or a stone on the same weapon, so it is kept apart from the weapon's
// temporary enchant.
func (shaman *Shaman) registerImbues() {
	shaman.mainHandID = shaman.Equipment[proto.ItemSlot_ItemSlotMainHand].ID
	if shaman.ItemSwap.IsEnabled() {
		shaman.swapMainHandID = shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotMainHand).ID
	}

	shaman.RegisterWindfuryImbue()
	shaman.RegisterFlametongueImbue()
	shaman.RegisterFrostbrandImbue()
	shaman.RegisterRockbiterImbue()
}

// Whether the equipped main hand or the swap set's carries the imbue. Shamans do not dual wield, so
// the main hand is the only weapon an imbue goes on.
func (shaman *Shaman) carriesImbue(imbue proto.ShamanImbue) bool {
	equipped := shaman.mainHandID != 0 && shaman.SelfBuffs.ImbueMH == imbue
	swapped := shaman.swapMainHandID != 0 && shaman.SelfBuffs.ImbueMHSwap == imbue
	return equipped || swapped
}

// The imbue on the main hand in use. The two main hands are told apart by item id, so a swap set
// holding the same weapon keeps the equipped one's imbue.
func (shaman *Shaman) mainHandImbue() proto.ShamanImbue {
	if id := shaman.MainHand().ID; id == shaman.swapMainHandID && id != shaman.mainHandID {
		return shaman.SelfBuffs.ImbueMHSwap
	}
	return shaman.SelfBuffs.ImbueMH
}

// Rolls on the main hand's hits while it carries the imbue, at the weapon's speed for a rate per
// minute, and nothing rolls while it does not.
func (shaman *Shaman) imbueProcManager(imbue proto.ShamanImbue, ppm float64, chance float64) *core.DynamicProcManager {
	build := func() core.DynamicProcManager {
		procMask := core.ProcMaskUnknown
		if shaman.mainHandImbue() == imbue {
			procMask = core.ProcMaskMeleeMH
		}
		if ppm != 0 {
			return *shaman.NewStaticLegacyPPMManager(ppm, procMask)
		}
		return *shaman.NewFixedProcChanceManager(chance, procMask)
	}
	dpm := build()
	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(_ *core.Simulation, _ proto.ItemSlot) {
		dpm = build()
	})
	return &dpm
}

// "When applied to main hand, disables any benefit you personally receive from" the totem of the
// same name: while the main hand carries the imbue, an aura outbids the totem for its category.
func (shaman *Shaman) holdTotemCategoryOnMainHand(label string, actionID core.ActionID, category string, imbue proto.ShamanImbue) {
	aura := shaman.RegisterAura(core.Aura{
		Label:    label,
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if shaman.mainHandImbue() == imbue {
				aura.Activate(sim)
			}
		},
	})
	aura.NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: math.MaxFloat64,
	})

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if shaman.mainHandImbue() == imbue {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	})
}

// The extra attacks are a main-hand weapon special logged as 439440, with 16361's attack power added:
// they roll miss, dodge, parry and crit, never glance, and auto attack procs do not hear them. Like
// every extra attack they report as a main-hand attack tagged with the spell that grants them.
func (shaman *Shaman) newWindfuryImbueSpell() *core.Spell {
	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := windfuryWeaponAttack.EffectN(1).Average() * shaman.windfuryWeaponMultiplier

	return shaman.RegisterSpell(core.SpellConfig{
		ActionID:         core.ActionID{OtherID: proto.OtherAction_OtherActionAttack, Tag: windfuryWeaponAttack.ID},
		SpellSchool:      core.SpellSchoolPhysical,
		DefenseType:      core.DefenseTypeMelee,
		ProcMask:         core.ProcMaskMeleeMHSpecial,
		Flags:            core.SpellFlagMeleeMetrics | core.SpellFlagPassiveSpell | core.SpellFlagNoOnCastComplete,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := shaman.MHWeaponDamage(sim, spell.MeleeAttackPower(target)+attackPower)
			spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
		},
	})
}

// 439431 rolls its 20% on the main hand's melee hits, once per 1.5 s, and fires the rank's 16361,
// whose extra attacks count is how many strikes follow. It rolls off the hit's cast result, and the
// strikes follow once that hit has been dealt, in the same instant.
func (shaman *Shaman) RegisterWindfuryImbue() {
	if !shaman.carriesImbue(proto.ShamanImbue_WindfuryWeapon) {
		return
	}

	strike := shaman.newWindfuryImbueSpell()
	count := int(windfuryWeaponAttack.ExtraAttacksEffect().BaseValue())

	chance := windfuryWeaponPassive.StatedChance()
	trigger := spelldata.ProcTrigger(&shaman.Character, windfuryWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		target := result.Target
		sim.RunAfterCurrentEvent(func(sim *core.Simulation) {
			for range count {
				strike.Cast(sim, target)
			}
		})
	}, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_WindfuryWeapon, 0, chance)
		trigger.ExtraCondition = func(_ *core.Simulation, spell *core.Spell, _ *core.SpellResult) bool {
			return spell != strike
		}
	})
	trigger.Name = "Windfury Weapon Trigger"
	trigger.Callback = core.CallbackOnCastResult
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)

	shaman.holdTotemCategoryOnMainHand("Windfury Weapon (Main Hand)", core.ActionID{SpellID: spellData.WindfuryWeapon.Highest().ID},
		buffs.WindfuryTotemCategory, proto.ShamanImbue_WindfuryWeapon)
}

// The hit is 10444, and the tooltip's ${$16344m1/77-1+($10444m1)} to ${$16344m1/25+($10444m1)}
// over weapon speeds of 1.3 to 4.0 s adds 16344's m1 for every 100 s of swing to its damage.
func (shaman *Shaman) newFlametongueImbueSpell() *core.Spell {
	config := spelldata.SpellConfig(&shaman.Unit, flametongueWeaponAttack,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(),
		spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	config.ClassSpellMask = SpellMaskFlametongueWeapon

	perSwing := flametongueWeaponProc.EffectN(1)
	damage := flametongueWeaponAttack.EffectN(1)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := perSwing.Average()*shaman.AutoAttacks.MH().SwingSpeed/100 + damage.Average()
		spell.CalcAndDealDamage(sim, target, baseDamage, spell.OutcomeMagicHitAndCrit)
	}
	return shaman.RegisterSpell(config)
}

// 436519 has no roll: every melee hit of the main hand adds its fire damage.
func (shaman *Shaman) RegisterFlametongueImbue() {
	if !shaman.carriesImbue(proto.ShamanImbue_FlametongueWeapon) {
		return
	}

	flametongueSpell := shaman.newFlametongueImbueSpell()

	trigger := spelldata.ProcTrigger(&shaman.Character, flametongueWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		flametongueSpell.Cast(sim, result.Target)
	}, func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_FlametongueWeapon, 0, 1)
	})
	trigger.Name = "Flametongue Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)

	shaman.holdTotemCategoryOnMainHand("Flametongue Weapon (Main Hand)", core.ActionID{SpellID: spellData.FlametongueWeapon.Highest().ID},
		buffs.FlametongueTotemCategory, proto.ShamanImbue_FlametongueWeapon)
}

// Enchant 1668 casts 16353 off the main hand's hits. The slow, effect 1, is not
// modelled.
func (shaman *Shaman) RegisterFrostbrandImbue() {
	if !shaman.carriesImbue(proto.ShamanImbue_FrostbrandWeapon) {
		return
	}

	config := spelldata.SpellConfig(&shaman.Unit, frostbrandWeaponAttack,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	config.ClassSpellMask = SpellMaskFrostbrandWeapon
	damage := frostbrandWeaponAttack.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Average(), spell.OutcomeMagicHitAndCrit)
	}
	frostbrandSpell := shaman.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&shaman.Character, frostbrandWeaponAttack, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		frostbrandSpell.Cast(sim, result.Target)
	}, spelldata.WeaponProc(), func(_ *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = shaman.imbueProcManager(proto.ShamanImbue_FrostbrandWeapon, float64(frostbrandWeaponAttack.RPPM), 0)
	})
	trigger.Name = "Frostbrand Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)
}

// Enchant 1664 puts 16313 on the main hand as an equip aura: the attack power, and the threat Spirit
// Weapons gives effect 2, which the row states as 0.
func (shaman *Shaman) RegisterRockbiterImbue() {
	if !shaman.carriesImbue(proto.ShamanImbue_RockbiterWeapon) {
		return
	}

	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := rockbiterWeaponPassive.EffectN(1).Average() * shaman.rockbiterWeaponMultiplier
	threat := 1 + rockbiterWeaponPassive.EffectN(2).Percent() + shaman.spiritWeaponsRockbiterThreat

	config := spelldata.AuraConfig(rockbiterWeaponPassive)
	config.Duration = core.NeverExpires
	config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		if shaman.mainHandImbue() == proto.ShamanImbue_RockbiterWeapon {
			aura.Activate(sim)
		}
	}
	aura := shaman.RegisterAura(config).
		AttachStatBuff(stats.AttackPower, attackPower).
		AttachMultiplicativePseudoStatBuff(&shaman.PseudoStats.ThreatMultiplier, threat)

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if shaman.mainHandImbue() == proto.ShamanImbue_RockbiterWeapon {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	})
}
