package forever

import (
	"fmt"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/buffs"
	"github.com/wowsims/forever/sim/core/spelldata"
)

func init() {
	// Graverobber's Shovel
	//
	// Equip 1319047: 12 attack power and 12 ranged attack power, which the item's own stats carry,
	// and effects 3 and 4, 12 of each against Undead.
	core.NewItemEffect(273817, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(1319047)
		aura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(row)))
		spelldata.ParseEffects(character, aura, row, spelldata.Effects(3, 4))
		character.ItemSwap.RegisterProc(273817, aura)
	})

	// Sulfuras, Hand of Ragnaros
	// https://www.wowhead.com/forever/item=17182/sulfuras-hand-of-ragnaros
	//
	// Equip: Deals 5 Fire damage to anyone who strikes you with a melee attack.
	// Chance on hit: Hurls a fiery ball that causes 303 Fire damage and an additional 75 damage over
	// 10s.
	//
	// The game files state no rate for the Fireball; the item's chance-on-hit entry carries it.
	core.NewItemEffect(17182, func(agent core.Agent) {
		character := agent.GetCharacter()

		immolation := spelldata.MustFind(21142)
		shieldAura := core.MakePermanent(character.RegisterAura(spelldata.AuraConfig(immolation, spelldata.Label("Sulfuras, Hand of Ragnaros Immolation"))))
		for _, shield := range spelldata.ItemAuraDamageShields(immolation) {
			buffs.AttachDamageShield(shieldAura, immolation.SpellSchool(), shield.Average(), shield.Coeff())
		}
		character.ItemSwap.RegisterProc(17182, shieldAura)

		row := spelldata.MustFind(21162)
		damage := row.DamageEffect()
		config := procSpellConfig(character, row)
		config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
		config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			result := spell.CalcAndDealDamage(sim, target, damage.DirectAmount(sim), spell.OutcomeMagicHitAndCrit)
			if result.Landed() {
				spell.Dot(target).Apply(sim)
			}
		}
		fireball := character.RegisterSpell(config)

		ppm := weaponProcPPM(17182, row.ID)
		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			fireball.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(17182, ppm, 0)
		})
		trigger.Name = "Sulfuras, Hand of Ragnaros"
		trigger.ActionID = core.ActionID{ItemID: 17182}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(17182, character.MakeProcTriggerAura(trigger))
	})

	// Ravager
	// https://www.wowhead.com/forever/item=7717/ravager
	//
	// Chance on hit: Attack all nearby enemies, dealing weapon damage plus 5 every 3 sec for 9 sec, and
	// reducing your own movement speed by 80%.
	//
	// The game files state no rate for Bladestorm; the item's chance-on-hit entry carries it. The row
	// is a channel, which the sim applies as a 9 s aura on the wearer that leaves the wearer free to
	// act. Each tick casts 9633, a main-hand strike at every enemy.
	core.NewItemEffect(7717, func(agent core.Agent) {
		character := agent.GetCharacter()

		strikeRow := spelldata.MustFind(9633)
		strikeConfig := spelldata.SpellConfig(&character.Unit, strikeRow, spelldata.Melee(core.ProcMaskMeleeMHSpecial), spelldata.Proc())
		if strikeRow.IsAProc() {
			strikeConfig.Flags |= core.SpellFlagProc
		}
		strikeConfig.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			baseDamage := strikeRow.WeaponStrike(sim, spell, spell.Unit.MHWeaponDamage(sim, spell.MeleeAttackPower(target)), 0)
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeMeleeWeaponSpecialHitAndCrit)
			spell.DealBatchedAoeDamage(sim)
		}
		strike := character.RegisterSpell(strikeConfig)

		row := spelldata.MustFind(9632)
		config := procSpellConfig(character, row)
		config.Flags &^= core.SpellFlagChanneled
		config.Dot = spelldata.DotConfig(row, row.EffectN(1), spelldata.Label("Ravager Bladestorm"))
		config.Dot.SelfOnly = true
		config.Dot.OnTick = func(sim *core.Simulation, _ *core.Unit, _ *core.Dot) {
			strike.Cast(sim, character.CurrentTarget)
		}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		}
		bladestorm := character.RegisterSpell(config)

		ppm := weaponProcPPM(7717, row.ID)
		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, result *core.SpellResult) {
			bladestorm.Cast(sim, result.Target)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(7717, ppm, 0)
		})
		trigger.Name = "Ravager"
		trigger.ActionID = core.ActionID{ItemID: 7717}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(7717, character.MakeProcTriggerAura(trigger))
	})

	// Hand of Edward the Odd
	// https://www.wowhead.com/forever/item=2243/hand-of-edward-the-odd
	//
	// Chance on hit: Next spell cast within 4 sec will cast instantly.
	//
	// The game files state no rate for Focus; the item's chance-on-hit entry carries it. The row makes
	// the cast instant with +100000% cast speed, which the sim states as no cast time at all, so the
	// global cooldown stays whole. The next spell with a cast time spends the one charge.
	core.NewItemEffect(2243, func(agent core.Agent) {
		character := agent.GetCharacter()
		row := spelldata.MustFind(18803)

		instant := character.AddDynamicMod(core.SpellModConfig{
			Kind:       core.SpellMod_CastTime_Pct,
			FloatValue: -1,
		})
		config := spelldata.AuraConfig(row, spelldata.Label("Hand of Edward the Odd Focus"))
		config.OnGain = func(_ *core.Aura, _ *core.Simulation) {
			instant.Activate()
		}
		config.OnExpire = func(_ *core.Aura, _ *core.Simulation) {
			instant.Deactivate()
		}
		config.OnCastComplete = func(aura *core.Aura, sim *core.Simulation, spell *core.Spell) {
			if spell.DefaultCast.CastTime > 0 {
				aura.Deactivate(sim)
			}
		}
		focus := character.RegisterAura(config)

		ppm := weaponProcPPM(2243, row.ID)
		trigger := spelldata.ProcTrigger(character, row, func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			focus.Activate(sim)
		}, spelldata.WeaponProc(), func(character *core.Character, trigger *core.ProcTrigger) {
			trigger.ProcChance = 0
			trigger.DPM = character.NewDynamicLegacyProcForWeapon(2243, ppm, 0)
		})
		trigger.Name = "Hand of Edward the Odd"
		trigger.ActionID = core.ActionID{ItemID: 2243}
		trigger.TriggerImmediately = true
		character.ItemSwap.RegisterProc(2243, character.MakeProcTriggerAura(trigger))
	})
}

// The procs per minute the item's chance-on-hit entry for the spell carries.
func weaponProcPPM(itemID int32, spellID int32) float64 {
	for _, effect := range core.GetItemByID(itemID).ItemEffects {
		if effect.BuffId == spellID && effect.GetProc().GetPpm() > 0 {
			return effect.GetProc().GetPpm()
		}
	}
	panic(fmt.Sprintf("item %d states no procs per minute for spell %d", itemID, spellID))
}
