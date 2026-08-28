package services

import "openelis-go/internal/common/util"

// JavaHashSetOrder reorders {id,value} pairs into the iteration order a
// java.util.HashSet<String> of those ids would produce.
//
// WHY THIS EXISTS: getUserSampleTypes collects its sample-type ids into a
// HashSet and then iterates it, so the wire order of `sampleTypes` on
// SamplePatientEntry and SampleEdit is HashMap bucket order — not id order, not
// sort order, not insertion order. On the dev dataset Java returns
// 34,1,2,35,3,36,4,37,26,30,31,32, which looks arbitrary and is fully
// determined: bucket = spread(hashCode) & (capacity-1), buckets walked in
// ascending index, entries within a bucket in insertion order.
//
// Reproducing a JDK internal is not something to do lightly, but the
// alternative is emitting a different order than Java on a live endpoint, and
// the ordering is deterministic rather than incidental — the same ids always
// produce the same sequence.
//
// capacity follows HashMap's growth: 16 initially, doubling once size exceeds
// 0.75 * capacity.
func JavaHashSetOrder(pairs []util.IdValuePair) []util.IdValuePair {
	if len(pairs) == 0 {
		return pairs
	}

	capacity := 16
	for len(pairs) > (capacity*3)/4 {
		capacity *= 2
	}

	buckets := make([][]util.IdValuePair, capacity)
	for _, p := range pairs {
		idx := javaSpread(javaStringHashCode(p.Id)) & (capacity - 1)
		buckets[idx] = append(buckets[idx], p)
	}

	out := make([]util.IdValuePair, 0, len(pairs))
	for _, b := range buckets {
		out = append(out, b...)
	}
	return out
}

// javaStringHashCode is String.hashCode: s[0]*31^(n-1) + s[1]*31^(n-2) + …,
// computed with int32 wraparound.
func javaStringHashCode(s string) int32 {
	var h int32
	for i := 0; i < len(s); i++ {
		h = 31*h + int32(s[i])
	}
	return h
}

// javaSpread is HashMap.hash: h ^ (h >>> 16), an unsigned shift.
func javaSpread(h int32) int {
	u := uint32(h)
	return int(u ^ (u >> 16))
}
