package database

// The override table is hand-kept data with a rule per field: the generator has to refuse a value
// the client has started stating itself, and a table nobody can regenerate is a table nobody
// trusts. These run the merge on synthetic rows, so they need no client database.

import (
	"strings"
	"testing"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/tools/database/overrides"
)

func TestOverridesBakeOntoTheRow(t *testing.T) {
	rows := []storeSpell{
		{ID: 10, Effects: []storeEffect{
			{Index: 0, Type: int32(dbcenums.E_SCHOOL_DAMAGE)},
			{Index: 1, Type: int32(dbcenums.E_APPLY_AURA), Aura: int32(dbcenums.A_PERIODIC_DAMAGE)},
		}},
	}

	list := []overrides.Override{
		{SpellID: 10, Field: overrides.PPM, Value: 1.5, Reason: "measured"},
		{SpellID: 10, Field: overrides.FlatThreat, Value: 200, Reason: "measured"},
		{SpellID: 10, Field: overrides.APCoefDirect, Value: 0.2, Reason: "measured"},
		{SpellID: 10, Field: overrides.APCoefPeriodic, Value: 0.1, Reason: "measured"},
		{SpellID: 10, Field: overrides.DurationMs, Value: 15000, Reason: "measured"},
		{SpellID: 10, Field: overrides.Hint, Value: float64(core.ProcHintCrit), Reason: "measured"},
	}
	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the overrides: %v", err)
	}

	row := rows[0]
	if row.RPPM != 1.5 || row.FlatThreat != 200 || row.DurationMs != 15000 {
		t.Errorf("the row reads RPPM %v, flat threat %v, duration %d", row.RPPM, row.FlatThreat, row.DurationMs)
	}
	if row.Effects[0].APCoef != 0.2 || row.Effects[1].APCoef != 0.1 {
		t.Errorf("the effects read %v direct and %v periodic attack power",
			row.Effects[0].APCoef, row.Effects[1].APCoef)
	}
	if row.ProcChanceSource != procChancePPM {
		t.Errorf("a PPM override leaves the chance source at %s", row.ProcChanceSource)
	}
	if row.ProcHint != core.ProcHintCrit {
		t.Errorf("the row reads hint %d", row.ProcHint)
	}
	if len(row.overrideNotes) != len(list) {
		t.Errorf("the row states %d override reasons for %d overrides", len(row.overrideNotes), len(list))
	}
}

// A proc chance override states the percentage the tooltip does not, and the store keeps the column
// as the source of it.
func TestProcChanceOverrideBecomesTheColumn(t *testing.T) {
	rows := []storeSpell{{ID: 10, ProcChanceSource: procChanceAlways, ProcChance: 101}}
	list := []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 15, Reason: "measured"}}

	if err := applyOverrides(rows, list); err != nil {
		t.Fatalf("baking the override: %v", err)
	}
	if rows[0].ProcChance != 15 || rows[0].ProcChanceSource != procChanceColumn {
		t.Errorf("the row reads chance %d from %s", rows[0].ProcChance, rows[0].ProcChanceSource)
	}
}

func TestOverridesRefuseWhatTheClientNowStates(t *testing.T) {
	for _, tc := range []struct {
		name  string
		rows  []storeSpell
		list  []overrides.Override
		wants string
	}{
		{
			name:  "a spell the store does not carry",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 11, Field: overrides.PPM, Value: 1, Reason: "measured"}},
			wants: "which the store does not carry",
		},
		{
			name:  "no reason",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.PPM, Value: 1}},
			wants: "states no reason",
		},
		{
			name: "the same field twice",
			rows: []storeSpell{{ID: 10}},
			list: []overrides.Override{
				{SpellID: 10, Field: overrides.PPM, Value: 1, Reason: "measured"},
				{SpellID: 10, Field: overrides.PPM, Value: 2, Reason: "measured again"},
			},
			wants: "two PPM overrides",
		},
		{
			name:  "a client procs-per-minute row",
			rows:  []storeSpell{{ID: 10, procsPerMinuteID: 7}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.PPM, Value: 1, Reason: "measured"}},
			wants: "now states SpellProcsPerMinuteID 7",
		},
		{
			name:  "a client threat effect",
			rows:  []storeSpell{{ID: 10, Effects: []storeEffect{{Index: 1, Type: int32(dbcenums.E_THREAT)}}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.FlatThreat, Value: 200, Reason: "measured"}},
			wants: "now states threat on effect 1",
		},
		{
			name: "a client attack-power coefficient",
			rows: []storeSpell{{ID: 10, Effects: []storeEffect{
				{Index: 0, Type: int32(dbcenums.E_SCHOOL_DAMAGE), APCoef: 0.5},
			}}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefDirect, Value: 0.2, Reason: "measured"}},
			wants: "now states 0.5 attack power on its direct effect",
		},
		{
			name:  "no effect to carry the coefficient",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.APCoefPeriodic, Value: 0.2, Reason: "measured"}},
			wants: "no periodic effect",
		},
		{
			name:  "a chance the tooltip states",
			rows:  []storeSpell{{ID: 10, tooltipStatesChance: true}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 15, Reason: "measured"}},
			wants: "states its own proc chance in the tooltip",
		},
		{
			name:  "a chance that is no percentage",
			rows:  []storeSpell{{ID: 10}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.ProcChancePct, Value: 150, Reason: "measured"}},
			wants: "no whole percentage",
		},
		{
			name:  "a client duration",
			rows:  []storeSpell{{ID: 10, DurationMs: 8000}},
			list:  []overrides.Override{{SpellID: 10, Field: overrides.DurationMs, Value: 15000, Reason: "measured"}},
			wants: "now states a duration of 8000 ms",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := applyOverrides(tc.rows, tc.list)
			if err == nil {
				t.Fatalf("the merge accepted %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wants) {
				t.Errorf("the error reads %q, which does not name %q", err, tc.wants)
			}
		})
	}
}

// Every shipped override has to name a field the merge knows, so a row added with a field the merge
// forgot fails here rather than at the next regeneration.
func TestShippedOverridesAreWellFormed(t *testing.T) {
	seen := map[int32]bool{}
	for _, o := range overrides.Spells {
		if o.Reason == "" || o.Source == "" {
			t.Errorf("spell %d's %s override states reason %q and source %q", o.SpellID, o.Field, o.Reason, o.Source)
		}
		if o.Field.String() == "unknown field" {
			t.Errorf("spell %d carries an override of field %d", o.SpellID, uint8(o.Field))
		}
		seen[o.SpellID] = true
	}
	if len(seen) == 0 {
		t.Error("the override table is empty, so nothing above checked anything")
	}
}
