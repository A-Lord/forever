package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// Scrolls share StatBuffCategory with the raid buffs granting the same stat, so a
// scroll doesn't stack with e.g. Arcane Brilliance or Divine Spirit; only the
// strongest source of that stat applies.
func registerScrollAura(character *Character, label string, itemID int32, stat stats.Stat, amount float64) *Aura {
	aura := character.GetOrRegisterAura(Aura{
		Label:      label,
		ActionID:   ActionID{ItemID: itemID},
		Duration:   NeverExpires,
		BuildPhase: CharacterBuildPhaseConsumes,
	})
	makeExclusiveFlatStatBuff(aura, stat, amount, StatBuffCategory)
	return MakePermanent(aura)
}

// Registers all consume-related effects to the Agent.
func applyConsumeEffects(agent Agent, partyBuffs *proto.PartyBuffs) {
	character := agent.GetCharacter()
	consumables := character.Consumables
	if consumables == nil {
		return
	}

	if consumables.FlaskId != 0 {
		flask := GetConsumableByID(consumables.FlaskId)
		character.AddStats(flask.Stats)
	}

	if consumables.BattleElixirId != 0 {
		// Elixir of Demonslaying
		if consumables.BattleElixirId == 9224 {
			character.Env.RegisterPostFinalizeEffect(func() {
				for _, at := range character.AttackTables {
					at.MobTypeBonusStats[proto.MobType_MobTypeDemon] = at.MobTypeBonusStats[proto.MobType_MobTypeDemon].Add(stats.Stats{
						stats.AttackPower:       265,
						stats.RangedAttackPower: 265,
					})
				}
			})
		} else {
			elixir := GetConsumableByID(consumables.BattleElixirId)
			character.AddStats(elixir.Stats)
		}
	}

	if consumables.GuardianElixirId != 0 {
		// Gift of Arthas
		if consumables.GuardianElixirId == 9088 {
			character.AddStat(stats.ShadowResistance, 10)
			auras := character.NewEnemyAuraArray(func(target *Unit) *Aura {
				return GiftOfArthasAura(target)
			})
			procSpell := character.RegisterSpell(SpellConfig{
				ActionID:    ActionID{SpellID: 11374},
				SpellSchool: SpellSchoolNature,
				ProcMask:    ProcMaskEmpty,

				FlatThreatBonus: 90,

				ApplyEffects: func(sim *Simulation, target *Unit, spell *Spell) {
					spell.CalcAndDealOutcome(sim, target, spell.OutcomeAlwaysHit)
					auras.Get(target).Activate(sim)
				},
			})

			character.MakeProcTriggerAura(ProcTrigger{
				Name:       "Gift of Arthas - Trigger",
				ICD:        time.Second * 3,
				ProcChance: 0.3,
				Outcome:    OutcomeLanded,
				Callback:   CallbackOnSpellHitTaken,
				Handler: func(sim *Simulation, spell *Spell, _ *SpellResult) {
					procSpell.Cast(sim, spell.Unit)
				},
			})
		} else {
			elixir := GetConsumableByID(consumables.GuardianElixirId)
			character.AddStats(elixir.Stats)
		}
	}
	if consumables.FoodId != 0 {
		food := GetConsumableByID(consumables.FoodId)
		character.AddStats(food.Stats)
	}

	// Static Imbues
	if consumables.MhImbueId != 0 && partyBuffs.WindfuryTotem == proto.TristateEffect_TristateEffectMissing {
		registerStaticImbue(agent, consumables.MhImbueId)
	}
	if consumables.OhImbueId != 0 {
		registerStaticImbue(agent, consumables.OhImbueId)
	}

	// Scrolls
	if consumables.ScrollAgi {
		registerScrollAura(character, "Scroll of Agility", 27498, stats.Agility, 20)
	}
	if consumables.ScrollStr {
		registerScrollAura(character, "Scroll of Strength", 27503, stats.Strength, 20)
	}
	if consumables.ScrollInt {
		registerScrollAura(character, "Scroll of Intellect", 27499, stats.Intellect, 20)
	}
	if consumables.ScrollSpi {
		registerScrollAura(character, "Scroll of Spirit", 27501, stats.Spirit, 30)
	}
	if consumables.ScrollArm {
		registerScrollAura(character, "Scroll of Protection", 27500, stats.Armor, 300)
	}

	// Bloodthistle (Blood Elf only): +10 spell damage and healing for 10 min.
	if consumables.Bloodthistle && character.Race == proto.Race_RaceBloodElf {
		character.AddStat(stats.SpellDamage, 10)
		character.AddStat(stats.HealingPower, 10)
	}

	// Bogling Root: +1 physical damage for 10 min (item 5206, spell 5665).
	if consumables.BoglingRoot {
		character.AddStat(stats.PhysicalDamage, 1)
	}

	// Pet Consumes
	for _, pet := range character.Pets {
		if pet.isGuardian {
			continue
		}

		if consumables.PetScrollAgi {
			pet.AddStat(stats.Agility, 20)
		}
		if consumables.PetScrollStr {
			pet.AddStat(stats.Strength, 20)
		}
		if consumables.PetFoodId != 0 {
			petFood := GetConsumableByID(consumables.PetFoodId)
			pet.AddStats(petFood.Stats)
		}
	}

	drumsBombsSharedTimer := character.NewTimer()

	registerPotionCD(agent, consumables)
	registerConjuredCD(agent, consumables)
	registerExplosivesCD(agent, consumables, drumsBombsSharedTimer)
	registerDrumsCD(agent, consumables, drumsBombsSharedTimer)
}
