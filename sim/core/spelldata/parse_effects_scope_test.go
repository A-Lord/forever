package spelldata

import (
	"math"
	"slices"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/stats"
)

// A player with a sword in the main hand, a dagger in the off hand and a bow, whose talents are
// whatever the test hands to scopeTalents: a weapon-bound mod needs a player in a raid to find the
// weapon behind the attack.
type scopeTester struct {
	core.Character
}

var scopeTalents func(character *core.Character)

func (st *scopeTester) GetCharacter() *core.Character { return &st.Character }
func (st *scopeTester) Initialize()                   {}
func (st *scopeTester) ApplyTalents() {
	if scopeTalents != nil {
		scopeTalents(&st.Character)
	}
}
func (st *scopeTester) Reset(_ *core.Simulation)            {}
func (st *scopeTester) OnGCDReady(_ *core.Simulation)       {}
func (st *scopeTester) OnEncounterStart(_ *core.Simulation) {}

const (
	scopeSwordID int32 = 990951 + iota
	scopeDaggerID
	scopeBowID
)

func init() {
	// The raid buffs register from a package that imports this one, so the tester has none.
	core.RegisterBuffHooks(core.BuffHooks{
		ApplyBuffs:   func(*core.Character, *proto.RaidBuffs, *proto.PartyBuffs, *proto.IndividualBuffs) {},
		ApplyDebuffs: func(*core.Unit, *proto.Debuffs, *proto.Raid) {},
	})

	core.RegisterAgentFactory(
		proto.Player_Rogue{},
		proto.Spec_SpecRogue,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			st := &scopeTester{Character: *character}
			st.EnableAutoAttacks(st, core.AutoAttackOptions{
				MainHand:        st.WeaponFromMainHand(),
				OffHand:         st.WeaponFromOffHand(),
				Ranged:          st.WeaponFromRanged(),
				AutoSwingMelee:  true,
				AutoSwingRanged: true,
			})
			return st
		},
		func(player *proto.Player, spec interface{}) {
			player.Spec = spec.(*proto.Player_Rogue)
		},
	)

	damage := map[int32]*proto.ScalingItemProperties{0: {WeaponDamageMin: 50, WeaponDamageMax: 90}}
	core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{
		{Id: scopeSwordID, Name: "Scope Sword", Type: proto.ItemType_ItemTypeWeapon,
			WeaponType: proto.WeaponType_WeaponTypeSword, HandType: proto.HandType_HandTypeOneHand,
			WeaponSpeed: 2.6, ScalingOptions: damage},
		{Id: scopeDaggerID, Name: "Scope Dagger", Type: proto.ItemType_ItemTypeWeapon,
			WeaponType: proto.WeaponType_WeaponTypeDagger, HandType: proto.HandType_HandTypeOneHand,
			WeaponSpeed: 1.8, ScalingOptions: damage},
		{Id: scopeBowID, Name: "Scope Bow", Type: proto.ItemType_ItemTypeRanged,
			RangedWeaponType: proto.RangedWeaponType_RangedWeaponTypeBow, WeaponSpeed: 3.0, ScalingOptions: damage},
	}})
}

type scopeResult struct {
	sim       *core.Simulation
	character *core.Character
	spell     *core.Spell
	aura      *core.Aura
	parsed    *Parsed
	before    stats.Stats
	after     stats.Stats
}

// Builds the tester, parses row as its talent, and answers what the attacks and stats look like.
// A spell-damage spell is registered beside the autos, for the rows that must not reach spells.
func parseScoped(t *testing.T, row *Spell) *scopeResult {
	t.Helper()
	result := parseScopedWith(t, row, scopeWeapons(scopeSwordID, scopeDaggerID, scopeBowID), nil, false)
	if !result.character.AutoAttacks.IsDualWielding {
		t.Fatal("the sword and dagger should dual wield")
	}
	return result
}

func scopeWeapons(mainHand, offHand, ranged int32) []*proto.ItemSpec {
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotRanged+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}
	items[proto.ItemSlot_ItemSlotMainHand] = &proto.ItemSpec{Id: mainHand}
	items[proto.ItemSlot_ItemSlotOffHand] = &proto.ItemSpec{Id: offHand}
	items[proto.ItemSlot_ItemSlotRanged] = &proto.ItemSpec{Id: ranged}
	return items
}

// parseScoped with the weapons and item swap set the test names. With onAura the row is parsed onto
// an aura the tester registers rather than as a talent.
func parseScopedWith(t *testing.T, row *Spell, equipped, swapped []*proto.ItemSpec, onAura bool) *scopeResult {
	t.Helper()

	result := &scopeResult{}
	scopeTalents = func(character *core.Character) {
		result.spell = character.RegisterSpell(core.SpellConfig{
			ActionID:    core.ActionID{SpellID: 990960},
			ProcMask:    core.ProcMaskSpellDamage,
			SpellSchool: core.SpellSchoolFire,
		})
		result.before = character.GetStats()
		if onAura {
			result.aura = character.RegisterAura(core.Aura{Label: "Scope Aura", ActionID: core.ActionID{SpellID: row.ID},
				Duration: core.NeverExpires})
			result.parsed = ParseEffects(character, result.aura, row)
		} else {
			result.parsed = ParseStatic(character, row)
		}
		result.after = character.GetStats()
	}
	t.Cleanup(func() { scopeTalents = nil })

	player := &proto.Player{
		Name:      "Scope Tester",
		Class:     proto.Class_ClassRogue,
		Race:      proto.Race_RaceUndead,
		Buffs:     &proto.IndividualBuffs{},
		Spec:      &proto.Player_Rogue{},
		Equipment: &proto.EquipmentSpec{Items: equipped},
		Rotation:  &proto.APLRotation{Type: proto.APLRotation_TypeAPL},
	}
	if swapped != nil {
		player.EnableItemSwap = true
		player.ItemSwap = &proto.ItemSwap{Items: swapped}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: []*proto.Player{player}, Buffs: &proto.PartyBuffs{}}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 63, MobType: proto.MobType_MobTypeHumanoid}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	result.sim = sim
	result.character = sim.Raid.Parties[0].Players[0].GetCharacter()
	return result
}

type bonus int

const (
	bonusHit bonus = iota
	bonusCrit
)

func (r *scopeResult) check(t *testing.T, kind bonus, mainHand, offHand, ranged, spell float64) {
	t.Helper()
	for _, c := range []struct {
		name  string
		spell *core.Spell
		want  float64
	}{
		{"main-hand auto", r.character.AutoAttacks.MHAuto(), mainHand},
		{"off-hand auto", r.character.AutoAttacks.OHAuto(), offHand},
		{"ranged auto", r.character.AutoAttacks.RangedAuto(), ranged},
		{"spell", r.spell, spell},
	} {
		got := c.spell.BonusHitPercent
		if kind == bonusCrit {
			got = c.spell.BonusCritPercent
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%s bonus %s = %v, want %v", c.name, map[bonus]string{bonusHit: "hit", bonusCrit: "crit"}[kind], got, c.want)
		}
	}
}

func (r *scopeResult) statChange(stat stats.Stat) float64 {
	return r.after[stat] - r.before[stat]
}

func (r *scopeResult) checkNoStats(t *testing.T, sts ...stats.Stat) {
	t.Helper()
	for _, stat := range sts {
		if got := r.statChange(stat); got != 0 {
			t.Errorf("%s moved by %v, want a scoped mod and no stat", stat.StatName(), got)
		}
	}
}

func casterAura(aura dbcenums.EffectAuraType, points float64) Effect {
	return Effect{Type: dbcenums.E_APPLY_AURA, Aura: aura, BasePoints: points,
		Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_CASTER}}
}

// 22780: +3 hit with a bow, gun or crossbow.
func TestParseScopedRangedHitBonus(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 22780, Name: "Ranged Hit Bonus +3", EquipClass: 2, EquipSubclass: 262156,
		Effects: []Effect{casterAura(dbcenums.A_MOD_HIT_CHANCE, 3)}})

	r.check(t, bonusHit, 0, 0, 3, 0)
	r.checkNoStats(t, stats.PhysicalHitPercent, stats.SpellHitPercent)
}

// 23584: the warrior's off-hand hit, named by ATTR_EX_3_REQ_OFF_HAND on a row of the warrior's family.
func TestParseScopedDualWieldSpecialization(t *testing.T) {
	row := &Spell{ID: 23584, Name: "Dual Wield Specialization", EquipClass: 2, EquipSubclass: 41105,
		ClassFlags: core.ClassFlags{Family: 4},
		Attr:       [17]uint32{0: 0x1d0, 3: dbcenums.ATTR_EX_3_REQ_OFF_HAND, 15: 0x2000},
		Effects: []Effect{
			casterAura(dbcenums.A_MOD_OFFHAND_DAMAGE_PCT, 5),
			casterAura(dbcenums.A_DUMMY, 5),
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 10),
		}}
	r := parseScoped(t, row)

	if kinds := appliedKinds(r.parsed); !slices.Contains(kinds, "SpellMod_BonusHit_Percent") {
		t.Fatalf("attached %v, want a hit mod", kinds)
	}
	r.check(t, bonusHit, 0, 10, 0, 0)
	r.checkNoStats(t, stats.PhysicalHitPercent)
}

// 12320: crit with any melee weapon, which reaches neither the bow nor spells.
func TestParseScopedCruelty(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 12320, Name: "Cruelty", EquipClass: 2, EquipSubclass: 173555,
		Effects: []Effect{casterAura(dbcenums.A_MOD_WEAPON_CRIT_PERCENT, 5)}})

	r.check(t, bonusCrit, 5, 5, 0, 0)
	r.checkNoStats(t, stats.PhysicalCritPercent, stats.SpellCritPercent)
	if got := r.character.ScopedRatingPercent(core.SpellMod_BonusCrit_Percent, core.ProcMaskMeleeMHAuto); got != 0 {
		t.Errorf("a percent talent counts %v percent as rating, want 0", got)
	}
}

// 1306681: hit with a one-hander held in the off hand, by the inventory type as well as the hand bit.
func TestParseScopedOffHandInventoryType(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 1306681, Name: "Increased Off Hand Hit 02", EquipClass: 2,
		EquipSubclass: 41105, EquipInvType: dbcenums.INVTYPE_BIT_WEAPONOFFHAND,
		Attr:    [17]uint32{0: 0x1c0, 3: dbcenums.ATTR_EX_3_REQ_OFF_HAND, 15: 0x2000},
		Effects: []Effect{casterAura(dbcenums.A_MOD_HIT_CHANCE, 2)}})

	r.check(t, bonusHit, 0, 2, 0, 0)
}

// 13705: the row's melee weapons bind its melee hit, and its spell hit is every spell's.
func TestParseScopedPrecision(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 13705, Name: "Precision", EquipClass: 2, EquipSubclass: 173555,
		Effects: []Effect{
			casterAura(dbcenums.A_MOD_HIT_CHANCE, 1),
			casterAura(dbcenums.A_MOD_SPELL_HIT_CHANCE, 1),
		}})

	if kinds := appliedKinds(r.parsed); !slices.Equal(kinds, []string{"SpellMod_BonusHit_Percent", "stat SpellHitPercent"}) {
		t.Errorf("attached %v", kinds)
	}
	r.check(t, bonusHit, 1, 1, 0, 0)
	r.checkNoStats(t, stats.PhysicalHitPercent)
	if got := r.statChange(stats.SpellHitPercent); got != 1 {
		t.Errorf("spell hit moved by %v, want 1", got)
	}
}

// A mask of the ranged bits alone is the rating on ranged attacks, as the percent it is worth there.
func TestParseScopedRangedOnlyRating(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 6100, Name: "Ranged Hit Rating",
		Effects: []Effect{{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_RATING, Misc: 64, BasePoints: 20}}})

	r.check(t, bonusHit, 0, 0, 20/core.HitRatingPerHitPercent, 0)
	r.checkNoStats(t, stats.HitRating, stats.PhysicalHitPercent)
	if got := r.character.ScopedRatingPercent(core.SpellMod_BonusHit_Percent, core.ProcMaskRangedAuto); math.Abs(got-20/core.HitRatingPerHitPercent) > 1e-9 {
		t.Errorf("the ranged rating counts %v percent as rating, want %v", got, 20/core.HitRatingPerHitPercent)
	}
}

// 16255: A_MOD_CRIT_PCT with any melee weapon is physical and spell crit while one is equipped, on
// every spell and attack, and no mod on the weapon's attacks.
func TestParseCritPctWithAMeleeWeaponHoldsWhileEquipped(t *testing.T) {
	row := &Spell{ID: 16255, Name: "Thundering Strikes", EquipClass: 2, EquipSubclass: 173555,
		Effects: []Effect{casterAura(dbcenums.A_MOD_CRIT_PCT, 1)}}
	r := parseScoped(t, row)

	if kinds := appliedKinds(r.parsed); !slices.Equal(kinds, []string{"stat-while-equipped PhysicalCritPercent+SpellCritPercent"}) {
		t.Errorf("attached %v, want the crit stats while equipped", kinds)
	}
	r.check(t, bonusCrit, 0, 0, 0, 0)
	for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
		if got := r.statChange(stat); got != 1 {
			t.Errorf("%s moved by %v, want 1", stat.StatName(), got)
		}
	}
}

// The crit follows the melee weapon through item swaps, as a talent and on an aura alike, and the
// aura's crit holds only while the aura is up.
func TestParseCritPctFollowsTheMeleeWeapon(t *testing.T) {
	row := &Spell{ID: 16255, Name: "Thundering Strikes", EquipClass: 2, EquipSubclass: 173555,
		Effects: []Effect{casterAura(dbcenums.A_MOD_CRIT_PCT, 1)}}
	bowOnly := scopeWeapons(0, 0, scopeBowID)
	melee := scopeWeapons(scopeSwordID, scopeDaggerID, scopeBowID)

	for _, onAura := range []bool{false, true} {
		r := parseScopedWith(t, row, bowOnly, melee, onAura)
		base := r.character.GetStats()
		check := func(when string, want float64) {
			t.Helper()
			for _, stat := range []stats.Stat{stats.PhysicalCritPercent, stats.SpellCritPercent} {
				if got := r.character.GetStat(stat) - base[stat]; math.Abs(got-want) > 1e-9 {
					t.Errorf("on aura %v, %s: %s moved by %v, want %v", onAura, when, stat.StatName(), got, want)
				}
			}
		}

		if r.aura != nil {
			r.aura.Activate(r.sim)
		}
		check("bow alone", 0)

		r.character.ItemSwap.SwapItems(r.sim, proto.APLActionItemSwap_Swap1, false)
		check("sword swapped in", 1)

		if r.aura != nil {
			r.aura.Deactivate(r.sim)
			check("aura down", 0)
			r.aura.Activate(r.sim)
		}

		r.character.ItemSwap.SwapItems(r.sim, proto.APLActionItemSwap_Main, false)
		check("bow alone again", 0)
	}
}

// 1310308: A_MOD_CRIT_PCT with ranged weapons alone is a mod on the ranged attacks.
func TestParseCritPctWithRangedWeaponsIsRangedMod(t *testing.T) {
	r := parseScoped(t, &Spell{ID: 1310308, Name: "Ranged Crit Bonus +02", EquipClass: 2, EquipSubclass: 262156,
		Effects: []Effect{casterAura(dbcenums.A_MOD_CRIT_PCT, 2)}})

	if kinds := appliedKinds(r.parsed); !slices.Equal(kinds, []string{"SpellMod_BonusCrit_Percent"}) {
		t.Errorf("attached %v, want one crit mod", kinds)
	}
	r.check(t, bonusCrit, 0, 0, 2, 0)
	r.checkNoStats(t, stats.PhysicalCritPercent, stats.SpellCritPercent)
}

// Each ranged bit alone: hit and crit become ranged mods at the rating's worth in percent, and haste,
// which no mod carries, attaches nothing.
func TestParseStaticRangedOnlyRatingKinds(t *testing.T) {
	withRows(t, parseRows())

	for _, c := range []struct {
		mask  int32
		kind  string
		value float64
	}{
		{64, "SpellMod_BonusHit_Percent", 10 / core.HitRatingPerHitPercent},
		{512, "SpellMod_BonusCrit_Percent", 10 / core.CritRatingPerCritPercent},
		{262144, "", 0},
		{64 | 512 | 262144, "SpellMod_BonusHit_Percent+SpellMod_BonusCrit_Percent", 10 / core.HitRatingPerHitPercent},
	} {
		parsed := ParseStatic(parseWarrior(), oneEffectRow(dbcenums.A_MOD_RATING, c.mask, 10))
		if c.kind == "" {
			if len(parsed.Applied) != 0 {
				t.Errorf("mask %d attached %v, want nothing", c.mask, appliedKinds(parsed))
			}
			continue
		}
		if len(parsed.Applied) != 1 || parsed.Applied[0].Kind != c.kind ||
			math.Abs(parsed.Applied[0].Value-c.value) > 1e-9 {
			t.Errorf("mask %d attached %v, want %s worth %v", c.mask, parsed.Applied, c.kind, c.value)
		}
	}
}

// A row that names no weapon keeps its channel stat, and so does a row whose weapon is the one the
// caster needs to cast it at an enemy.
func TestParseUnscopedHitKeepsTheStat(t *testing.T) {
	withRows(t, parseRows())

	unscoped := &Spell{ID: 6200, Name: "Hit", Effects: []Effect{casterAura(dbcenums.A_MOD_HIT_CHANCE, 2)}}
	castAtEnemy := &Spell{ID: 3043, Name: "Scorpid Sting", EquipClass: 2, EquipSubclass: 262156,
		Effects: []Effect{{Type: dbcenums.E_APPLY_AURA, Aura: dbcenums.A_MOD_HIT_CHANCE, BasePoints: -2,
			Target: [2]dbcenums.ImplicitTarget{dbcenums.TARGET_UNIT_TARGET_ENEMY}}}}

	for _, row := range []*Spell{unscoped, castAtEnemy} {
		character := parseWarrior()
		before := character.GetStat(stats.PhysicalHitPercent)
		parsed := ParseStatic(character, row)
		if kinds := appliedKinds(parsed); !slices.Equal(kinds, []string{"stat PhysicalHitPercent"}) {
			t.Errorf("%s attached %v, want the physical hit stat", row.Name, kinds)
		}
		if got := character.GetStat(stats.PhysicalHitPercent) - before; got != row.Effects[0].BasePoints {
			t.Errorf("%s moved physical hit by %v, want %v", row.Name, got, row.Effects[0].BasePoints)
		}
	}
}
