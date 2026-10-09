//go:build with_db

package feralcat

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// The druid trees as the talent calculator draws them: per tree, each talent's row and highest rank.
type talentTree struct {
	Name    string `json:"name"`
	Talents []struct {
		FieldName string `json:"fieldName"`
		Location  struct {
			RowIdx int `json:"rowIdx"`
		} `json:"location"`
		MaxPoints int `json:"maxPoints"`
	} `json:"talents"`
}

func druidTrees(t *testing.T) []talentTree {
	t.Helper()
	raw, err := os.ReadFile("../../../ui/sim/talents/trees/druid.json")
	if err != nil {
		t.Fatal(err)
	}
	var trees []talentTree
	if err := json.Unmarshal(raw, &trees); err != nil {
		t.Fatal(err)
	}
	return trees
}

// talentsLegal reports why a talent string cannot be built at level 60: a rank over the talent's
// highest, a row reached before 5 points per row are spent above it in the same tree, or more than
// 51 points.
func talentsLegal(trees []talentTree, talents string) error {
	parts := strings.Split(talents, "-")
	if len(parts) != len(trees) {
		return fmt.Errorf("%d trees, want %d", len(parts), len(trees))
	}
	total := 0
	for i, tree := range trees {
		digits := parts[i]
		if len(digits) > len(tree.Talents) {
			return fmt.Errorf("%s: %d talents, the tree has %d", tree.Name, len(digits), len(tree.Talents))
		}
		perRow := map[int]int{}
		for j := range digits {
			points := int(digits[j] - '0')
			talent := tree.Talents[j]
			if points > talent.MaxPoints {
				return fmt.Errorf("%s: %s %d of %d", tree.Name, talent.FieldName, points, talent.MaxPoints)
			}
			perRow[talent.Location.RowIdx] += points
			total += points
		}
		for j := range digits {
			talent := tree.Talents[j]
			if digits[j] == '0' {
				continue
			}
			above := 0
			for row := 0; row < talent.Location.RowIdx; row++ {
				above += perRow[row]
			}
			if need := 5 * talent.Location.RowIdx; above < need {
				return fmt.Errorf("%s: %s is on row %d with %d points above, needs %d", tree.Name, talent.FieldName, talent.Location.RowIdx+1, above, need)
			}
		}
	}
	if total > 51 {
		return fmt.Errorf("%d points, a level 60 druid has 51", total)
	}
	return nil
}

// Every talent string the cat and bear UIs ship can be built in game.
func TestPresetTalentsAreLegal(t *testing.T) {
	trees := druidTrees(t)
	pattern := regexp.MustCompile(`talentsString: '([0-9-]+)'`)
	for _, file := range []string{
		"../../../ui/specs/druid/feralcat/presets.ts",
		"../../../ui/specs/druid/feralbear/presets.ts",
		"../../../ui/specs/druid/feralbear/spec.ts",
	} {
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range pattern.FindAllStringSubmatch(string(source), -1) {
			if err := talentsLegal(trees, match[1]); err != nil {
				t.Errorf("%s: %s: %v", file, match[1], err)
			}
		}
	}
	for name, talents := range map[string]string{"catTalents": catTalents, "bearweaveTalents": bearweaveTalents} {
		if err := talentsLegal(trees, talents); err != nil {
			t.Errorf("%s %s: %v", name, talents, err)
		}
	}
}
