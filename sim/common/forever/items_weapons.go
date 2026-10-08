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
