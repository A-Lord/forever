package mage

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// One of each conjured gem, best first, one use each per fight. They share the gems' 60 sec
// category cooldown. Each is shown as its item, which the spell data does not link to the spell
// it uses; the pairs are Wowhead's.
var manaGems = []struct {
	itemID int32
	name   string
	rank   *spelldata.Spell
}{
	{8008, "Mana Ruby", spellData.ConjureManaRubyTriggered.Highest()},
	{8007, "Mana Citrine", spellData.ConjureManaCitrineTriggered.Highest()},
	{5513, "Mana Jade", spellData.ConjureManaJadeTriggered.Highest()},
	{5514, "Mana Agate", spellData.ConjureManaAgateTriggered.Highest()},
}

func (mage *Mage) registerManaGems() {
	// The charges of the gems already registered, which are the better ones.
	var betterCharges []*core.Aura
	for _, gem := range manaGems {
		gemRank := gem.rank
		actionID := core.ActionID{ItemID: gem.itemID}
		manaMetrics := mage.NewManaMetrics(actionID)
		manaEffect := gemRank.EnergizeEffect()

		charge := core.MakePermanent(mage.GetOrRegisterAura(core.Aura{
			Label:     gem.name,
			ActionID:  actionID,
			MaxStacks: 1,
		})).ApplyOnReset(func(aura *core.Aura, sim *core.Simulation) {
			aura.SetStacks(sim, 1)
		})

		config := spelldata.SpellConfig(&mage.Unit, gemRank, spelldata.Flags(core.SpellFlagAPL|core.SpellFlagHelpful))
		config.ActionID = actionID
		config.Rank = 0
		config.ClassSpellMask = MageSpellManaGem
		config.Cast.DefaultCast.NonEmpty = true
		config.Cast.CD.Timer = mage.GetConjuredCD()

		config.ExtraCastCondition = func(_ *core.Simulation, _ *core.Unit) bool {
			return charge.GetStacks() > 0
		}

		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			charge.RemoveStack(sim)
			mage.AddMana(sim, manaEffect.Roll(sim, core.CharacterLevel), manaMetrics)
		}

		// Hold this gem until every better one has been used.
		higherGems := betterCharges
		betterCharges = append(betterCharges, charge)

		mage.AddMajorCooldown(core.MajorCooldown{
			Spell: mage.RegisterSpell(config),
			Type:  core.CooldownTypeMana,
			ShouldActivate: func(_ *core.Simulation, character *core.Character) bool {
				for _, higher := range higherGems {
					if higher.GetStacks() > 0 {
						return false
					}
				}
				return character.CurrentMana()+manaEffect.Max(core.CharacterLevel) <= character.MaxMana()
			},
		})
	}
}
