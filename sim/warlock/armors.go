package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Demon Armor (11735), parsed from its row: armor and Shadow resistance. Its in-combat health regen is left
// out by the parse. Demonic Aegis raises every amount it states.
func (warlock *Warlock) registerArmors() {
	rankedSpell := spellData.DemonArmor.ByID(11735)

	aegis := 1.0
	opts := []spelldata.ParseOpt{}
	if warlock.Talents.DemonicAegis > 0 {
		aegis += spellData.DemonicAegis.Rank(warlock.Talents.DemonicAegis).EffectN(1).Percent()
		opts = append(opts, spelldata.Count(aegis))
	}

	warlock.DemonArmor = warlock.RegisterAura(spelldata.AuraConfig(rankedSpell))
	spelldata.ParseEffects(&warlock.Character, warlock.DemonArmor, rankedSpell, opts...)

	if warlock.Options.Armor == proto.WarlockOptions_DemonArmor {
		core.MakePermanent(warlock.DemonArmor)
		warlock.demonArmorArmor = rankedSpell.Effect(dbcenums.A_MOD_RESISTANCE, 1).BaseValue() * aegis
	}
}
