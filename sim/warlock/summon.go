package warlock

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// The summons the APL casts, each by its own row: a 10 sec cast for a share of base mana, both of which
// Master Summoner and Fel Domination lower through the class flags the four rows share. The Soul Shard
// the Voidwalker, Succubus and Felhunter take is not modelled.
func (warlock *Warlock) registerSummons() {
	warlock.SummonImp = warlock.registerSummon(688, warlock.Imp)
	warlock.SummonVoidwalker = warlock.registerSummon(697, warlock.Voidwalker)
	warlock.SummonSuccubus = warlock.registerSummon(712, warlock.Succubus)
	warlock.SummonFelhunter = warlock.registerSummon(691, warlock.Felhunter)
}

func (warlock *Warlock) registerSummon(spellID int32, pet *WarlockPet) *core.Spell {
	config := spelldata.SpellConfig(&warlock.Unit, spelldata.MustFind(spellID), spelldata.Flags(core.SpellFlagAPL))
	config.ProcMask = core.ProcMaskEmpty
	config.ClassSpellMask = WarlockSpellSummonDemon

	config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
		warlock.summonDemon(sim, pet)
	}

	return warlock.RegisterSpell(config)
}

// Dismisses the demon out and brings in this one, the same demon included, which returns at full mana. Summoning cancels Demonic Sacrifice's effect: any
// demon without Demonic Pact, and with it only the demon sacrificed.
func (warlock *Warlock) summonDemon(sim *core.Simulation, pet *WarlockPet) {
	if warlock.ActivePet != nil {
		warlock.ActivePet.Disable(sim)
	}
	warlock.ActivePet = pet
	pet.Enable(sim, pet)

	if sacrificed := warlock.sacrificedDemon; sacrificed != proto.WarlockOptions_NoSummon && (!warlock.Talents.DemonicPact || pet.summon == sacrificed) {
		warlock.DemonicSacrificeAuras[sacrificed].Deactivate(sim)
		warlock.sacrificedDemon = proto.WarlockOptions_NoSummon
	}
}
