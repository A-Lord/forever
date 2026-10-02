package shaman

import (
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

// The game files state no chance for Frostbrand Weapon: enchant 1668 is a combat spell enchant with
// 0 in its amount.
//
// TODO: Ingame test needed: the procs per minute.
const frostbrandWeaponPPM = 9.0

// Each imbue registers only where the main hand, equipped or in the swap set, carries it.
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

// Puts the imbue's enchant on the main hand where the options name it, and on the swap set's, and
// answers whether either carries it. Shamans do not dual wield, so the main hand is the only weapon
// an imbue goes on.
func (shaman *Shaman) applyImbue(imbue proto.ShamanImbue, enchantID int32) bool {
	equipped := shaman.HasMHWeapon() && shaman.SelfBuffs.ImbueMH == imbue
	if equipped {
		shaman.MainHand().TempEnchant = enchantID
		if shaman.ItemSwap.IsEnabled() {
			shaman.ItemSwap.AddTempEnchant(enchantID, proto.ItemSlot_ItemSlotMainHand, false)
		}
	}
	swapped := false
	if shaman.ItemSwap.IsEnabled() {
		if mhSwap := shaman.ItemSwap.GetUnequippedItemBySlot(proto.ItemSlot_ItemSlotMainHand); mhSwap.ID != 0 && shaman.SelfBuffs.ImbueMHSwap == imbue {
			mhSwap.TempEnchant = enchantID
			shaman.ItemSwap.AddTempEnchant(enchantID, proto.ItemSlot_ItemSlotMainHand, true)
			swapped = true
		}
	}
	return equipped || swapped
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

// 439431 rolls its 20% on the main hand's melee hits, once per 1.5 s, and fires the rank's 16361: the attack power for its charges and two extra main-hand swings.
func (shaman *Shaman) RegisterWindfuryImbue() {
	if !shaman.applyImbue(proto.ShamanImbue_WindfuryWeapon, windfuryEnchantID) {
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
	if !shaman.applyImbue(proto.ShamanImbue_FlametongueWeapon, flametongueEnchantID) {
		return
	}

	flametongueSpell := shaman.newFlametongueImbueSpell()

	trigger := spelldata.ProcTrigger(&shaman.Character, flametongueWeaponPassive, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
		flametongueSpell.Cast(sim, result.Target)
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

// Enchant 1668 casts 16353 off the main hand's hits. The slow, effect 1, is not
// modelled.
func (shaman *Shaman) RegisterFrostbrandImbue() {
	if !shaman.applyImbue(proto.ShamanImbue_FrostbrandWeapon, frostbrandEnchantID) {
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

// Enchant 1664 puts 16313 on the main hand as an equip aura: the attack power, and the threat Spirit
// Weapons gives effect 2, which the row states as 0.
func (shaman *Shaman) RegisterRockbiterImbue() {
	if !shaman.applyImbue(proto.ShamanImbue_RockbiterWeapon, rockbiterEnchantID) {
		return
	}

	// Elemental Weapons raises effect 1, which the parse would read as damage.
	attackPower := rockbiterWeaponPassive.EffectN(1).Average() * shaman.rockbiterWeaponMultiplier
	threat := 1 + rockbiterWeaponPassive.EffectN(2).Percent() + shaman.spiritWeaponsRockbiterThreat

	config := spelldata.AuraConfig(rockbiterWeaponPassive)
	config.Duration = core.NeverExpires
	config.OnReset = func(aura *core.Aura, sim *core.Simulation) {
		if shaman.MainHand().TempEnchant == rockbiterEnchantID {
			aura.Activate(sim)
		}
	}
	aura := shaman.RegisterAura(config).
		AttachStatBuff(stats.AttackPower, attackPower).
		AttachMultiplicativePseudoStatBuff(&shaman.PseudoStats.ThreatMultiplier, threat)

	shaman.RegisterItemSwapCallback([]proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}, func(sim *core.Simulation, _ proto.ItemSlot) {
		if shaman.MainHand().TempEnchant == rockbiterEnchantID {
			aura.Activate(sim)
		} else {
			aura.Deactivate(sim)
		}
	})
}
