package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func rangeOf(values ...float64) *ValueRange {
	r := &ValueRange{}
	for _, v := range values {
		r.add(v)
	}
	return r
}

func TestValueRangeAddTracksMinMaxAndAggregate(t *testing.T) {
	r := rangeOf(4, 2, 9, 5)
	if r.n != 4 || r.sum != 20 || r.sumSq != 16+4+81+25 {
		t.Fatalf("aggregate wrong: n=%d sum=%v sumSq=%v", r.n, r.sum, r.sumSq)
	}
	if r.min != 2 || r.max != 9 {
		t.Fatalf("min/max wrong: %v/%v", r.min, r.max)
	}
}

func TestValueRangeFirstValueSetsMinAndMax(t *testing.T) {
	r := rangeOf(-3)
	if r.min != -3 || r.max != -3 {
		t.Fatalf("min/max after one negative value: %v/%v", r.min, r.max)
	}
}

func TestValueRangeMerge(t *testing.T) {
	a := rangeOf(3, 7)
	b := rangeOf(1, 5)

	merged := &ValueRange{}
	merged.merge(a)
	if *merged != *a {
		t.Fatalf("empty receiver should copy: %+v vs %+v", *merged, *a)
	}
	merged.merge(&ValueRange{})
	if *merged != *a {
		t.Fatalf("empty argument should be a no-op: %+v", *merged)
	}
	merged.merge(b)
	if merged.n != 4 || merged.sum != 16 || merged.min != 1 || merged.max != 7 {
		t.Fatalf("merge wrong: %+v", *merged)
	}
	if *merged != *rangeOf(3, 7, 1, 5) {
		t.Fatalf("merge differs from adding every value: %+v", *merged)
	}
}

func TestValueRangeToProto(t *testing.T) {
	m := rangeOf(2, 4, 4, 4, 5, 5, 7, 9).toProto()
	if m.Avg != 5 || math.Abs(m.Stdev-2) > 1e-9 || m.Min != 2 || m.Max != 9 {
		t.Fatalf("unexpected summary: %+v", m)
	}
	if m.AggregatorData.N != 8 || m.AggregatorData.SumSq != 4+16*3+25*2+49+81 {
		t.Fatalf("unexpected aggregator data: %+v", m.AggregatorData)
	}
}

func TestValueRangeToProtoConstantValuesHaveZeroStdev(t *testing.T) {
	m := rangeOf(0.1, 0.1, 0.1, 0.1, 0.1, 0.1, 0.1).toProto()
	if math.IsNaN(m.Stdev) || m.Stdev < 0 {
		t.Fatalf("stdev should clamp to zero, got %v", m.Stdev)
	}
}

func TestCombineHitRangesWeightsByCount(t *testing.T) {
	base := rangeOf(10, 20).toProto()
	add := rangeOf(1, 2, 3, 40).toProto()
	got := combineHitRanges(base, add)

	want := rangeOf(10, 20, 1, 2, 3, 40).toProto()
	if got.AggregatorData.N != 6 || math.Abs(got.Avg-want.Avg) > 1e-9 || math.Abs(got.Stdev-want.Stdev) > 1e-9 {
		t.Fatalf("combined %+v, want %+v", got, want)
	}
	if got.Min != 1 || got.Max != 40 {
		t.Fatalf("min/max %v/%v", got.Min, got.Max)
	}
	if base.AggregatorData.N != 2 {
		t.Fatalf("base must stay untouched")
	}
}

func TestCombineHitRangesSkipsEmptySides(t *testing.T) {
	r := rangeOf(5, 9).toProto()
	for name, got := range map[string]*proto.DistributionMetrics{
		"nil base":    combineHitRanges(nil, r),
		"nil add":     combineHitRanges(r, nil),
		"empty base":  combineHitRanges(&proto.DistributionMetrics{}, r),
		"empty add":   combineHitRanges(r, &proto.DistributionMetrics{AggregatorData: &proto.AggregatorData{}}),
		"zero-N base": combineHitRanges(&proto.DistributionMetrics{Min: 0, Max: 0, AggregatorData: &proto.AggregatorData{}}, r),
	} {
		if got.Min != 5 || got.Max != 9 || got.Avg != 7 || got.AggregatorData.N != 2 {
			t.Fatalf("%s: %+v", name, got)
		}
	}
}

func TestCombineTargetHitRangesMergesIntoNilMapsAndCustomKeys(t *testing.T) {
	base := &proto.TargetedActionMetrics{}
	first := &proto.TargetedActionMetrics{
		Ranges: map[int32]*proto.DistributionMetrics{
			int32(proto.HitKind_HitKindHit):  rangeOf(100, 200).toProto(),
			int32(proto.HitKind_HitKindCrit): rangeOf(300).toProto(),
		},
		CustomMetrics: []*proto.CustomActionMetric{{Key: "a", Range: rangeOf(1).toProto()}},
	}
	second := &proto.TargetedActionMetrics{
		Ranges: map[int32]*proto.DistributionMetrics{
			int32(proto.HitKind_HitKindHit): rangeOf(400).toProto(),
		},
		CustomMetrics: []*proto.CustomActionMetric{
			{Key: "a", Range: rangeOf(3).toProto()},
			{Key: "b", Range: rangeOf(8).toProto()},
		},
	}
	combineTargetHitRanges(base, first)
	combineTargetHitRanges(base, second)

	hit := base.Ranges[int32(proto.HitKind_HitKindHit)]
	if hit.AggregatorData.N != 3 || hit.Min != 100 || hit.Max != 400 || math.Abs(hit.Avg-700.0/3) > 1e-9 {
		t.Fatalf("hit range %+v", hit)
	}
	crit := base.Ranges[int32(proto.HitKind_HitKindCrit)]
	if crit.AggregatorData.N != 1 || crit.Min != 300 || crit.Max != 300 {
		t.Fatalf("crit range kept only by the first side %+v", crit)
	}
	if len(base.CustomMetrics) != 2 || base.CustomMetrics[0].Key != "a" || base.CustomMetrics[1].Key != "b" {
		t.Fatalf("custom metrics %+v", base.CustomMetrics)
	}
	if a := base.CustomMetrics[0].Range; a.AggregatorData.N != 2 || a.Avg != 2 || a.Min != 1 || a.Max != 3 {
		t.Fatalf("custom range a %+v", a)
	}
}

func TestActionRangesFromDotTicksMatchCounters(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	fa.Dot.Apply(sim)

	const ticks = 200
	for range ticks {
		fa.Dot.TickOnce(sim)
	}
	target := sim.Encounter.ActiveTargetUnits[0]
	fa.Spell.RecordCustomMetric(sim, target, "bonus", 7)
	fa.Spell.RecordCustomMetric(sim, target, "bonus", 9)

	spellMetrics := fa.Spell.SpellMetrics[target.UnitIndex]
	fa.Spell.doneIteration()
	fa.Spell.doneIteration()

	am := fa.Metrics.actions[fa.Spell.ActionID]
	if len(am.rangeSources) != 1 {
		t.Fatalf("one spell split should register one range source, got %d", len(am.rangeSources))
	}
	tam := am.ToProto(fa.Spell.ActionID).Targets[target.UnitIndex]

	tick := tam.Ranges[int32(proto.HitKind_HitKindTick)]
	critTick := tam.Ranges[int32(proto.HitKind_HitKindCritTick)]
	if critTick == nil || tick == nil {
		t.Fatalf("expected both tick and crit tick ranges over %d ticks, got %v", ticks, tam.Ranges)
	}
	if int32(critTick.AggregatorData.N) != spellMetrics.CritTicks {
		t.Fatalf("crit tick N %d, counter %d", critTick.AggregatorData.N, spellMetrics.CritTicks)
	}
	if tick.AggregatorData.N+critTick.AggregatorData.N != ticks {
		t.Fatalf("every tick landed, got %d + %d of %d", tick.AggregatorData.N, critTick.AggregatorData.N, ticks)
	}

	var total float64
	for kind, r := range tam.Ranges {
		if r.Min > r.Avg || r.Avg > r.Max {
			t.Fatalf("kind %d: min %v avg %v max %v out of order", kind, r.Min, r.Avg, r.Max)
		}
		total += r.Avg * float64(r.AggregatorData.N)
	}
	if !WithinToleranceFloat64(spellMetrics.TotalDamage, total, 0.001) {
		t.Fatalf("ranges sum to %v, spell total %v", total, spellMetrics.TotalDamage)
	}

	if len(tam.CustomMetrics) != 1 || tam.CustomMetrics[0].Key != "bonus" || tam.CustomMetrics[0].Range.Avg != 8 {
		t.Fatalf("custom metrics %+v", tam.CustomMetrics)
	}
}
