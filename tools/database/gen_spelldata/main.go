// Regenerates the spell store, sim/<class>/spell_data_auto_gen.go, the hunter's
// pet_families_auto_gen.go and ammo_auto_gen.go, and the raid buffs from the game files database.
//
// Its own binary, not a mode of gen_db: gen_db imports every class package, so a class file that
// does not compile stops it. This binary imports sim/core, sim/core/spelldata and sim/common - the
// item-proc routing in tools/database/gen_effects.go reads the store - so the class files are the
// generated files it can rewrite without compiling them, and a broken store stops it the same way a
// broken class file stops gen_db. Neither reaches the tree from here: spelldata_write.go
// type-checks the rendered files in a staging directory and writes none of them until they build -
// unless -unchecked says to write them anyway, for a class file whose call sites cannot compile
// until they move onto what it will state.
//
//	go run ./tools/database/gen_spelldata
//	go run ./tools/database/gen_spelldata -only druid,racials   # those portions from the database, the rest as committed
//	go run ./tools/database/gen_spelldata -check
//	go run ./tools/database/gen_spelldata -unchecked
//	go run ./tools/database/gen_spelldata -everyClassSpell -unchecked   # a store of every class spell, for tools/spelldata -dump
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/wowsims/forever/tools/database"
)

var dbPath = flag.String("dbPath", "./tools/database/wowsims.db", "Location of the wowsims.db file produced by tools/db2tool")
var check = flag.Bool("check", false, "Name the generated files that are not what this generator writes, and write nothing")
var everyClassSpell = flag.Bool("everyClassSpell", false, "Root the store at every class spell the game files have, not only what the sim reaches - for an export with tools/spelldata -dump, not to commit")
var unchecked = flag.Bool("unchecked", false, "Write the rendered files without type-checking them first, for a class file whose call sites cannot compile until they move onto what it will state")
var onlyList = flag.String("only", "", "Take only these portions from the database (class names and racials, comma-separated; items and shared always), keeping the rest as the committed inputs state them")

func main() {
	flag.Parse()
	database.DatabasePath = *dbPath
	database.EveryClassSpell = *everyClassSpell

	only, err := database.ParsePortions(*onlyList)
	if err != nil {
		log.Fatalf("-only: %v", err)
	}
	if only != nil && (*check || *everyClassSpell) {
		log.Fatalf("-only does not combine with -check or -everyClassSpell")
	}

	helper, err := database.NewDBHelper()
	if err != nil {
		log.Fatalf("failed to open %s: %v", *dbPath, err)
	}
	defer helper.Close()

	if *check {
		stale, behind, err := database.CheckSpellDataFiles(helper)
		if err != nil {
			log.Fatalf("failed to generate spell data tables: %v", err)
		}
		for _, line := range behind {
			fmt.Fprintf(os.Stderr, "not ported: %s\n", line)
		}
		for _, path := range stale {
			fmt.Fprintln(os.Stderr, path)
		}
		if len(stale) > 0 {
			os.Exit(1)
		}
		return
	}

	if only == nil && !*everyClassSpell {
		var behind []string
		if only, behind, err = database.DefaultPortions(helper); err != nil {
			log.Fatalf("failed to read the committed portions: %v", err)
		}
		for _, line := range behind {
			fmt.Fprintf(os.Stderr, "kept, not ported: %s\n", line)
		}
	}
	if err := database.GenerateSpellDataFiles(helper, *unchecked, only); err != nil {
		log.Fatalf("failed to generate spell data tables: %v", err)
	}
}
