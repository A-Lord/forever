package database

// Rebuilds sim/core/spelldata/spells_auto_gen.go from the client rows committed in
// assets/db_inputs/spell_store_inputs.json and asserts the committed file is what comes out, so a
// hand-edited or stale store fails here. No client database and no build tag: this is the one gate
// over the store's 7035 rows that CI can run, since tools/database/wowsims.db is gitignored.
//
// It re-runs everything the generator derives - the closure over triggers, overrides and tooltip
// references, the hand links, the talent curves, the tooltip hints, the overrides and the emitter -
// against rows it did not derive. A change to any of those fails until the store is regenerated with
// `go run ./tools/database/gen_spelldata`, which rewrites both files together.

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The generator reads its paths from the repository root - the enum names out of sim/core/dbcenums,
// the extra ids out of the store's own source - and a test runs in its own package's directory.
const repositoryRoot = "../.."

func TestStoreRegeneratesFromTheCommittedInputs(t *testing.T) {
	inRepositoryRoot(t)

	inputs, err := readStoreInputs(spellStoreInputsPath)
	if err != nil {
		t.Fatalf("%v", err)
	}

	namer, err := newRankEnumNamer()
	if err != nil {
		t.Fatalf("reading the spell enum names: %v", err)
	}

	rendered, err := renderStore(inputs, namer)
	if err != nil {
		t.Fatalf("rendering the store: %v", err)
	}

	const storePath = "sim/core/spelldata/spells_auto_gen.go"
	committed, err := os.ReadFile(storePath)
	if err != nil {
		t.Fatalf("%v", err)
	}
	if bytes.Equal(committed, rendered) {
		return
	}

	line, want, got := firstDifference(committed, rendered)
	t.Errorf("%s is not what the committed inputs render, from line %d:\n  committed: %s\n  rendered:  %s\n"+
		"regenerate both with `go run ./tools/database/gen_spelldata`",
		storePath, line, want, got)
}

// The rendering counts its rows on stderr, which says nothing a passing gate needs, so it goes
// nowhere for the duration.
func inRepositoryRoot(t *testing.T) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%v", err)
	}
	if err := os.Chdir(filepath.Join(cwd, repositoryRoot)); err != nil {
		t.Fatalf("%v", err)
	}

	progress = io.Discard
	t.Cleanup(func() {
		progress = os.Stderr
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("%v", err)
		}
	})
}

// The first line the two differ on, so a diff of megabytes reports as one row.
func firstDifference(want, got []byte) (int, string, string) {
	a, b := lines(want), lines(got)
	for i := range a {
		if i >= len(b) {
			return i + 1, a[i], "(the rendered store ends here)"
		}
		if a[i] != b[i] {
			return i + 1, a[i], b[i]
		}
	}
	if len(b) > len(a) {
		return len(a) + 1, "(the committed store ends here)", b[len(a)]
	}
	return 0, "", ""
}

func lines(b []byte) []string {
	var out []string
	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Buffer(make([]byte, 0, 1<<20), 1<<20)
	for scanner.Scan() {
		out = append(out, strings.TrimRight(scanner.Text(), " \t"))
	}
	return out
}
