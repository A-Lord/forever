package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

var SuperSapperActionID = ActionID{ItemID: 23827}
var GoblinSapperActionID = ActionID{ItemID: 10646}
var EzThroDynamiteTwoActionID = ActionID{ItemID: 18588}
var CrystalChargeActionID = ActionID{ItemID: 11566}
var FelIronBombActionID = ActionID{ItemID: 23736}
var AdamantiteGrenadeActionID = ActionID{ItemID: 23737}
var GnomishFlameTurretActionID = ActionID{ItemID: 23841}

func registerExplosivesCD(agent Agent, consumes *proto.ConsumesSpec, sharedTimer *Timer) {
	character := agent.GetCharacter()
	if !character.HasProfession(proto.Profession_Engineering) {
		return
	}
	if !consumes.GoblinSapper && !consumes.SuperSapper && consumes.ExplosiveId == 0 {
		return
	}

	if consumes.SuperSapper {
		character.AddMajorCooldown(MajorCooldown{
			Spell:    character.newSuperSapperSpell(sharedTimer),
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 30,
		})
	}
	if consumes.GoblinSapper {
		character.AddMajorCooldown(MajorCooldown{
			Spell:    character.newGoblinSapperSpell(sharedTimer),
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 20,
		})
	}
	if consumes.ExplosiveId > 0 {
		var filler *Spell
		switch consumes.ExplosiveId {
		case 18588:
			filler = character.newEzThroDynamiteTwoSpell(sharedTimer)
		case 15239:
			filler = character.newCrystalChargeSpell(sharedTimer)
		case 30217:
			filler = character.newAdamantiteGrenadeSpell(sharedTimer)
		case 30216:
			filler = character.newFelIronBombSpell(sharedTimer)
		case 30526:
			// Summon Gnomish Turret? Just treat it like a DoT? TBD
		}

		character.AddMajorCooldown(MajorCooldown{
			Spell:    filler,
			Type:     CooldownTypeDPS | CooldownTypeExplosive,
			Priority: CooldownPriorityLow + 10,
		})
	}
}

// Creates a spell object for the common explosive case.
func (character *Character) newBasicExplosiveSpellConfig(sharedTimer *Timer, actionID ActionID, school SpellSchool, minDamage float64, maxDamage float64, speed float64, castTime time.Duration, cooldown Cooldown) SpellConfig {
	var selfDamage *Spell
	if actionID.SameAction(SuperSapperActionID) || actionID.SameAction(GoblinSapperActionID) {
		selfDamage = character.newSapperSelfDamageSpell(actionID, school)
	}

	return SpellConfig{
		ActionID:     actionID,
		SpellSchool:  school,
		DefenseType:  DefenseTypeMagic, // Every explosive's damage spell is Magic in SpellCategories, so they crit for 150%
		ProcMask:     ProcMaskEmpty,
		Flags:        SpellFlagExplosive,
		MissileSpeed: speed,

		Cast: CastConfig{
			DefaultCast: Cast{
				CastTime: castTime,
			},
			CD: cooldown,
			SharedCD: Cooldown{
				Timer:    sharedTimer,
				Duration: time.Minute,
			},
		},

		// Explosives always have 1% resist chance, so just give them hit cap.
		BonusHitPercent:  100,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,

		ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
			baseDamage := sim.Roll(minDamage, maxDamage) * sim.Encounter.AOECapMultiplier()
			spell.CalcAoeDamage(sim, baseDamage, spell.OutcomeMagicHitAndCrit)
			if speed > 0 {
				spell.WaitTravelTime(sim, func(sim *Simulation) {
					spell.DealBatchedAoeDamage(sim)
				})
			} else {
				spell.DealBatchedAoeDamage(sim)
			}

			if selfDamage != nil {
				baseDamage := sim.Roll(minDamage, maxDamage)
				selfDamage.CalcAndDealDamage(sim, &character.Unit, baseDamage, selfDamage.OutcomeMagicHitAndCrit)
			}
		},
	}
}

// The half of a sapper charge that goes off in the thrower's face. Its own spell so that the hit
// carries the kind it is - a harmful spell landing on the character - which is what a listener on
// spell damage taken hears. The tag keeps it apart from the charge's outgoing damage, whose own
// hits state no kind.
func (character *Character) newSapperSelfDamageSpell(actionID ActionID, school SpellSchool) *Spell {
	return character.GetOrRegisterSpell(SpellConfig{
		ActionID:    actionID.WithTag(1),
		SpellSchool: school,
		DefenseType: DefenseTypeMagic,
		ProcMask:    ProcMaskSpellDamage,
		Flags:       SpellFlagExplosive,

		BonusHitPercent:  100,
		DamageMultiplier: 1,
		ThreatMultiplier: 1,
	})
}
func (character *Character) newSuperSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, SuperSapperActionID, SpellSchoolFire, 900, 1500, 0, 0, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newGoblinSapperSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, GoblinSapperActionID, SpellSchoolFire, 450, 750, 0, 0, Cooldown{Timer: character.NewTimer(), Duration: time.Minute * 5}))
}
func (character *Character) newAdamantiteGrenadeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, AdamantiteGrenadeActionID, SpellSchoolFire, 450, 750, 14, time.Second, Cooldown{}))
}
func (character *Character) newFelIronBombSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, FelIronBombActionID, SpellSchoolFire, 330, 770, 14, time.Second, Cooldown{}))
}
func (character *Character) newCrystalChargeSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, CrystalChargeActionID, SpellSchoolFire, 383, 517, 0, 0, Cooldown{}))
}
func (character *Character) newEzThroDynamiteTwoSpell(sharedTimer *Timer) *Spell {
	return character.GetOrRegisterSpell(character.newBasicExplosiveSpellConfig(sharedTimer, EzThroDynamiteTwoActionID, SpellSchoolFire, 213, 287, 14, time.Second, Cooldown{}))
}
