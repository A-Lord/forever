package hunter

import (
	"strconv"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var summonHawkRank = spellData.SummonHawk.Highest()
var summonHawkBaseDamage = summonHawkRank.DamageEffect().Average()

// The creature the dive leaves behind, up for 18 s.
var hawkSummon = spellData.SummonHawkTriggered.ByID(1293248)

// The tooltip's ${$s1+($rap*($s4/100))}: effect 4 is the share of ranged attack power.
var summonHawkAPShare = summonHawkRank.EffectN(4).Percent()

// Quick Shots' ranged haste, which a hawk takes as melee speed.
var hawkQuickShotsSpeed = 1 + quickShots.Effect(dbcenums.A_MOD_RANGED_HASTE, 0).Percent()

// A hawk the dive leaves behind: a guardian with a pet's base stats and swing, summoned at the
// enemy by the cast.
type HawkPet struct {
	core.Pet

	atQuickShotsSpeed bool
}

func (hawk *HawkPet) setQuickShots(sim *core.Simulation, active bool) {
	if hawk.atQuickShotsSpeed == active {
		return
	}
	hawk.atQuickShotsSpeed = active
	if active {
		hawk.MultiplyMeleeSpeed(sim, hawkQuickShotsSpeed)
	} else {
		hawk.MultiplyMeleeSpeed(sim, 1/hawkQuickShotsSpeed)
	}
}

func (hawk *HawkPet) GetPet() *core.Pet                   { return &hawk.Pet }
func (hawk *HawkPet) Reset(_ *core.Simulation)            {}
func (hawk *HawkPet) OnEncounterStart(_ *core.Simulation) {}

// A hawk only swings: its rotation sleeps for as long as a hawk can be up.
func (hawk *HawkPet) ExecuteCustomRotation(sim *core.Simulation) {
	hawk.WaitUntil(sim, sim.CurrentTime+hawkSummon.Duration())
}

// Effect 3 is the number of hawks up at once. Hunter Guardian Hawk Pet Scaling 01 (1293586) gives
// each its owner's attack speed, damage done and melee crit chance, all of each; the row is not in
// the store. In game Quick Shots' haste reaches a hawk (hawksFollowQuickShots) and Berserking's does
// not. The crit is the tamed pet's (inheritedMeleeCrit), with none from Agility.
func (hunter *Hunter) newHawks() []*HawkPet {
	if !hunter.Talents.SummonHawk {
		return nil
	}

	hawks := make([]*HawkPet, int(summonHawkRank.EffectN(3).BaseValue()))
	for i := range hawks {
		hawk := &HawkPet{Pet: core.NewPet(core.PetConfig{
			Name:      "Hawk " + strconv.Itoa(i+1),
			Owner:     &hunter.Character,
			BaseStats: petBaseStats,
			StatInheritance: func(ownerStats stats.Stats) stats.Stats {
				return stats.Stats{stats.PhysicalCritPercent: hunter.inheritedMeleeCrit(ownerStats)}
			},
			IsGuardian: true,
			IsDynamic:  true,
		})}
		addPetStatDependencies(&hawk.Character)
		hawk.EnableAutoAttacks(hawk, core.AutoAttackOptions{
			MainHand:       petSwing,
			AutoSwingMelee: true,
		})
		// The damage done is the owner's as the hawk is summoned, on top of the hawk's own.
		inherited := 1.0
		hawk.OnPetEnable = func(sim *core.Simulation) {
			inherited = hunter.PseudoStats.DamageDealtMultiplier
			hawk.PseudoStats.DamageDealtMultiplier *= inherited
		}
		hawk.OnPetDisable = func(sim *core.Simulation) {
			hawk.PseudoStats.DamageDealtMultiplier /= inherited
		}
		hunter.AddPet(hawk)
		hawks[i] = hawk
	}
	return hawks
}

// A hawk swings faster by Quick Shots' haste while it is up, live: from the summon when it is up
// then, back to its own speed when it drops.
func (hunter *Hunter) hawksFollowQuickShots(quickShotsAura *core.Aura) {
	quickShotsAura.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
		for _, hawk := range hunter.Hawks {
			if hawk.IsEnabled() {
				hawk.setQuickShots(sim, true)
			}
		}
	}).ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
		for _, hawk := range hunter.Hawks {
			hawk.setQuickShots(sim, false)
		}
	})

	for _, hawk := range hunter.Hawks {
		hawk.RegisterEnableEffect(func(sim *core.Simulation) {
			hawk.setQuickShots(sim, quickShotsAura.IsActive())
		})
		hawk.RegisterDisableEffect(func(sim *core.Simulation) {
			hawk.setQuickShots(sim, false)
		})
	}
}

// The tamed pet and the hawks: what the talents that name the pet reach.
func (hunter *Hunter) pets() []*core.Pet {
	var pets []*core.Pet
	if hunter.Pet != nil {
		pets = append(pets, &hunter.Pet.Pet)
	}
	for _, hawk := range hunter.Hawks {
		pets = append(pets, &hawk.Pet)
	}
	return pets
}

// The dive's damage, then a hawk for 18 s. The hawks take turns, so one past the count replaces
// the one with the least time left.
func (hunter *Hunter) registerSummonHawk() {
	if !hunter.Talents.SummonHawk {
		return
	}

	config := spelldata.SpellConfig(&hunter.Unit, summonHawkRank, spelldata.Melee(core.ProcMaskRangedSpecial))
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := summonHawkBaseDamage + summonHawkAPShare*spell.RangedAttackPower(target)
		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedCritOnly)
		spell.DealDamageAfterTravel(sim, result)

		hawk := hunter.Hawks[hunter.nextHawk]
		hunter.nextHawk = (hunter.nextHawk + 1) % len(hunter.Hawks)
		if hawk.IsEnabled() {
			hawk.Disable(sim)
		}
		hawk.EnableWithTimeout(sim, hawk, hawkSummon.Duration())
	}

	hunter.RegisterSpell(config)
}
