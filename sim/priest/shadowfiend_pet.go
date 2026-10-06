package priest

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/stats"
)

type Shadowfiend struct {
	core.Pet
	Priest          *Priest
	ManaRestoreAura *core.Aura
}

// TODO: Verify in game. The game files carry no creature stats for the Shadowfiend.
var baseStats = stats.Stats{
	stats.Strength:    153,
	stats.Agility:     108,
	stats.Stamina:     297,
	stats.Intellect:   175,
	stats.Spirit:      122,
	stats.AttackPower: -20, // Negative base offset; Str*2 dependency brings displayed AP to 286
	stats.Armor:       5290,
}

func (priest *Priest) NewShadowfiend() *Shadowfiend {
	shadowfiend := &Shadowfiend{
		Pet: core.NewPet(core.PetConfig{
			Name:            "Shadowfiend",
			Owner:           &priest.Character,
			BaseStats:       baseStats,
			StatInheritance: priest.shadowfiendStatInheritance(),
		}),
		Priest: priest,
	}

	// Mana Leech: each melee swing that lands returns a share of the priest's maximum mana.
	manaLeech := spellData.ShadowfiendTriggered.Highest()
	manaShare := manaLeech.EffectN(1).Percent()
	manaMetric := priest.NewManaMetrics(core.ActionID{SpellID: manaLeech.ID})
	shadowfiend.ManaRestoreAura = shadowfiend.MakeProcTriggerAura(core.ProcTrigger{
		Name:     "Shadowfiend Mana Restore",
		Duration: core.NeverExpires,
		Callback: core.CallbackOnSpellHitDealt,
		ProcMask: core.ProcMaskMeleeWhiteHit,
		Outcome:  core.OutcomeLanded,
		Handler: func(sim *core.Simulation, _ *core.Spell, _ *core.SpellResult) {
			priest.AddMana(sim, priest.MaxMana()*manaShare, manaMetric)
		},
	})

	shadowfiend.EnableAutoAttacks(shadowfiend, core.AutoAttackOptions{
		MainHand: core.Weapon{
			BaseDamageMin:        68, // TODO: Verify in game; the game files carry no creature damage.
			BaseDamageMax:        92, // TODO: Verify in game; the game files carry no creature damage.
			SwingSpeed:           1.5,
			NormalizedSwingSpeed: 1.5,
			SpellSchool:          core.SpellSchoolShadow,
			AttackPowerPerDPS:    core.DefaultAttackPowerPerDPS,
		},
		AutoSwingMelee: true,
	})

	shadowfiend.AutoAttacks.MHConfig().BonusCoefficient = 1.0
	priest.AddPet(shadowfiend)

	shadowfiend.OnPetEnable = func(sim *core.Simulation) {
		shadowfiend.ManaRestoreAura.Activate(sim)
	}

	shadowfiend.OnPetDisable = func(sim *core.Simulation) {
		shadowfiend.ManaRestoreAura.Deactivate(sim)
	}

	return shadowfiend
}

// Shadowfiend Scaling (416153) states every inherited stat without an amount, the server pricing them:
// stamina, attack power, spell damage, the resistances, melee and spell hit, haste and crit.
func (priest *Priest) shadowfiendStatInheritance() core.PetStatInheritance {
	return func(ownerStats stats.Stats) stats.Stats {
		return stats.Stats{
			stats.AttackPower: (ownerStats[stats.SpellDamage] + ownerStats[stats.ShadowDamage]) * 0.57, // TODO: Verify the share in game.
		}
	}
}

func (shadowfiend *Shadowfiend) Initialize() {
	shadowfiend.AddStatDependency(stats.Strength, stats.AttackPower, 2)
}

func (shadowfiend *Shadowfiend) ExecuteCustomRotation(sim *core.Simulation) {
}

func (shadowfiend *Shadowfiend) Reset(sim *core.Simulation) {
	shadowfiend.Disable(sim)
}

func (shadowfiend *Shadowfiend) OnEncounterStart(_ *core.Simulation) {
}

func (shadowfiend *Shadowfiend) GetPet() *core.Pet {
	return &shadowfiend.Pet
}
