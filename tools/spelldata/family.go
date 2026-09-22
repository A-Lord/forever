package main

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/wowsims/forever/sim/core/spelldata"
)

// The ladder a caller names: `warrior/Execute`, or `Execute` alone when one class states it, with the
// package the caller's file sits in as the tiebreak. A name several classes state and no package to
// choose between them is an error rather than a guess.
func findFamily(index map[string]*ladderFamily, spec, pkg string) (*ladderFamily, error) {
	field := spec
	if class, name, ok := strings.Cut(spec, "/"); ok {
		pkg, field = class, name
	}
	if field == "" {
		return nil, fmt.Errorf("name a ladder, as <class>/<Family> or <Family>")
	}

	if pkg != "" {
		family, ok := index[pkg+"/"+field]
		if !ok {
			return nil, fmt.Errorf("%s states no ladder named %q", pkg, field)
		}
		return family, nil
	}

	var found []*ladderFamily
	for _, family := range index {
		if family.field == field {
			found = append(found, family)
		}
	}
	switch len(found) {
	case 0:
		return nil, fmt.Errorf("no class file states a ladder named %q", field)
	case 1:
		return found[0], nil
	}

	classes := make([]string, 0, len(found))
	for _, family := range found {
		classes = append(classes, family.pkg)
	}
	sort.Strings(classes)
	return nil, fmt.Errorf("%s is a ladder in %s - name one as <class>/%s",
		field, strings.Join(classes, ", "), field)
}

// The ladder calls a class file reaches one rank through, which is what -expr takes: the field, with or
// without the `spellData.` the class files write, and one of the three accessors that names a rank.
var exprPattern = regexp.MustCompile(`^(?:spellData\.)?(\w+)\.(?:(Highest)\(\)|(Rank)\((\d+)\)|(ByID)\((\d+)\))$`)

// The row a ladder call reaches, and for a talent the rank it reads, counted from 1. A talent's row is
// the rank Talent builds from the curve, not the store's base row.
type pick struct {
	family *ladderFamily
	spell  *spelldata.Spell
	rank   int32
}

func resolveExpr(index map[string]*ladderFamily, expr, pkg string) (pick, error) {
	match := exprPattern.FindStringSubmatch(strings.TrimSpace(expr))
	if match == nil {
		return pick{}, fmt.Errorf("%q is not a ladder call: write spellData.<Family>.Highest(), .Rank(n) or .ByID(id)", expr)
	}

	family, err := findFamily(index, match[1], pkg)
	if err != nil {
		return pick{}, err
	}

	var id, rank int32
	switch {
	case match[2] != "":
		id, rank = family.highest(), family.rankCount()

	case match[3] != "":
		n, err := strconv.ParseInt(match[4], 10, 32)
		if err != nil {
			return pick{}, err
		}
		found, ok := family.rankID(int32(n))
		if !ok {
			return pick{}, fmt.Errorf("%s spellData.%s has %d ranks, not rank %d",
				family.pkg, family.field, family.rankCount(), n)
		}
		id, rank = found, int32(n)

	default:
		n, err := strconv.ParseInt(match[6], 10, 32)
		if err != nil {
			return pick{}, err
		}
		if !family.carries(int32(n)) {
			return pick{}, fmt.Errorf("%s spellData.%s has no rank with id %d",
				family.pkg, family.field, n)
		}
		// Ladder.ByID answers the first rank that carries the id, which on a talent is rank 1.
		id, rank = int32(n), 1
	}

	if family.talentRanks == 0 {
		s := spelldata.Find(id)
		if s == spelldata.Nil {
			return pick{}, fmt.Errorf("spell %d is not in the store", id)
		}
		return pick{family: family, spell: s}, nil
	}

	ladder, err := family.talentLadder()
	if err != nil {
		return pick{}, err
	}
	return pick{family: family, spell: ladder.Rank(rank), rank: rank}, nil
}

// The ladder the class file builds for a talent. Talent panics on a spell the store does not carry and
// on a curve whose rank count disagrees with the talent's, and that panic is answered as an error.
func (f *ladderFamily) talentLadder() (ladder spelldata.Ladder, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	return spelldata.Talent(f.ranks[0].id, f.talentRanks), nil
}

func (p pick) title() string {
	return rankTitle(p.spell, p.rank, p.family.talentRanks)
}

// The rank a row a ladder reached is: the talent rank where it is one, else the store's rank column.
func rankLabel(s *spelldata.Spell, rank, ranks int32) string {
	if rank == 0 {
		return s.Rank
	}
	return fmt.Sprintf("rank %d of %d", rank, ranks)
}

// The heading of a row a ladder reached: a talent's rank is not the store's rank column, so it is
// stated here. A rank of 0 is a row that is not a talent's rank.
func rankTitle(s *spelldata.Spell, rank, ranks int32) string {
	if rank == 0 {
		return title(s)
	}
	return fmt.Sprintf("%s (rank %d of %d)", title(s), rank, ranks)
}

func (f *ladderFamily) rankCount() int32 {
	if f.talentRanks > 0 {
		return f.talentRanks
	}
	return int32(len(f.ranks))
}

// A talent's ranks are all the one spell, so every rank the curve states answers that id.
func (f *ladderFamily) rankID(n int32) (int32, bool) {
	if n <= 0 || n > f.rankCount() {
		return 0, false
	}
	if f.talentRanks > 0 {
		return f.ranks[0].id, true
	}
	return f.ranks[n-1].id, true
}

func (f *ladderFamily) carries(id int32) bool {
	for _, rank := range f.ranks {
		if rank.id == id {
			return true
		}
	}
	return false
}

func (f *ladderFamily) highest() int32 {
	return f.ranks[len(f.ranks)-1].id
}

type familyRankJSON struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Rank     string `json:"rank"`
	Accessor string `json:"accessor"`
	Value    string `json:"value,omitempty"`
}

type familyJSON struct {
	Family  string           `json:"family"`
	Ranks   []familyRankJSON `json:"ranks"`
	Highest spellJSON        `json:"highest"`
}

// The rank index, with each id's name and rank read out of the store. An id the store does not carry
// says so in place of a name rather than stopping the listing, since the ladder is still worth reading.
func familyRows(f *ladderFamily) []familyRankJSON {
	if f.talentRanks > 0 {
		if rows, err := talentRows(f); err == nil {
			return rows
		}
	}

	rows := make([]familyRankJSON, 0, len(f.ranks))
	for _, rank := range f.ranks {
		row := familyRankJSON{ID: rank.id, Accessor: rank.accessor, Name: "not in the store"}
		if s := spelldata.Find(rank.id); s != spelldata.Nil {
			row.Name, row.Rank = s.Name, s.Rank
		}
		rows = append(rows, row)
	}
	return rows
}

// One row per rank of the talent's ladder, with effect 1's value where the curve changes it by rank.
func talentRows(f *ladderFamily) ([]familyRankJSON, error) {
	ladder, err := f.talentLadder()
	if err != nil {
		return nil, err
	}

	varies := false
	ladder.Each(func(rank int32, s *spelldata.Spell) {
		varies = varies || s.EffectN(1).BasePoints != ladder.Rank(1).EffectN(1).BasePoints
	})

	rows := make([]familyRankJSON, 0, ladder.Len())
	ladder.Each(func(rank int32, s *spelldata.Spell) {
		row := familyRankJSON{
			ID:       s.ID,
			Name:     s.Name,
			Rank:     fmt.Sprintf("rank %d of %d", rank, ladder.Len()),
			Accessor: fmt.Sprintf("Rank(%d)", rank),
		}
		if rank == ladder.Len() {
			row.Accessor = "Highest()"
		}
		if varies {
			row.Value = "effect 1 = " + number(s.EffectN(1).BasePoints)
		}
		rows = append(rows, row)
	})
	return rows, nil
}

// The rank a bare Highest() reaches.
func familyHighest(f *ladderFamily) (pick, error) {
	if f.talentRanks == 0 {
		s := spelldata.Find(f.highest())
		if s == spelldata.Nil {
			return pick{}, fmt.Errorf("spell %d is not in the store", f.highest())
		}
		return pick{family: f, spell: s}, nil
	}
	ladder, err := f.talentLadder()
	if err != nil {
		return pick{}, err
	}
	return pick{family: f, spell: ladder.Highest(), rank: ladder.Len()}, nil
}

func writeFamilyText(out io.Writer, f *ladderFamily) {
	rows := familyRows(f)

	width, rankWidth := 0, 8
	for _, row := range rows {
		width = max(width, len(row.Name))
		rankWidth = max(rankWidth, len(row.Rank))
	}

	fmt.Fprintf(out, "%s spellData.%s\n", f.pkg, f.field)
	for _, row := range rows {
		if row.Value == "" {
			fmt.Fprintf(out, "%-8d %-*s  %-*s %s\n", row.ID, width, row.Name, rankWidth, row.Rank, row.Accessor)
			continue
		}
		fmt.Fprintf(out, "%-8d %-*s  %-*s %-9s %s\n", row.ID, width, row.Name, rankWidth, row.Rank, row.Accessor, row.Value)
	}
	fmt.Fprintln(out)
}
