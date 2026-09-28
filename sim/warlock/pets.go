package warlock

import (
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

type WarlockPet struct {
	core.Pet

	owner  *Warlock
	summon proto.WarlockOptions_Summon

	AutoCastAbilities []*core.Spell
	MinMana           float64 // The mana the pet keeps before it casts again
}

// A demon's level 60 stats and what its own stats convert into, from the team's pet model sheet.
type petModel struct {
	baseStats stats.Stats // before the conversions below: Health and Mana are the base pools

	healthPerStamina float64
	manaPerIntellect float64
	mp5PerIntellect  float64
	apPerStrength    float64

	// Melee, for the demons that have it: the weapon's damage range and swing, and a multiplier on
	// everything the demon deals.
	meleeMin, meleeMax float64
	swingSpeed         float64
	damageMod          float64
}

// The sheet gives no crit from Agility or Intellect, and models regen as MP5 and Intellect alone. Its
// totals give a demon no crit or armor of its own: both are the warlock's (see petStatInheritance).
var petModels = map[proto.WarlockOptions_Summon]petModel{
	proto.WarlockOptions_Imp: {
		baseStats: stats.Stats{
			stats.Strength:  122,
			stats.Agility:   35,
			stats.Stamina:   86,
			stats.Intellect: 264,
			stats.Spirit:    260,

			stats.Health: 142.5,
			stats.Mana:   576,
			stats.MP5:    95,
		},
		healthPerStamina: 3.5,
		manaPerIntellect: 5,
		mp5PerIntellect:  0.1,
		apPerStrength:    2,
		damageMod:        1,
	},
	proto.WarlockOptions_Voidwalker: {
		baseStats:        meleeDemonStats(28),
		healthPerStamina: 14,
		manaPerIntellect: 11.5,
		apPerStrength:    2,
		meleeMin:         31,
		meleeMax:         46,
		swingSpeed:       2,
		damageMod:        1,
	},
	proto.WarlockOptions_Succubus: {
		baseStats:        meleeDemonStats(14),
		healthPerStamina: 7,
		manaPerIntellect: 11.5,
		apPerStrength:    2,
		meleeMin:         95,
		meleeMax:         131,
		swingSpeed:       2,
		damageMod:        1.05,
	},
	proto.WarlockOptions_Felhunter: {
		baseStats:        meleeDemonStats(0),
		healthPerStamina: 9.5,
		manaPerIntellect: 11.55,
		mp5PerIntellect:  0.02,
		apPerStrength:    2,
		meleeMin:         70,
		meleeMax:         97,
		swingSpeed:       2,
		damageMod:        0.8,
	},
}

// The Voidwalker, Sayaad and Felhunter share their level 60 stats and base mana; only base health differs.
func meleeDemonStats(baseHealth float64) stats.Stats {
	return stats.Stats{
		stats.Strength:  129,
		stats.Agility:   85,
		stats.Stamina:   234,
		stats.Intellect: 70,
		stats.Spirit:    150,

		stats.Health: baseHealth,
		stats.Mana:   1066,
	}
}

// What a demon takes from the warlock. Warlock Pet Scaling (416189) names the stats it inherits - health,
// attack power, spell damage, resistances, armor, hit, crit, haste, mana regen - but states no amounts,
// which the server supplies, so the shares are hand-kept. The pet model sheet gives Stamina and Intellect
// a share of 0, yet its totals give every demon more health and ~10,870 more mana than its own stats, and
// the same in-combat regen (its own MP5 only out of combat) - all from the warlock, by rules not yet known.
// TODO: The warlock's share of health, mana and in-combat regen; resistances and haste.
func (warlock *Warlock) petStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		// The warlock's spell damage, as its own tooltips read it: the better of its two schools
		spellDamage := ownerStats[stats.SpellDamage] + max(ownerStats[stats.ShadowDamage], ownerStats[stats.FireDamage])

		return stats.Stats{
			stats.Armor:       ownerStats[stats.Armor] * 0.35,
			stats.AttackPower: spellDamage * 0.17,
			stats.SpellDamage: spellDamage * 0.10,

			// TODO: Assumed 1:1, not yet measured.
			stats.PhysicalHitPercent: ownerStats[stats.SpellHitPercent],
			stats.SpellHitPercent:    ownerStats[stats.SpellHitPercent],

			stats.PhysicalCritPercent: ownerStats[stats.SpellCritPercent],
			stats.SpellCritPercent:    ownerStats[stats.SpellCritPercent],
		}
	}
}

func (warlock *Warlock) registerPets() {
	warlock.Imp = warlock.registerPet(proto.WarlockOptions_Imp, "Imp")
	warlock.Voidwalker = warlock.registerPet(proto.WarlockOptions_Voidwalker, "Voidwalker")
	warlock.Succubus = warlock.registerPet(proto.WarlockOptions_Succubus, "Sayaad")
	warlock.Felhunter = warlock.registerPet(proto.WarlockOptions_Felhunter, "Felhunter")
}

func (warlock *Warlock) registerPet(summon proto.WarlockOptions_Summon, name string) *WarlockPet {
	enabledOnStart := warlock.Options.Summon == summon
	model := petModels[summon]

	pet := &WarlockPet{
		owner:  warlock,
		summon: summon,
		Pet: core.NewPet(core.PetConfig{
			Name:            name,
			Owner:           &warlock.Character,
			BaseStats:       model.baseStats,
			StatInheritance: warlock.petStatInheritance(),
			EnabledOnStart:  enabledOnStart,
		}),
	}
	pet.EnableManaBar()

	// Core gives every character 10 health per Stamina and 2 armor per Agility, and the links add up, so
	// only the difference is added here. A pet gets none of the player's Intellect links from EnableManaBar.
	pet.AddStatDependency(stats.Stamina, stats.Health, model.healthPerStamina-10)
	pet.AddStatDependency(stats.Agility, stats.Armor, -2)
	pet.AddStatDependency(stats.Intellect, stats.Mana, model.manaPerIntellect)
	pet.AddStatDependency(stats.Intellect, stats.MP5, model.mp5PerIntellect)
	pet.AddStatDependency(stats.Strength, stats.AttackPower, model.apPerStrength)

	// Regen is MP5 and Intellect alone in the model: Spirit adds none.
	pet.PseudoStats.SpiritRegenMultiplier = 0
	pet.PseudoStats.DamageDealtMultiplier *= model.damageMod

	if model.meleeMax > 0 {
		pet.EnableAutoAttacks(pet, core.AutoAttackOptions{
			MainHand: core.Weapon{
				BaseDamageMin: model.meleeMin,
				BaseDamageMax: model.meleeMax,
				SwingSpeed:    model.swingSpeed,
			},
			AutoSwingMelee: true,
		})
	}

	if enabledOnStart {
		warlock.ActivePet = pet
		warlock.RegisterResetEffect(func(sim *core.Simulation) {
			warlock.ActivePet = pet
		})
	}

	// The warlock's effects that last while a demon is out
	pet.OnPetEnable = func(sim *core.Simulation) {
		for _, aura := range warlock.petActiveAuras {
			aura.Activate(sim)
		}
	}
	pet.OnPetDisable = func(sim *core.Simulation) {
		for _, aura := range warlock.petActiveAuras {
			aura.Deactivate(sim)
		}
	}

	warlock.AddPet(pet)
	return pet
}

// The demons registered, which is none when the warlock sacrifices its demon.
func (warlock *Warlock) demons() []*WarlockPet {
	demons := []*WarlockPet{}
	for _, pet := range []*WarlockPet{warlock.Imp, warlock.Voidwalker, warlock.Succubus, warlock.Felhunter} {
		if pet != nil {
			demons = append(demons, pet)
		}
	}
	return demons
}

// An effect of the warlock's that lasts while a demon is out, gained and lost with it.
func (warlock *Warlock) addPetActiveAura(aura *core.Aura) {
	warlock.petActiveAuras = append(warlock.petActiveAuras, aura)
}

func (pet *WarlockPet) GetPet() *core.Pet {
	return &pet.Pet
}

// Each demon registers its own abilities, as the pet initializes.
func (pet *WarlockPet) Initialize() {
	pet.Pet.Initialize()

	// The Voidwalker's Torment and Suffering are threat, and the Felhunter's abilities utility: both melee only.
	switch pet.summon {
	case proto.WarlockOptions_Imp:
		pet.registerFirebolt()
	case proto.WarlockOptions_Succubus:
		pet.registerLashOfPain()
	}

	if pet.owner.DemonicBrandAuras != nil {
		pet.registerDemonicBrand()
	}
}

func (pet *WarlockPet) Reset(_ *core.Simulation) {
}

func (pet *WarlockPet) OnEncounterStart(_ *core.Simulation) {
}

// Casts the first ability that is ready once the pet has its MinMana; otherwise waits until it has
// the mana for the cheapest one, or briefly when something else holds it back.
func (pet *WarlockPet) ExecuteCustomRotation(sim *core.Simulation) {
	// A demon with nothing to cast only swings, which its auto attacks do on their own.
	if len(pet.AutoCastAbilities) == 0 {
		return
	}

	waitFor := core.NeverExpires

	for _, spell := range pet.AutoCastAbilities {
		if pet.CurrentMana() >= pet.MinMana && spell.CanCast(sim, pet.CurrentTarget) {
			spell.Cast(sim, pet.CurrentTarget)
			return
		}

		needed := max(pet.MinMana, spell.Cost.GetCurrentCost()) - pet.CurrentMana()
		if regen := pet.ManaRegenPerSecondWhileCasting(); needed > 0 && regen > 0 {
			waitFor = min(waitFor, core.DurationFromSeconds(needed/regen))
		}
	}

	// Ready on mana but held by the GCD or a cooldown: look again shortly.
	if waitFor == core.NeverExpires {
		waitFor = time.Millisecond * 100
	}
	pet.WaitUntil(sim, sim.CurrentTime+max(waitFor, time.Millisecond*100))
}

// --- Imp ---

func (pet *WarlockPet) registerFirebolt() {
	rankedSpell := spellData.Firebolt.ByID(11763)
	fireboltHit := rankedSpell.DamageEffect()

	pet.MinMana = rankedSpell.Cost()

	config := spelldata.SpellConfig(&pet.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellImpFireBolt

	// Roll the damage and to hit, then land it after the bolt's travel time
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		result := spell.CalcDamage(sim, target, fireboltHit.Roll(sim, spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(config))
}

// --- Sayaad ---

// The 12s cooldown is the row's category cooldown, which SpellConfig reads.
func (pet *WarlockPet) registerLashOfPain() {
	rankedSpell := spellData.LashOfPain.ByID(11780)
	lashHit := rankedSpell.DamageEffect()

	pet.MinMana = rankedSpell.Cost()

	config := spelldata.SpellConfig(&pet.Unit, rankedSpell, spelldata.Magic(core.ProcMaskSpellDamage))
	config.ClassSpellMask = WarlockSpellSuccubusLashOfPain

	// Roll the damage and to hit
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		spell.CalcAndDealDamage(sim, target, lashHit.Roll(sim, spell.Unit.Level), spell.OutcomeMagicHitAndCrit)
	}

	pet.AutoCastAbilities = append(pet.AutoCastAbilities, pet.RegisterSpell(config))
}

// --- Demonic Brand ---

// Each of the demon's hits on a target the warlock branded consumes a charge and deals the brand's damage:
// Fire from the Imp (1293698), Shadow from the others (1293697). The rows state no amount; 1293695's
// description variables do: ((level - 26) * 1.5 + 14 to 17 + 0.078 * the warlock's better of Shadow and
// Fire spell power), times Master Demonologist's and Unholy Power's bonuses, which the demon's own damage
// multipliers already apply.
// TODO: Imp as Fire and the others as Shadow, every hit as an attack, and the hit and crit roll are assumed;
// the 1 Physical damage row (1293699), a threat marker, is left out.
func (pet *WarlockPet) registerDemonicBrand() {
	warlock := pet.owner

	brandRow := spelldata.MustFind(1293697)
	if pet.summon == proto.WarlockOptions_Imp {
		brandRow = spelldata.MustFind(1293698)
	}
	brandHit := pet.RegisterSpell(spelldata.SpellConfig(&pet.Unit, brandRow, spelldata.Magic(core.ProcMaskSpellDamage), spelldata.Proc()))

	core.MakePermanent(pet.RegisterAura(core.Aura{
		Label: "Demonic Brand - Consume",
		OnSpellHitDealt: func(_ *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if spell == brandHit || !result.Landed() {
				return
			}
			brand := warlock.DemonicBrandAuras.Get(result.Target)
			if !brand.IsActive() {
				return
			}
			brand.RemoveStack(sim)
			if brand.GetStacks() == 0 {
				brand.Deactivate(sim)
			}

			spellPower := warlock.GetStat(stats.SpellDamage) + max(warlock.GetStat(stats.ShadowDamage), warlock.GetStat(stats.FireDamage))
			base := float64(warlock.Level-int32(brandRow.BaseLevel))*1.5 + 0.078*spellPower
			brandHit.CalcAndDealDamage(sim, result.Target, sim.Roll(base+14, base+17), brandHit.OutcomeMagicHitAndCrit)
		},
	}))
}
