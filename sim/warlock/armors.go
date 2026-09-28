package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Demon Armor (11735), parsed from its row: armor and Shadow resistance. Its in-combat health regen is left
// out by the parse. Demonic Aegis raises every amount it states.
func (warlock *Warlock) registerArmors() {
	rankedSpell := spellData.DemonArmor.ByID(11735)

	opts := []spelldata.ParseOpt{}
	if warlock.Talents.DemonicAegis > 0 {
		opts = append(opts, spelldata.Count(1+spellData.DemonicAegis.Rank(warlock.Talents.DemonicAegis).EffectN(1).Percent()))
	}

	warlock.DemonArmor = warlock.RegisterAura(spelldata.AuraConfig(rankedSpell))
	spelldata.ParseEffects(&warlock.Character, warlock.DemonArmor, rankedSpell, opts...)

	switch warlock.Options.Armor {
	case proto.WarlockOptions_DemonArmor:
		core.MakePermanent(warlock.DemonArmor)
	case proto.WarlockOptions_FelArmor:
		// TODO: The game files carry no Fel Armor for Forever's warlock; the option does nothing until one exists.
	}
}
