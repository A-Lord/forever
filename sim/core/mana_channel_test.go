package core

import (
	"slices"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
)

// A channel of 1 s ticks, paying 100 mana when cost is set.
func manaChannelConfig(ticks int32, cost bool) SpellConfig {
	config := fakeWarlockDot(910300, "Mana Channel", ClassFlags{}, SpellFlagChanneled|SpellFlagAPL)
	config.Dot.NumberOfTicks = ticks
	if cost {
		config.ManaCost = ManaCostOptions{FlatCost: 100}
	}
	return config
}

// The warrior gets a mana bar of 10000 and the 2 s mana ticks; register adds the spell under test.
func newManaTickTest(register func(char *Character) *Spell) *channelTest {
	return newChannelTest(func(ct *channelTest, char *Character) *Spell {
		char.manaBar.unit = &char.Unit
		char.manaCastingMetrics = char.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 1})
		char.manaNotCastingMetrics = char.NewManaMetrics(ActionID{OtherID: proto.OtherAction_OtherActionManaRegen, Tag: 2})
		char.initialStats[stats.Mana] = 10000
		return register(char)
	})
}

// Casts the spell at 0 and runs to until, calling at[t] right after the mana tick at t. Returns the
// mana ticks that gave the regen of a unit outside the five-second rule.
func (ct *channelTest) fullRegenTicks(until time.Duration, at map[time.Duration]func()) []time.Duration {
	ct.start()
	if !ct.spell.Cast(ct.sim, ct.fw.CurrentTarget) {
		panic("the spell under test was not cast")
	}
	var full []time.Duration
	for t := 2 * time.Second; t <= until; t += 2 * time.Second {
		before := ct.fw.manaNotCastingMetrics.Events
		ct.runUntil(t)
		if ct.fw.manaNotCastingMetrics.Events > before {
			full = append(full, t)
		}
		if fn := at[t]; fn != nil {
			fn()
		}
	}
	return full
}

func TestFiveSecondRuleWindows(t *testing.T) {
	for _, tc := range []struct {
		name     string
		register func(char *Character) *Spell
		at       map[time.Duration]func(ct *channelTest)
		want     []float64
	}{
		{
			name:     "a paid 15 s channel holds regen until it ends",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, true)) },
			want:     []float64{16, 18, 20},
		},
		{
			name:     "a paid 5 s channel holds regen until it ends",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(5, true)) },
			want:     []float64{6, 8, 10, 12, 14, 16, 18, 20},
		},
		{
			name:     "a paid channel cut short at 2 s holds regen until 5 s",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, true)) },
			at:       map[time.Duration]func(ct *channelTest){2 * time.Second: func(ct *channelTest) { ct.fw.Interrupt(ct.sim) }},
			want:     []float64{6, 8, 10, 12, 14, 16, 18, 20},
		},
		{
			name:     "a costless channel leaves regen alone",
			register: func(char *Character) *Spell { return char.RegisterSpell(manaChannelConfig(15, false)) },
			want:     []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 20},
		},
		{
			name:     "Shoot leaves regen alone",
			register: func(char *Character) *Spell { return fakeShoot(char, new([]time.Duration)) },
			want:     []float64{2, 4, 6, 8, 10, 12, 14, 16, 18, 20},
		},
		{
			name: "a paid 3 s hardcast holds regen for 5 s from its end",
			register: func(char *Character) *Spell {
				config := fakeWarlockDot(910301, "Hardcast", ClassFlags{}, SpellFlagAPL)
				config.Dot = DotConfig{}
				config.ApplyEffects = func(_ *Simulation, _ *Unit, _ *Spell) {}
				config.ManaCost = ManaCostOptions{FlatCost: 100}
				config.Cast = CastConfig{DefaultCast: Cast{CastTime: 3 * time.Second}}
				return char.RegisterSpell(config)
			},
			want: []float64{2, 8, 10, 12, 14, 16, 18, 20},
		},
		{
			name: "a paid instant holds regen for 5 s",
			register: func(char *Character) *Spell {
				config := fakeWarlockDot(910302, "Instant", ClassFlags{}, SpellFlagAPL)
				config.Dot = DotConfig{}
				config.ApplyEffects = func(_ *Simulation, _ *Unit, _ *Spell) {}
				config.ManaCost = ManaCostOptions{FlatCost: 100}
				return char.RegisterSpell(config)
			},
			want: []float64{6, 8, 10, 12, 14, 16, 18, 20},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ct := newManaTickTest(tc.register)
			at := map[time.Duration]func(){}
			for when, fn := range tc.at {
				at[when] = func() { fn(ct) }
			}
			if got, want := ct.fullRegenTicks(20*time.Second, at), seconds(tc.want...); !slices.Equal(got, want) {
				t.Errorf("full regen on the ticks at %v, want %v", got, want)
			}
		})
	}
}
