package core

import (
	"time"

	"github.com/wowsims/forever/sim/core/proto"
)

func registerDrumsCD(agent Agent, consumables *proto.ConsumesSpec, sharedTimer *Timer) {
	if consumables.DrumsId > 0 && int(consumables.DrumsId) < len(proto.Drums_value) {
		character := agent.GetCharacter()
		config := drumsSpellConfig(character, consumables.DrumsId, false)
		config.Cast = CastConfig{
			DefaultCast: Cast{
				CastTime: TernaryDuration(consumables.DrumsId <= proto.Drums_GreaterDrumsOfWar, 0, time.Second),
				GCD:      GCDDefault,
			},
			CD: Cooldown{
				Timer:    character.NewTimer(),
				Duration: time.Minute * 2,
			},
			SharedCD: Cooldown{
				Timer:    sharedTimer,
				Duration: time.Minute * 2,
			},
		}
		spell := character.RegisterSpell(config)

		character.AddMajorCooldown(MajorCooldown{
			Spell:    spell,
			Type:     CooldownTypeDPS,
			Priority: CooldownPriorityDrums,
		})
	}
}
