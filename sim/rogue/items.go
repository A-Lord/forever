package rogue

import (
	"github.com/wowsims/forever/sim/common/shared"
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/spelldata"
	"github.com/wowsims/forever/sim/core/stats"
)

var (
	scraftResist         = spelldata.MustFind(18679)
	scraftAttackPower    = spelldata.MustFind(9336)
	scraftEnergyProc     = spelldata.MustFind(27787)
	scraftEnergyEnergize = spelldata.MustFind(27788)
)

var ItemSetShadowCraft = core.NewItemSet(core.ItemSet{
	Name: "Shadowcraft Armor",
	ID:   184,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.ArcaneResistance, scraftResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.FireResistance, scraftResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.FrostResistance, scraftResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.NatureResistance, scraftResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.ShadowResistance, scraftResist.EffectN(1).BaseValue())
		},
		3: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.AttackPower, scraftAttackPower.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.RangedAttackPower, scraftAttackPower.EffectN(1).BaseValue())
		},
		// 4 Skipped as it provides no sim benefit
		5: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			gain := shared.FiredSpellHandler(&rogue.Character, core.ActionID{SpellID: scraftEnergyEnergize.ID}, scraftEnergyProc)

			trigger := spelldata.ProcTrigger(&rogue.Character, scraftEnergyProc, gain, spelldata.PPM(1.0))
			trigger.Name = "Shadowcraft 5-piece Proc"
			setBonusAura.AttachProcTrigger(trigger)
		},
		// 6 Skipped as it provides no sim benefit
	},
})

var (
	dmantleResist         = spelldata.MustFind(18679)
	dmantleEnergyProc     = spelldata.MustFind(27787)
	dmantleEnergyEnergize = spelldata.MustFind(27788)
	dmantleAttackPower    = spelldata.MustFind(14049)
	dmantleArmor          = spelldata.MustFind(14803)
)

var ItemSetDarkMantle = core.NewItemSet(core.ItemSet{
	Name: "Darkmantle Armor",
	ID:   185,
	Bonuses: map[int32]core.ApplySetBonus{
		2: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.ArcaneResistance, dmantleResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.FireResistance, dmantleResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.FrostResistance, dmantleResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.NatureResistance, dmantleResist.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.ShadowResistance, dmantleResist.EffectN(1).BaseValue())
		},
		4: func(agent core.Agent, setBonusAura *core.Aura) {
			rogue := agent.(RogueAgent).GetRogue()
			gain := shared.FiredSpellHandler(&rogue.Character, core.ActionID{SpellID: dmantleEnergyEnergize.ID}, dmantleEnergyProc)

			trigger := spelldata.ProcTrigger(&rogue.Character, dmantleEnergyProc, gain, spelldata.PPM(1.0))
			trigger.Name = "Darkmantle 4-piece Proc"
			setBonusAura.AttachProcTrigger(trigger)
		},
		6: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.AttackPower, dmantleAttackPower.EffectN(1).BaseValue())
			setBonusAura.AttachStatBuff(stats.RangedAttackPower, dmantleAttackPower.EffectN(1).BaseValue())
		},
		8: func(agent core.Agent, setBonusAura *core.Aura) {
			setBonusAura.AttachStatBuff(stats.Armor, dmantleArmor.EffectN(1).BaseValue())
		},
	},
})
