package random

import "testing"

func TestIntnPanicsOnNonPositiveN(t *testing.T) {
	cases := []int{0, -1, -100}
	for _, n := range cases {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Intn(%d) did not panic", n)
				}
			}()
			CryptoSource{}.Intn(n)
		}()
	}
}

func TestIntnStaysInBounds(t *testing.T) {
	src := CryptoSource{}
	const n = 7
	for i := 0; i < 5000; i++ {
		v := src.Intn(n)
		if v < 0 || v >= n {
			t.Fatalf("Intn(%d) = %d, want value in [0, %d)", n, v, n)
		}
	}
}

// TestIntnCoversFullRange is a light distribution sanity check: over many
// trials, every bucket in [0, n) should get a roughly fair share. This is a
// probabilistic test with generous tolerance to avoid flakiness — its job
// is to catch a badly broken implementation (e.g. one that always returns
// 0, or applies naive modulo bias), not to certify statistical quality.
func TestIntnCoversFullRange(t *testing.T) {
	src := CryptoSource{}
	const n = 4
	const trials = 20000
	const expectedPerBucket = trials / n

	counts := make([]int, n)
	for i := 0; i < trials; i++ {
		counts[src.Intn(n)]++
	}

	minAllowed := expectedPerBucket / 2 // generous: 50% of a perfectly even split
	for bucket, count := range counts {
		if count < minAllowed {
			t.Errorf("Intn(%d) bucket %d got %d hits over %d trials, want at least %d",
				n, bucket, count, trials, minAllowed)
		}
	}
}

func TestShuffleProducesAPermutation(t *testing.T) {
	src := CryptoSource{}
	original := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	shuffled := append([]int(nil), original...)
	src.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	if len(shuffled) != len(original) {
		t.Fatalf("Shuffle changed length: got %d, want %d", len(shuffled), len(original))
	}

	seen := make(map[int]bool, len(original))
	for _, v := range shuffled {
		seen[v] = true
	}
	for _, v := range original {
		if !seen[v] {
			t.Fatalf("Shuffle lost element %d; result = %v", v, shuffled)
		}
	}
}

func TestShuffleVariesAcrossCalls(t *testing.T) {
	src := CryptoSource{}
	base := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	first := append([]int(nil), base...)
	src.Shuffle(len(first), func(i, j int) { first[i], first[j] = first[j], first[i] })

	// The probability that 19 further independent shuffles of 10 elements
	// all coincide with `first` is astronomically small (roughly (1/10!)^19)
	// unless Shuffle is broken (e.g. a no-op or a fixed permutation).
	for attempt := 0; attempt < 19; attempt++ {
		next := append([]int(nil), base...)
		src.Shuffle(len(next), func(i, j int) { next[i], next[j] = next[j], next[i] })
		if !equalInts(first, next) {
			return
		}
	}
	t.Fatal("20 independent Shuffle calls all produced the same order")
}

func TestShuffleHandlesSmallLengths(t *testing.T) {
	src := CryptoSource{}
	for _, n := range []int{0, 1} {
		called := false
		src.Shuffle(n, func(i, j int) { called = true })
		if called {
			t.Errorf("Shuffle(%d, ...) invoked swap, want no-op", n)
		}
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
