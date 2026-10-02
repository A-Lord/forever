package shaman

import (
	"fmt"
	"math"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

// The temporary enchants the highest rank of each imbue puts on the weapon, the misc value of the
// cast's effect 1.
const (
	rockbiterEnchantID   int32 = 1664
	flametongueEnchantID int32 = 1666
	frostbrandEnchantID  int32 = 1668
	windfuryEnchantID    int32 = 1669
)

// Windfury Weapon and Flametongue Weapon put the same passive on the weapon at every rank, and the
// rank decides only which spell the passive fires, so the rank's own row is named here.
var (
	windfuryWeaponPassive    = spelldata.MustFind(439431)
	windfuryWeaponAttack     = spellData.WindfuryWeaponTriggered.Highest()
	flametongueWeaponPassive = spelldata.MustFind(436519)
	flametongueWeaponProc    = spellData.FlametongueWeaponTriggered.Highest()
	frostbrandWeaponAttack   = spellData.FrostbrandWeaponTriggered.Highest()
	rockbiterWeaponPassive   = spellData.RockbiterWeaponTriggered.Highest()
)

// The game files state no chance for Frostbrand Weapon: enchant 1668 is a combat spell enchant with
// 0 in its amount.
//
// TODO: Ingame test needed: the procs per minute.
const frostbrandWeaponPPM = 9.0

// Each imbue registers only where a hand, equipped or in the swap set, carries it.
func (shaman *Shaman) registerImbues() {
	shaman.RegisterWindfuryImbue()
	shaman.RegisterFlametongueImbue()
	shaman.RegisterFrostbrandImbue()
	shaman.RegisterRockbiterImbue()
}

func (shaman *Shaman) clearConsumableImbue(slot proto.ItemSlot, imbue proto.ShamanImbue) {
	if imbue == proto.ShamanImbue_NoImbue {
		return
	}
	shaman.Equipment[slot].TempEnchant = 0
	shaman.Equipment[slot].TempEnchantWeaponDamage = 0
}

// Puts the imbue's enchant on each hand the options name it for, and on the swap set's, and answers
// the equipped hands carrying it.
func (shaman *Shaman) applyImbue(imbue proto.ShamanImbue, enchantID int32) core.ProcMask {
	mask := core.ProcMaskUnknown
	if shaman.HasMHWeapon() && shaman.SelfBuffs.ImbueMH == imbue {
		shaman.MainHand().TempEnchant = enchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(enchantID, proto.ItemSlot_ItemSlotMainHand, false)
		}
		mask |= core.ProcMaskMeleeMH
	}
	if shaman.HasOHWeapon() && shaman.SelfBuffs.ImbueOH == imbue {
		shaman.OffHand().TempEnchant = enchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(enchantID, proto.ItemSlot_ItemSlotOffHand, false)
		}
		mask |= core.ProcMaskMeleeOH
	}
	shaman.setupItemSwapImbue(imbue, enchantID)
	return mask
}

func (shaman *Shaman) setupItemSwapImbue(imbue proto.ShamanImbue, imbueID int32) {
	if shaman.ItemSwap.IsEnabled() {
		if mhSwap := shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotMainHand); mhSwap != nil && shaman.SelfBuffs.ImbueMHSwap == imbue {
			mhSwap.TempEnchant = imbueID
			shaman.ItemSwap.AddTempEnchant(imbueID, proto.ItemSlot_ItemSlotMainHand, true)
		}
		if ohSwap := shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotOffHand); ohSwap != nil && shaman.SelfBuffs.ImbueOHSwap == imbue {
			ohSwap.TempEnchant = imbueID
			shaman.ItemSwap.AddTempEnchant(imbueID, proto.ItemSlot_ItemSlotOffHand, true)
		}
	}
}

// Whether the imbue reaches any hand, equipped or in the swap set.
func (shaman *Shaman) imbueIsUsed(mask core.ProcMask) bool {
	return mask != core.ProcMaskUnknown || shaman.ItemSwap.IsEnabled()
}

// "When applied to main hand, disables any benefit you personally receive from" the totem of the
// same name: while the main hand carries the enchant, an aura outbids the totem for its category.
func (shaman *Shaman) holdTotemCategoryOnMainHand(label string, actionID core.ActionID, category string, enchantID int32) {
	aura := shaman.RegisterAura(core.Aura{
		Label:    label,
		ActionID: actionID,
		Duration: core.NeverExpires,
		OnReset: func(aura *core.Aura, sim *core.Simulation) {
			if shaman.MainHand().TempEnchant == enchantID {
				aura.Activate(sim)
			}
		},
	})
	aura.NewExclusiveEffect(category, false, core.ExclusiveEffect{
		Priority: math.MaxFloat64,
	})

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if shaman.MainHand().TempEnchant == enchantID {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	})
}

// 439431 rolls its 20% on the melee hits of the hands carrying the enchant, once per 1.5 s, and
// fires the rank's 16361: the attack power for its charges and two extra main-hand swings.
func (shaman *Shaman) RegisterWindfuryImbue() {
	mask := shaman.applyImbue(proto.ShamanImbue_WindfuryWeapon, windfuryEnchantID)
	if !shaman.imbueIsUsed(mask) {
		return
	}

	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := windfuryWeaponAttack.EffectN(1).Average() * shaman.windfuryWeaponMultiplier
	buff := shaman.RegisterAura(spelldata.AuraConfig(windfuryWeaponAttack)).
		AttachStatBuff(stats.AttackPower, attackPower)

	// The buff's own row hears the auto attacks that spend a charge.
	spender := spelldata.ProcTrigger(&shaman.Character, windfuryWeaponAttack, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
		buff.RemoveStack(sim)
	}, spelldata.Chance(1))
	spender.Name = "Windfury Weapon Charges"
	spender.TriggerImmediately = true
	buff.AttachProcTriggerCallback(&shaman.Unit, spender)

	// A main-hand auto that procs it spends the first charge itself, so the extra attacks are the
	// ones buffed; a special leaves the last charge to the auto after them.
	chance := windfuryWeaponPassive.StatedChance()
	trigger := spelldata.ProcTrigger(&shaman.Character, windfuryWeaponPassive, func(sim *core.Simulation, spell *core.Spell, _ *core.SpellResult) {
		buff.Activate(sim)
		charges := buff.MaxStacks
		if spell.ProcMask.Matches(core.ProcMaskMeleeMHAuto) {
			charges--
		}
		buff.SetStacks(sim, charges)
	}, func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(windfuryEnchantID, 0, func(core.ProcMask) float64 { return chance })
	})
	trigger.Name = "Windfury Weapon Trigger"
	trigger.MetricsActionID = core.ActionID{SpellID: windfuryWeaponAttack.ID}
	onInit := shaman.ExtraAttackProc(&trigger, windfuryWeaponAttack.ID, int(windfuryWeaponAttack.ExtraAttacksEffect().BaseValue()))
	shaman.MakeProcTriggerAura(trigger).ApplyOnInit(onInit)

	shaman.holdTotemCategoryOnMainHand("Windfury Weapon (Main Hand)", core.ActionID{SpellID: spellData.WindfuryWeapon.Highest().ID},
		buffs.WindfuryTotemCategory, windfuryEnchantID)
}

// The fire damage a hit of this hand adds: the tooltip's ${$16344m1/77-1} to ${$16344m1/25} over
// weapon speeds of 1.3 to 4.0 s is m1 for every 100 s of swing.
//
// TODO: Ingame test needed: the spell power share, which 16344 does not state.
func (shaman *Shaman) newFlametongueImbueSpell(isMH bool) *core.Spell {
	config := spelldata.SpellConfig(&shaman.Unit, flametongueWeaponProc,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(),
		spelldata.Tag(core.TernaryInt32(isMH, 1, 2)), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	// 16344 states no defense type, and a fire proc with none rolls on the spell hit and crit tables.
	config.DefenseType = core.DefenseTypeMagic
	config.ClassSpellMask = SpellMaskFlametongueWeapon

	damage := flametongueWeaponProc.EffectN(1)
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		weapon := core.Ternary(isMH, shaman.AutoAttacks.MH(), shaman.AutoAttacks.OH())
		spell.CalcAndDealDamage(sim, target, damage.Average()*weapon.SwingSpeed/100, spell.OutcomeMagicHitAndCrit)
	}
	return shaman.RegisterSpell(config)
}

// 436519 has no roll: every melee hit of a hand carrying the enchant adds that hand's fire damage.
func (shaman *Shaman) RegisterFlametongueImbue() {
	mask := shaman.applyImbue(proto.ShamanImbue_FlametongueWeapon, flametongueEnchantID)
	if !shaman.imbueIsUsed(mask) {
		return
	}

	mhSpell := shaman.newFlametongueImbueSpell(true)
	ohSpell := shaman.newFlametongueImbueSpell(false)

	trigger := spelldata.ProcTrigger(&shaman.Character, flametongueWeaponPassive, func(sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
		core.Ternary(spell.IsMH(), mhSpell, ohSpell).Cast(sim, result.Target)
	}, func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(flametongueEnchantID, 0, func(core.ProcMask) float64 { return 1 })
	})
	trigger.Name = "Flametongue Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)

	shaman.holdTotemCategoryOnMainHand("Flametongue Weapon (Main Hand)", core.ActionID{SpellID: spellData.FlametongueWeapon.Highest().ID},
		buffs.FlametongueTotemCategory, flametongueEnchantID)
}

// Enchant 1668 casts 16353 off the hits of the hands carrying it. The slow, effect 1, is not
// modelled.
func (shaman *Shaman) RegisterFrostbrandImbue() {
	mask := shaman.applyImbue(proto.ShamanImbue_FrostbrandWeapon, frostbrandEnchantID)
	if !shaman.imbueIsUsed(mask) {
		return
	}

	config := spelldata.SpellConfig(&shaman.Unit, frostbrandWeaponAttack,
		spelldata.Magic(core.ProcMaskSpellDamageProc), spelldata.Proc(), spelldata.Flags(core.SpellFlagProc|SpellFlagShamanSpell))
	config.ClassSpellMask = SpellMaskFrostbrandWeapon
	damage := frostbrandWeaponAttack.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, damage.Roll(sim), spell.OutcomeMagicHitAndCrit)
	}
	frostbrandSpell := shaman.RegisterSpell(config)

	trigger := spelldata.ProcTrigger(&shaman.Character, frostbrandWeaponAttack, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		frostbrandSpell.Cast(sim, result.Target)
	}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
		trigger.ProcChance = 0
		trigger.DPM = character.NewDynamicLegacyProcForTempEnchant(frostbrandEnchantID, frostbrandWeaponPPM, func(core.ProcMask) float64 { return 0 })
	})
	trigger.Name = "Frostbrand Weapon Trigger"
	trigger.TriggerImmediately = true
	shaman.MakeProcTriggerAura(trigger)
}

// Enchant 1664 puts 16313 on the weapon as an equip aura, one per hand carrying it: the attack power,
// and the threat Spirit Weapons gives effect 2, which the row states as 0.
//
// TODO: Ingame test needed: whether a Rockbiter Weapon on each hand gives the attack power twice.
func (shaman *Shaman) RegisterRockbiterImbue() {
	mask := shaman.applyImbue(proto.ShamanImbue_RockbiterWeapon, rockbiterEnchantID)
	if !shaman.imbueIsUsed(mask) {
		return
	}

	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := rockbiterWeaponPassive.EffectN(1).Average() * shaman.rockbiterWeaponMultiplier
	threat := 1 + rockbiterWeaponPassive.EffectN(2).Percent() + shaman.spiritWeaponsRockbiterThreat

	for _, slot := range []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand} {
		item := func() *core.Item {
			return core.Ternary(slot == proto.ItemSlot_ItemSlotMainHand, shaman.MainHand(), shaman.OffHand())
		}

		config := spelldata.AuraConfig(rockbiterWeaponPassive, spelldata.Label(fmt.Sprintf("Rockbiter Weapon (%s)", slot)))
		config.Duration = core.NeverExpires
		config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
			if item().TempEnchant == rockbiterEnchantID {
				aura.Activate(sim)
			}
		}
		aura := shaman.RegisterAura(config).
			AttachStatBuff(stats.AttackPower, attackPower).
			AttachMultiplicativePseudoStatBuff(&shaman.PseudoStats.ThreatMultiplier, threat)

		shaman.RegisterItemSwapCallback([]proto.ItemSlot{slot}, func(sim *core.Simulation, _ proto.ItemSlot) {
			if item().TempEnchant == rockbiterEnchantID {
				aura.Activate(sim)
			} else {
				aura.Deactivate(sim)
			}
		})
	}
}
