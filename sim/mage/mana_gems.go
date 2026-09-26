package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The use effect of the conjured Mana Ruby: Replenish Mana (10058), 1000-1200 mana on a
// 60 s category cooldown.
var manaRubyRank = spellData.ConjureManaRubyTriggered.Highest()
var manaRubyEnergize = manaRubyRank.EnergizeEffect()

func (mage *Mage) registerManaGems() {
	actionID := core.ActionID{ItemID: 8008} // Mana Ruby
	manaMetrics := mage.NewManaMetrics(actionID)

	// One conjured ruby per fight.
	manaGemAura := core.MakePermanent(mage.GetOrRegisterAura(core.Aura{
		Label:     "Mana Gem Charges",
		ActionID:  actionID,
		MaxStacks: 1,
	})).ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
		aura.SetStacks(sim, 1)
	})

	config := spelldata.SpellConfig(&mage.Unit, manaRubyRank)
	config.ActionID = actionID
	config.ProcMask = core.ProcMaskEmpty
	config.Flags |= core.SpellFlagAPL | core.SpellFlagHelpful
	config.ClassSpellMask = MageSpellManaGem
	config.Cast.DefaultCast.NonEmpty = true
	config.Cast.CD = core.Cooldown{
		Timer:    mage.GetConjuredCD(),
		Duration: manaRubyRank.CategoryCooldown(),
	}

	config.ExtraCastCondition = func(sim *core.Simulation, target *core.Unit) bool {
		return manaGemAura.GetStacks() > 0
	}

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		manaGemAura.RemoveStack(sim)
		mage.AddMana(sim, manaRubyEnergize.Roll(sim, core.CharacterLevel), manaMetrics)
	}

	manaGem := mage.RegisterSpell(config)

	mage.AddMajorCooldown(core.MajorCooldown{
		Spell: manaGem,
		Type:  core.CooldownTypeMana,
		ShouldActivate: func(sim *core.Simulation, char *core.Character) bool {
			maxGain := manaRubyEnergize.Max(core.CharacterLevel)
			return char.CurrentMana()+maxGain+char.SpiritManaRegenPerSecond() <= char.MaxMana()
		},
	})
}
