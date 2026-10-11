package buffs

import (
	"github.com/wowsims/forever/sim/core"
)

// A totem stands for the fight, and a true pet stands under every totem its owner does; a guardian
// stands under none. The pet's copy is the same aura built on the pet's own unit, so it bids in the
// same categories there as the owner's copy does on the owner.
func standTotem(char *core.Character, build func(unit *core.Unit, isPlayer bool, talentPoints int32) *core.Aura, talentPoints int32) {
	totem := core.MakePermanent(build(&char.Unit, false, talentPoints))
	inheritTotem(char, totem, func(pet *core.Pet) *core.Aura {
		return build(&pet.Unit, false, talentPoints)
	})
}

// A weapon totem stands for the fight, and its whole chain (totem aura, category bid and proc) stands
// on each true pet that melees.
func inheritWeaponTotem(char *core.Character, stand func(char *core.Character) *core.Aura) {
	inheritTotem(char, core.MakePermanent(stand(char)), func(pet *core.Pet) *core.Aura {
		if !pet.AutoAttacks.AutoSwingMelee {
			return nil
		}
		return stand(&pet.Character)
	})
}

// Each true pet's copy of the owner's totem aura is up while the owner's is and the pet is
// summoned, however often the pet is dismissed and summoned again. A nil copy leaves that pet out.
func inheritTotem(char *core.Character, totem *core.Aura, petCopy func(pet *core.Pet) *core.Aura) {
	for _, pet := range char.Pets {
		if pet.IsGuardian() {
			continue
		}
		aura := petCopy(pet)
		if aura == nil {
			continue
		}
		// The owner's copy decides when the pet's is up, and a pet's stats are never measured
		// through the build phases.
		aura.Duration = core.NeverExpires
		aura.BuildPhase = core.CharacterBuildPhaseNone

		totem.ApplyOnGain(func(_ *core.Aura, sim *core.Simulation) {
			if pet.IsEnabled() {
				aura.Activate(sim)
			}
		})
		totem.ApplyOnExpire(func(_ *core.Aura, sim *core.Simulation) {
			aura.Deactivate(sim)
		})
		pet.RegisterEnableEffect(func(sim *core.Simulation) {
			if totem.IsActive() {
				aura.Activate(sim)
			}
		})
		pet.RegisterDisableEffect(aura.Deactivate)
	}
}
