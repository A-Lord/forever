package database

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
)

// Where the store's hand-kept extra ids live. Parsed out of the source rather than read off the
// package: the ids a render is built from are then the ones the committed file states, so adding one
// without regenerating fails the check instead of passing it.
const extraIDsPath = "sim/core/spelldata/extra_ids.go"

// The spells the sim can reach without anything naming them first: everything the nine class files
// are built from, and every spell an item, an enchant or a set bonus casts. What those spells trigger
// is reached from here by reachableSpells.
//
// The store's own extra ids are not here: they are read out of sim/core/spelldata/extra_ids.go while
// the store is rendered, so that adding one and forgetting to regenerate fails the regeneration check
// instead of passing it - see withExtraIDs.
// The gear half is returned on its own as well: it is the universe the proc audit asks its question
// over, and which query a root came from cannot be recovered from the merged list.
func storeRoots(db *sql.DB, t *spellTables, ladderIDs []int32) (roots []int32, gear []int32, err error) {
	all := map[int32]bool{}
	for _, id := range ladderIDs {
		all[id] = true
	}

	gearRoots := map[int32]bool{}
	for _, source := range []struct {
		name  string
		query string
	}{
		{"item effects", itemEffectSpellQuery()},
		{"enchant effects", enchantSpellQuery},
		{"set bonuses", `SELECT DISTINCT SpellID FROM ItemSetSpell WHERE SpellID > 0`},
	} {
		if err := eachRow(db, source.query, func(rows *sql.Rows) error {
			var id int32
			if err := rows.Scan(&id); err != nil {
				return err
			}
			all[id] = true
			gearRoots[id] = true
			return nil
		}); err != nil {
			return nil, nil, fmt.Errorf("%s: %w", source.name, err)
		}
	}

	ids := make([]int32, 0, len(all))
	for id := range all {
		ids = append(ids, id)
	}

	gearIDs := make([]int32, 0, len(gearRoots))
	for id := range gearRoots {
		gearIDs = append(gearIDs, id)
	}

	return namedIDs(t, ids), namedIDs(t, gearIDs), nil
}

// The spells a rank reads its numbers off by name. The class-table generator falls back to them
// wherever a rank states no amount of its own - Frenzied Regeneration's heal is on 22845, Tiger's
// Fury's energize on 417045 - and the store has to carry the same spells, or the sim can read a
// number off the table that it cannot read off the store. Mirrors SiblingRankEffects: same name,
// same rank subtext, a shared class bit, and taught by a skill line rather than merely existing.
//
// One pass over everything already reached is enough: the relation is the name and subtext, so a
// sibling's siblings are the ones already in hand.
func siblingSpells(db *sql.DB, ids []int32) ([]int32, error) {
	list := make([]string, len(ids))
	for i, id := range ids {
		list[i] = strconv.Itoa(int(id))
	}

	var siblings []int32
	err := eachRow(db, `
		SELECT DISTINCT b.Spell
		FROM SkillLineAbility a
		JOIN SkillLineAbility b ON b.Spell != a.Spell AND (a.ClassMask & b.ClassMask) != 0
		JOIN SpellName na ON na.ID = a.Spell
		JOIN SpellName nb ON nb.ID = b.Spell AND nb.Name_lang = na.Name_lang
		JOIN Spell sa ON sa.ID = a.Spell
		JOIN Spell sb ON sb.ID = b.Spell AND sb.NameSubtext_lang = sa.NameSubtext_lang
		WHERE a.Spell IN (`+strings.Join(list, ", ")+`)
		ORDER BY b.Spell`, func(rows *sql.Rows) error {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return err
		}
		siblings = append(siblings, id)
		return nil
	})
	return siblings, err
}

// The store's hand-kept extra ids, added to the captured roots while the store is rendered rather
// than while the client tables are read: the ids live in the store's own source, so resolving them
// here is what makes an id added without a regeneration show up as a row the committed store lacks.
func withExtraIDs(t *spellTables, roots []int32) ([]int32, error) {
	extras, err := parseExtraIDs()
	if err != nil {
		return nil, err
	}

	// A hand-kept id this build does not name is dropped like any other unnamed id, and unlike an
	// item effect pointing at a spell the client retired, it was written down on purpose.
	for _, id := range extras {
		if _, named := t.names[id]; !named {
			fmt.Fprintf(progress, "spelldata: %s names %d, which is no spell in this client, so the store does not carry it\n",
				extraIDsPath, id)
		}
	}

	return namedIDs(t, roots, extras), nil
}

// The ids of every list that this build names as a spell, deduped and in search order. An id no
// SpellName row carries is not a spell here: ItemEffect keeps rows for spells the client dropped,
// and following one would put an empty row in the store.
func namedIDs(t *spellTables, lists ...[]int32) []int32 {
	set := map[int32]bool{}
	for _, list := range lists {
		for _, id := range list {
			if _, named := t.names[id]; named {
				set[id] = true
			}
		}
	}
	return sortedIDs(set)
}

// The gear and consumables gen_db ships, which is where the item procs the sim registers come from.
// The gear half is the predicate gen_db itself passes to LoadAndWriteRawItems; the consumable half
// states the classes LoadAndWriteConsumables selects, whose query cannot be reused as it stands
// because its own filter reads a subquery alias. Both halves let the allowlists through, as gen_db
// does: Hand of Justice and the raid consumables are shipped by id rather than by predicate.
//
// The predicate runs over Item left-joined to ItemSparse alone, without the inner joins to
// ItemClass, RandPropPoints and the armour tables that gen_db's own query carries, so this selects
// at least the items gen_db ships and possibly a few more. Extra rows only add spells to the store,
// which is the safe direction: a missing one is a spell the sim cannot read.
func itemEffectSpellQuery() string {
	return `
		SELECT DISTINCT ie.SpellID
		FROM ItemEffect ie
		JOIN ItemXItemEffect ixie ON ixie.ItemEffectID = ie.ID
		JOIN Item i ON i.ID = ixie.ItemID
		LEFT JOIN ItemSparse s ON s.ID = i.ID
		WHERE ie.SpellID > 0 AND (
			(` + SimItemFilter(core.CharacterLevel) + `)
			OR (((i.ClassID = 0 AND i.SubclassID IS NOT 0 AND i.SubclassID IS NOT 8 AND i.SubclassID IS NOT 6)
			     OR (i.ClassID = 7 AND i.SubclassID = 2))
			    AND s.RequiredLevel >= 50 AND s.Display_lang != ''
			    AND s.Display_lang NOT LIKE '%Test%' AND s.Display_lang NOT LIKE 'QA%')
			OR i.ID IN (` + allowListedItemIDs() + `))`
}

// The items gen_db ships by id: ItemAllowList bypasses the gear filter and ConsumableAllowList is
// added to the consumable query. Hand of Justice and Skullflame Shield have no ItemSparse row in
// this build, which is why the join to it above is a left one: an allowlisted item is taken by id,
// whatever the sparse table says about it.
func allowListedItemIDs() string {
	ids := make([]int32, 0, len(ItemAllowList)+len(ConsumableAllowList))
	for id := range ItemAllowList {
		ids = append(ids, id)
	}
	ids = append(ids, ConsumableAllowList...)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	list := make([]string, len(ids))
	for i, id := range ids {
		list[i] = strconv.Itoa(int(id))
	}
	return strings.Join(list, ", ")
}

// An enchant states the spell it applies in the EffectArg of the effect that applies it, which is
// what LoadAndWriteRawEnchants reads as its spell id: effect 1 and effect 3 are the two that name a
// spell.
const enchantSpellQuery = `
	SELECT DISTINCT EffectArg_0 FROM SpellItemEnchantment WHERE Effect_0 IN (1, 3) AND EffectArg_0 > 0
	UNION
	SELECT DISTINCT EffectArg_1 FROM SpellItemEnchantment WHERE Effect_1 IN (1, 3) AND EffectArg_1 > 0
	UNION
	SELECT DISTINCT EffectArg_2 FROM SpellItemEnchantment WHERE Effect_2 IN (1, 3) AND EffectArg_2 > 0`

// Every spell the roots reach: what an effect triggers, what an actionbar override swaps in, what a
// hand link names, and what a tooltip's $<id> token points at. A spell the sim registers reads its
// numbers off those, so the store carries them all rather than the roots alone.
func reachableSpells(t *spellTables, roots []int32) []int32 {
	seen := map[int32]bool{}
	queue := make([]int32, 0, len(roots))
	for _, id := range roots {
		if !seen[id] {
			seen[id] = true
			queue = append(queue, id)
		}
	}

	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]

		for _, next := range spellEdges(t, id) {
			if seen[next] {
				continue
			}
			if _, named := t.names[next]; !named {
				continue
			}
			seen[next] = true
			queue = append(queue, next)
		}
	}
	return sortedIDs(seen)
}

// A_OVERRIDE_ACTIONBAR_SPELLS states the spell it swaps in as its base points, the way
// overrideReplacements reads it.
func spellEdges(t *spellTables, id int32) []int32 {
	var next []int32
	for _, e := range t.effects[id] {
		if e.TriggerID > 0 {
			next = append(next, e.TriggerID)
		}
		if e.Aura == int32(dbcenums.A_OVERRIDE_ACTIONBAR_SPELLS) && e.BasePoints > 0 {
			next = append(next, int32(e.BasePoints))
		}
	}
	next = append(next, handTriggers[id]...)
	return append(next, t.referencedIDs(id)...)
}

// The store's ExtraIDs, read out of its source: a Go program cannot ask a package for the values of
// a variable it does not import.
func parseExtraIDs() ([]int32, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, extraIDsPath, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", extraIDsPath, err)
	}

	var ids []int32
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		vs, ok := n.(*ast.ValueSpec)
		if !ok || len(vs.Names) != 1 || vs.Names[0].Name != "ExtraIDs" || len(vs.Values) != 1 {
			return true
		}
		lit, ok := vs.Values[0].(*ast.CompositeLit)
		if !ok {
			return true
		}
		found = true
		for _, elt := range lit.Elts {
			id, ok := constInt(elt)
			if !ok {
				fmt.Fprintf(progress, "spelldata: %s:%d is not a literal id, so the store does not carry it\n",
					extraIDsPath, fset.Position(elt.Pos()).Line)
				continue
			}
			ids = append(ids, id)
		}
		return false
	})
	if !found {
		return nil, fmt.Errorf("%s declares no ExtraIDs slice", extraIDsPath)
	}
	return ids, nil
}

func parseSpellID(s string) int32 {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int32(id)
}
