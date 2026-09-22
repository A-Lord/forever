package database

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/wowsims/forever/sim/core"
)

// The store's rows as the generator builds them: the fields sim/core/spelldata.Spell, .Effect and
// .Power carry, mirrored here rather than imported from that package, with the generator's own
// bookkeeping alongside them and the JSON shape the captured client rows are read back in.
//
// The mirror is not what keeps a generated file that does not compile from stopping the generator
// that rewrites it: this package imports the store for the item-proc routing in gen_effects.go, so
// a store that does not build takes gen_spelldata down with it. What keeps one out of the tree is
// the overlay gate in spelldata_write.go, which writes nothing until the sim compiles against the
// rendered files.
//
// Every value is the client's own column, in the client's units - the conversions live in the
// store's accessors - and the float fields are kept as the DB states them so the emitted number is
// the one the client resolves its tooltip from.
type storeSpell struct {
	ID   int32
	Name string
	Rank string

	School uint8
	Speed  float64
	Attr   [17]uint32

	SpellLevel, BaseLevel, MaxLevel int16

	CastTimeMs, DurationMs int32
	MinRange, MaxRange     float64

	CooldownMs, CategoryCooldownMs, GCDMs int32

	Category, StartRecoveryCategory, ChargeCategory int16

	DefenseType, DispelType, Mechanic, PreventionType uint8

	MaxStack    int16
	ProcChance  uint8
	ProcCharges int16
	ProcFlags   [2]uint32
	ICDMs       int32

	// Procs per minute, which no client row states: it is here only when an override supplies it.
	RPPM float64

	// A flat threat bonus for a spell the client states no E_THREAT effect on, from an override.
	FlatThreat float64

	ClassFlags storeClassFlags

	AuraInterrupt, ChannelInterrupt [2]uint32

	StanceMask uint64

	MaxTargets int16

	EquipClass                  int8
	EquipSubclass, EquipInvType int32

	Labels []int16
	RefIDs []int32

	// Read off Spell.Description_lang by spelldata_hints.go rather than out of a column: the
	// tooltip's wording is what says whether ProcChance is a roll at all, and which effect holds it.
	ProcChanceSource storeProcChanceSource
	ProcChanceEffect int8
	ProcHint         core.ProcHint

	Effects []storeEffect
	Powers  []storePower

	// The generator's own bookkeeping, which the store has no field for. The procs-per-minute id is
	// what a PPM override is checked against, the tooltip flag what a proc chance override is, and
	// the notes are the reasons the emitted row states next to a hand-supplied value.
	procsPerMinuteID    int32
	tooltipStatesChance bool
	overrideNotes       []string
}

// core.ClassFlags: the family a spell belongs to and the four mask words naming it.
type storeClassFlags struct {
	Family int32
	Mask   [4]uint32
}

func (f storeClassFlags) isZero() bool {
	return f.Family == 0 && f.Mask == [4]uint32{}
}

type storeEffect struct {
	ID, SpellID int32
	Index       uint8

	Type int32
	Aura int32

	BasePoints float64
	PPL        float64
	Variance   float64

	// The owning spell's SpellLevels, stamped on while the row is built so that the store's Average
	// does not look its owner up. Never captured: it is derived from the captured Levels row.
	SpellLevel, MaxLevel int16 `json:"-"`

	SPCoef  float64
	APCoef  float64
	PvpMult float64

	Amplitude float64
	PeriodMs  int32

	RadiusMin, RadiusMax float64

	Misc, Misc2 int32

	ClassFlags storeClassFlags

	TriggerID int32

	ChainTargets int16
	ChainAmp     float64

	Mechanic uint8

	PointsPerResource float64

	Target [2]uint8

	Attributes int32

	// The client states no edge from this effect to the spell it fires, and handTriggers does. Set
	// on the effect the hand link was written onto, so the emitted row says where the trigger came
	// from. Never captured: it is decided while the store is rendered, and a capture that carried it
	// would hand the regeneration the answer it is meant to re-derive.
	HandLinked bool `json:"-"`
}

type storePower struct {
	Type               int8
	Cost, CostPerLevel int32
	CostPct            float64
	PerSecond          int32
}

// Every table the store reads, loaded once for the whole client database rather than per spell: the
// closure walks the effects and descriptions of spells it has not selected yet, so the rows have to
// be there before the set of wanted ids is known.
type spellTables struct {
	names        map[int32]string
	subtexts     map[int32]string
	descriptions map[int32]string

	misc         map[int32]miscRow
	levels       map[int32]levelsRow
	cooldowns    map[int32]cooldownRow
	categories   map[int32]categoryRow
	auraOptions  map[int32]auraOptionRow
	classOptions map[int32]storeClassFlags
	interrupts   map[int32]interruptRow
	shapeshift   map[int32]uint64
	targets      map[int32]int16
	equipped     map[int32]equippedRow

	labels  map[int32][]int16
	powers  map[int32][]storePower
	effects map[int32][]storeEffect
}

type miscRow struct {
	Attr                   [17]uint32
	School                 uint8
	Speed                  float64
	CastTimeMs, DurationMs int32
	MinRange, MaxRange     float64
}

type levelsRow struct {
	SpellLevel, BaseLevel, MaxLevel int16
}

type cooldownRow struct {
	CooldownMs, CategoryCooldownMs, GCDMs int32
}

type categoryRow struct {
	Category, StartRecoveryCategory, ChargeCategory   int16
	DefenseType, DispelType, Mechanic, PreventionType uint8
}

type auraOptionRow struct {
	MaxStack         int16
	ProcChance       uint8
	ProcCharges      int16
	ProcFlags        [2]uint32
	ICDMs            int32
	ProcsPerMinuteID int32
}

type interruptRow struct {
	AuraInterrupt, ChannelInterrupt [2]uint32
}

type equippedRow struct {
	Class              int8
	Subclass, InvTypes int32
}

// A spell's rows exist once per difficulty on a few hundred spells - 29213 caps 20 targets at
// difficulty 0 and 10 at 186 - and difficulty 0 is the one the sim plays.
func loadSpellTables(db *sql.DB) (*spellTables, error) {
	t := &spellTables{
		names:        map[int32]string{},
		subtexts:     map[int32]string{},
		descriptions: map[int32]string{},
		misc:         map[int32]miscRow{},
		levels:       map[int32]levelsRow{},
		cooldowns:    map[int32]cooldownRow{},
		categories:   map[int32]categoryRow{},
		auraOptions:  map[int32]auraOptionRow{},
		classOptions: map[int32]storeClassFlags{},
		interrupts:   map[int32]interruptRow{},
		shapeshift:   map[int32]uint64{},
		targets:      map[int32]int16{},
		equipped:     map[int32]equippedRow{},
		labels:       map[int32][]int16{},
		powers:       map[int32][]storePower{},
		effects:      map[int32][]storeEffect{},
	}

	if err := t.loadNames(db); err != nil {
		return nil, err
	}
	if err := t.loadMisc(db); err != nil {
		return nil, err
	}
	if err := t.loadLevels(db); err != nil {
		return nil, err
	}
	if err := t.loadCooldowns(db); err != nil {
		return nil, err
	}
	if err := t.loadCategories(db); err != nil {
		return nil, err
	}
	if err := t.loadAuraOptions(db); err != nil {
		return nil, err
	}
	if err := t.loadClassOptions(db); err != nil {
		return nil, err
	}
	if err := t.loadInterrupts(db); err != nil {
		return nil, err
	}
	if err := t.loadShapeshift(db); err != nil {
		return nil, err
	}
	if err := t.loadTargetRestrictions(db); err != nil {
		return nil, err
	}
	if err := t.loadEquippedItems(db); err != nil {
		return nil, err
	}
	if err := t.loadLabels(db); err != nil {
		return nil, err
	}
	if err := t.loadPowers(db); err != nil {
		return nil, err
	}
	if err := t.loadEffects(db); err != nil {
		return nil, err
	}
	return t, nil
}

// The row the store carries for a spell, from the tables loaded above. The description is read here
// too, for the ids the tooltip names, but never stored: the store is data the sim reads, and the
// tooltip text is generator input.
func (t *spellTables) row(id int32) storeSpell {
	s := storeSpell{ID: id, Name: t.names[id], Rank: t.subtexts[id]}

	m := t.misc[id]
	s.School, s.Speed, s.Attr = m.School, m.Speed, m.Attr
	s.CastTimeMs, s.DurationMs = m.CastTimeMs, m.DurationMs
	s.MinRange, s.MaxRange = m.MinRange, m.MaxRange

	// No SpellLevels row is the client's "no level scaling", which is the spell's own level at the
	// cap and no maximum - the reading levelsOf gives the generated rank tables.
	if l, ok := t.levels[id]; ok {
		s.SpellLevel, s.BaseLevel, s.MaxLevel = l.SpellLevel, l.BaseLevel, l.MaxLevel
	} else {
		s.SpellLevel = RankLevel
	}

	c := t.cooldowns[id]
	s.CooldownMs, s.CategoryCooldownMs, s.GCDMs = c.CooldownMs, c.CategoryCooldownMs, c.GCDMs

	cat := t.categories[id]
	s.Category, s.StartRecoveryCategory, s.ChargeCategory = cat.Category, cat.StartRecoveryCategory, cat.ChargeCategory
	s.DefenseType, s.DispelType = cat.DefenseType, cat.DispelType
	s.Mechanic, s.PreventionType = cat.Mechanic, cat.PreventionType

	a := t.auraOptions[id]
	s.MaxStack, s.ProcChance, s.ProcCharges = a.MaxStack, a.ProcChance, a.ProcCharges
	s.ProcFlags, s.ICDMs = a.ProcFlags, a.ICDMs
	s.procsPerMinuteID = a.ProcsPerMinuteID

	s.ClassFlags = t.classOptions[id]

	i := t.interrupts[id]
	s.AuraInterrupt, s.ChannelInterrupt = i.AuraInterrupt, i.ChannelInterrupt

	s.StanceMask = t.shapeshift[id]
	s.MaxTargets = t.targets[id]

	e := t.equipped[id]
	s.EquipClass, s.EquipSubclass, s.EquipInvType = e.Class, e.Subclass, e.InvTypes

	s.Labels = t.labels[id]
	s.RefIDs = t.referencedIDs(id)
	s.Powers = t.powers[id]

	// The effect's class mask is read against the owning spell's family: the mask words alone name
	// nothing, since the same bit is a different spell in each family. The family is carried only
	// where the effect states a mask, because Matches needs an overlapping mask word as well and a
	// family on its own can never produce one.
	//
	// The spell's levels go onto every effect for the same reason the family does: the amount an
	// effect scales to is priced from them, and the effect is what the caller holds.
	s.Effects = make([]storeEffect, len(t.effects[id]))
	copy(s.Effects, t.effects[id])
	for i := range s.Effects {
		if !s.Effects[i].ClassFlags.isZero() {
			s.Effects[i].ClassFlags.Family = s.ClassFlags.Family
		}
		s.Effects[i].SpellLevel, s.Effects[i].MaxLevel = s.SpellLevel, s.MaxLevel
	}

	return s
}

func (t *spellTables) loadNames(db *sql.DB) error {
	if err := eachRow(db, `SELECT ID, Name_lang FROM SpellName ORDER BY ID`, func(rows *sql.Rows) error {
		var id int32
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		t.names[id] = name
		return nil
	}); err != nil {
		return err
	}

	return eachRow(db, `
		SELECT ID, COALESCE(NameSubtext_lang, ''), COALESCE(Description_lang, '')
		FROM Spell ORDER BY ID`, func(rows *sql.Rows) error {
		var id int32
		var subtext, description string
		if err := rows.Scan(&id, &subtext, &description); err != nil {
			return err
		}
		t.subtexts[id] = subtext
		t.descriptions[id] = description
		return nil
	})
}

func (t *spellTables) loadMisc(db *sql.DB) error {
	return eachRow(db, `
		SELECT m.SpellID,
		       m.Attributes_0, m.Attributes_1, m.Attributes_2, m.Attributes_3, m.Attributes_4,
		       m.Attributes_5, m.Attributes_6, m.Attributes_7, m.Attributes_8, m.Attributes_9,
		       m.Attributes_10, m.Attributes_11, m.Attributes_12, m.Attributes_13, m.Attributes_14,
		       m.Attributes_15, m.Attributes_16,
		       COALESCE(m.SchoolMask, 0), COALESCE(m.Speed, 0),
		       COALESCE(ct.Base, 0), COALESCE(d.Duration, 0),
		       COALESCE(r.RangeMin_0, 0), COALESCE(r.RangeMax_0, 0)
		FROM SpellMisc m
		LEFT JOIN SpellCastTimes ct ON ct.ID = m.CastingTimeIndex
		LEFT JOIN SpellDuration d ON d.ID = m.DurationIndex
		LEFT JOIN SpellRange r ON r.ID = m.RangeIndex
		WHERE m.DifficultyID = 0
		ORDER BY m.SpellID`, func(rows *sql.Rows) error {
		var id int32
		var attr [17]int64
		var m miscRow
		var school int64
		dest := []any{&id}
		for i := range attr {
			dest = append(dest, &attr[i])
		}
		dest = append(dest, &school, &m.Speed, &m.CastTimeMs, &m.DurationMs, &m.MinRange, &m.MaxRange)
		if err := rows.Scan(dest...); err != nil {
			return err
		}
		for i, word := range attr {
			m.Attr[i] = uint32(word)
		}
		m.School = uint8(school)
		return t.putMisc(id, m)
	})
}

func (t *spellTables) putMisc(id int32, m miscRow) error {
	if _, dup := t.misc[id]; dup {
		return fmt.Errorf("spell %d has two SpellMisc rows at difficulty 0", id)
	}
	t.misc[id] = m
	return nil
}

func (t *spellTables) loadLevels(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, SpellLevel, BaseLevel, MaxLevel
		FROM SpellLevels WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var l levelsRow
		if err := rows.Scan(&id, &l.SpellLevel, &l.BaseLevel, &l.MaxLevel); err != nil {
			return err
		}
		if _, dup := t.levels[id]; dup {
			return fmt.Errorf("spell %d has two SpellLevels rows at difficulty 0", id)
		}
		t.levels[id] = l
		return nil
	})
}

func (t *spellTables) loadCooldowns(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(RecoveryTime, 0), COALESCE(CategoryRecoveryTime, 0),
		       COALESCE(StartRecoveryTime, 0)
		FROM SpellCooldowns WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var c cooldownRow
		if err := rows.Scan(&id, &c.CooldownMs, &c.CategoryCooldownMs, &c.GCDMs); err != nil {
			return err
		}
		if _, dup := t.cooldowns[id]; dup {
			return fmt.Errorf("spell %d has two SpellCooldowns rows at difficulty 0", id)
		}
		t.cooldowns[id] = c
		return nil
	})
}

func (t *spellTables) loadCategories(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(Category, 0), COALESCE(StartRecoveryCategory, 0),
		       COALESCE(ChargeCategory, 0), COALESCE(DefenseType, 0), COALESCE(DispelType, 0),
		       COALESCE(Mechanic, 0), COALESCE(PreventionType, 0)
		FROM SpellCategories WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var c categoryRow
		var defense, dispel, mechanic, prevention int64
		if err := rows.Scan(&id, &c.Category, &c.StartRecoveryCategory, &c.ChargeCategory,
			&defense, &dispel, &mechanic, &prevention); err != nil {
			return err
		}
		c.DefenseType, c.DispelType = uint8(defense), uint8(dispel)
		c.Mechanic, c.PreventionType = uint8(mechanic), uint8(prevention)
		if _, dup := t.categories[id]; dup {
			return fmt.Errorf("spell %d has two SpellCategories rows at difficulty 0", id)
		}
		t.categories[id] = c
		return nil
	})
}

func (t *spellTables) loadAuraOptions(db *sql.DB) error {
	// ProcTypeMask_0 and _1 are generated columns over the JSON array, so they are named here rather
	// than reached through SELECT *.
	return eachRow(db, `
		SELECT SpellID, COALESCE(CumulativeAura, 0), COALESCE(ProcChance, 0), COALESCE(ProcCharges, 0),
		       COALESCE(ProcTypeMask_0, 0), COALESCE(ProcTypeMask_1, 0), COALESCE(ProcCategoryRecovery, 0),
		       COALESCE(SpellProcsPerMinuteID, 0)
		FROM SpellAuraOptions WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var a auraOptionRow
		var chance, mask0, mask1 int64
		if err := rows.Scan(&id, &a.MaxStack, &chance, &a.ProcCharges, &mask0, &mask1, &a.ICDMs,
			&a.ProcsPerMinuteID); err != nil {
			return err
		}
		a.ProcChance = uint8(chance)
		a.ProcFlags = [2]uint32{uint32(mask0), uint32(mask1)}
		if _, dup := t.auraOptions[id]; dup {
			return fmt.Errorf("spell %d has two SpellAuraOptions rows at difficulty 0", id)
		}
		t.auraOptions[id] = a
		return nil
	})
}

func (t *spellTables) loadClassOptions(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(SpellClassSet, 0), COALESCE(SpellClassMask_0, 0),
		       COALESCE(SpellClassMask_1, 0), COALESCE(SpellClassMask_2, 0), COALESCE(SpellClassMask_3, 0)
		FROM SpellClassOptions ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var f storeClassFlags
		var mask [4]int64
		if err := rows.Scan(&id, &f.Family, &mask[0], &mask[1], &mask[2], &mask[3]); err != nil {
			return err
		}
		for i, word := range mask {
			f.Mask[i] = uint32(word)
		}
		if _, dup := t.classOptions[id]; dup {
			return fmt.Errorf("spell %d has two SpellClassOptions rows", id)
		}
		t.classOptions[id] = f
		return nil
	})
}

func (t *spellTables) loadInterrupts(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(AuraInterruptFlags_0, 0), COALESCE(AuraInterruptFlags_1, 0),
		       COALESCE(ChannelInterruptFlags_0, 0), COALESCE(ChannelInterruptFlags_1, 0)
		FROM SpellInterrupts WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var aura0, aura1, channel0, channel1 int64
		if err := rows.Scan(&id, &aura0, &aura1, &channel0, &channel1); err != nil {
			return err
		}
		if _, dup := t.interrupts[id]; dup {
			return fmt.Errorf("spell %d has two SpellInterrupts rows at difficulty 0", id)
		}
		t.interrupts[id] = interruptRow{
			AuraInterrupt:    [2]uint32{uint32(aura0), uint32(aura1)},
			ChannelInterrupt: [2]uint32{uint32(channel0), uint32(channel1)},
		}
		return nil
	})
}

// The two mask words are one 64-bit set of forms, which is how core reads a stance mask.
func (t *spellTables) loadShapeshift(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(ShapeshiftMask_0, 0), COALESCE(ShapeshiftMask_1, 0)
		FROM SpellShapeshift ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var low, high int64
		if err := rows.Scan(&id, &low, &high); err != nil {
			return err
		}
		if _, dup := t.shapeshift[id]; dup {
			return fmt.Errorf("spell %d has two SpellShapeshift rows", id)
		}
		t.shapeshift[id] = uint64(uint32(low)) | uint64(uint32(high))<<32
		return nil
	})
}

func (t *spellTables) loadTargetRestrictions(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(MaxTargets, 0)
		FROM SpellTargetRestrictions WHERE DifficultyID = 0 ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var maxTargets int16
		if err := rows.Scan(&id, &maxTargets); err != nil {
			return err
		}
		if _, dup := t.targets[id]; dup {
			return fmt.Errorf("spell %d has two SpellTargetRestrictions rows at difficulty 0", id)
		}
		t.targets[id] = maxTargets
		return nil
	})
}

func (t *spellTables) loadEquippedItems(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(EquippedItemClass, 0), COALESCE(EquippedItemSubclass, 0),
		       COALESCE(EquippedItemInvTypes, 0)
		FROM SpellEquippedItems ORDER BY SpellID`, func(rows *sql.Rows) error {
		var id int32
		var class int64
		var e equippedRow
		if err := rows.Scan(&id, &class, &e.Subclass, &e.InvTypes); err != nil {
			return err
		}
		e.Class = int8(class)
		if _, dup := t.equipped[id]; dup {
			return fmt.Errorf("spell %d has two SpellEquippedItems rows", id)
		}
		t.equipped[id] = e
		return nil
	})
}

func (t *spellTables) loadLabels(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, LabelID FROM SpellLabel ORDER BY SpellID, LabelID`, func(rows *sql.Rows) error {
		var id int32
		var label int16
		if err := rows.Scan(&id, &label); err != nil {
			return err
		}
		t.labels[id] = append(t.labels[id], label)
		return nil
	})
}

// Every bar the spell costs from, in the client's order, so Power(t) finds the one it asks for.
func (t *spellTables) loadPowers(db *sql.DB) error {
	return eachRow(db, `
		SELECT SpellID, COALESCE(PowerType, 0), COALESCE(ManaCost, 0), COALESCE(ManaCostPerLevel, 0),
		       COALESCE(PowerCostPct, 0), COALESCE(ManaPerSecond, 0)
		FROM SpellPower ORDER BY SpellID, OrderIndex`, func(rows *sql.Rows) error {
		var id int32
		var powerType int64
		var p storePower
		if err := rows.Scan(&id, &powerType, &p.Cost, &p.CostPerLevel, &p.CostPct, &p.PerSecond); err != nil {
			return err
		}
		p.Type = int8(powerType)
		t.powers[id] = append(t.powers[id], p)
		return nil
	})
}

// EffectMiscValue_0/_1, EffectSpellClassMask_0..3, EffectRadiusIndex_0 and ImplicitTarget_0/_1 are
// generated columns over the JSON arrays, so each is named rather than reached through SELECT *.
// RadiusMin and RadiusMax are SpellRadius' own columns: two radius rows state a Radius with no
// maximum, and those read as zero here rather than as a radius the client does not state there.
func (t *spellTables) loadEffects(db *sql.DB) error {
	return eachRow(db, `
		SELECT e.SpellID, e.ID, e.EffectIndex, e.Effect, e.EffectAura,
		       COALESCE(e.EffectBasePointsF, 0), COALESCE(e.EffectRealPointsPerLevel, 0),
		       COALESCE(e.Variance, 0), COALESCE(e.EffectBonusCoefficient, 0),
		       COALESCE(e.BonusCoefficientFromAP, 0), COALESCE(e.PvpMultiplier, 0),
		       COALESCE(e.EffectAmplitude, 0), COALESCE(e.EffectAuraPeriod, 0),
		       COALESCE(r.RadiusMin, 0), COALESCE(r.RadiusMax, 0),
		       COALESCE(e.EffectMiscValue_0, 0), COALESCE(e.EffectMiscValue_1, 0),
		       COALESCE(e.EffectSpellClassMask_0, 0), COALESCE(e.EffectSpellClassMask_1, 0),
		       COALESCE(e.EffectSpellClassMask_2, 0), COALESCE(e.EffectSpellClassMask_3, 0),
		       COALESCE(e.EffectTriggerSpell, 0), COALESCE(e.EffectChainTargets, 0),
		       COALESCE(e.EffectChainAmplitude, 0), COALESCE(e.EffectMechanic, 0),
		       COALESCE(e.EffectPointsPerResource, 0),
		       COALESCE(e.ImplicitTarget_0, 0), COALESCE(e.ImplicitTarget_1, 0),
		       COALESCE(e.EffectAttributes, 0)
		FROM SpellEffect e
		LEFT JOIN SpellRadius r ON r.ID = e.EffectRadiusIndex_0
		WHERE e.DifficultyID = 0
		ORDER BY e.SpellID, e.EffectIndex, e.ID`, func(rows *sql.Rows) error {
		var e storeEffect
		var index, mask [4]int64
		var mechanic, target0, target1 int64
		if err := rows.Scan(&e.SpellID, &e.ID, &index[0], &e.Type, &e.Aura,
			&e.BasePoints, &e.PPL, &e.Variance, &e.SPCoef, &e.APCoef, &e.PvpMult,
			&e.Amplitude, &e.PeriodMs, &e.RadiusMin, &e.RadiusMax,
			&e.Misc, &e.Misc2,
			&mask[0], &mask[1], &mask[2], &mask[3],
			&e.TriggerID, &e.ChainTargets, &e.ChainAmp, &mechanic, &e.PointsPerResource,
			&target0, &target1, &e.Attributes); err != nil {
			return err
		}
		e.Index = uint8(index[0])
		for i, word := range mask {
			e.ClassFlags.Mask[i] = uint32(word)
		}
		e.Mechanic = uint8(mechanic)
		e.Target = [2]uint8{uint8(target0), uint8(target1)}

		// The rows arrive in index order, so a second row at one index is the one just read. The
		// store would carry both, and the position EffectN counts by would answer the first.
		if prior := t.effects[e.SpellID]; len(prior) > 0 && prior[len(prior)-1].Index == e.Index {
			return fmt.Errorf("spell %d states effect index %d twice, as rows %d and %d",
				e.SpellID, e.Index, prior[len(prior)-1].ID, e.ID)
		}

		t.effects[e.SpellID] = append(t.effects[e.SpellID], e)
		return nil
	})
}

// The spell ids the tooltip names, in the order it names them, each once. An id no SpellName row
// carries is left out: the closure walks the same tokens and only follows the ones that are spells.
func (t *spellTables) referencedIDs(id int32) []int32 {
	var refs []int32
	seen := map[int32]bool{}
	for _, m := range descriptionSpellRef.FindAllStringSubmatch(t.descriptions[id], -1) {
		ref := parseSpellID(m[1])
		if ref == 0 || ref == id || seen[ref] {
			continue
		}
		seen[ref] = true
		if _, named := t.names[ref]; named {
			refs = append(refs, ref)
		}
	}
	return refs
}

// The ids in the store, in the order Find binary searches them in.
func sortedIDs(set map[int32]bool) []int32 {
	ids := make([]int32, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func eachRow(db *sql.DB, query string, scan func(*sql.Rows) error) error {
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("%s: %w", query, err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("%s: %w", query, err)
		}
	}
	return rows.Err()
}
