package adler64rolling

import (
	"strings"
	"testing"
)

// This is a slow but simple implementation of the Adler-32 checksum, except
// it has been modified to use a 64-bit digest. It is also initialized with
// a 0 rather than 1. From my understanding, the initial 1 is used to remove
// the necessity to check the checksum length, but since this is designed to
// be used in a rolling checksum, we're working with fixed window sizes.
func checksum(p []byte) uint64 {
	s1, s2 := uint64(0), uint64(0)
	for _, x := range p {
		s1 = (s1 + uint64(x)) % mod
		s2 = (s2 + s1) % mod
	}
	return s2<<32 | s1
}

var rollStrings = []string{
	"You're a lizard, Harry.",
	"Don't tell me you've never seen a leprechaun before.",
	strings.Repeat("Too legit, too legit to quit, (hay haaaay).", 1e2),
	strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 1e3),
	strings.Repeat("\xff\x00", 1e4) + "hotdog",
}

func TestRolling(t *testing.T) {
	blocksize := 16
	rolling := New()

	for _, teststring := range rollStrings {
		test := []byte(teststring)
		for i := 0; i < len(test)-blocksize-1; i++ {
			in := test[i : i+blocksize]
			if i == 0 {
				rolling.Write(in)
			}

			slow := checksum(in)
			if got := rolling.Sum64(); got != slow {
				t.Errorf("rolling hash: %q = %d want %d", in, got, slow)
			}

			rolling.Roll(uint64(blocksize), test[i], test[i+blocksize])
		}
		rolling.Reset()
	}
}
