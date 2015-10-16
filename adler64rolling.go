// This is practically entirely a copy of the hash/adler32/adler32.go
// except it also includes a Roll() function that allows to pop an old
// byte and push a new byte, giving a rolling hash method.

package adler64rolling

import "hash"

const (
	// mod is the largest prime (that I could find) that is less than 2^64.
	mod = 0xfffffffb

	// nmax is the largest hash such that
	// 255 * n * (n+1) / 2 + (n+1) * (mod-1) < 2^64-1.
	// It is mentioned in RFC 1950 (search for "5552").
	// 2^28 + 2^26 + 2^24 + 2^23 + 2^21 + 2^20 + 2^15 + 2^13 + 2^10 + 2^9 + 2^5 + 2^3 + 2^2 + 3
	nmax = 363898415
)

// Size of the checksum in bytes.
const Size = 8

type digest uint64

func (d *digest) Reset() { *d = 0 }

type Hash64 interface {
	hash.Hash64
	Roll(blocksize uint64, p byte, r byte)
}

func New() Hash64 {
	d := new(digest)
	d.Reset()
	return d
}

func (d *digest) Size() int { return Size }

func (d *digest) BlockSize() int { return 1 }

// Add p to the running checksum d.
func update(d digest, p []byte) digest {
	s1, s2 := uint64(d&0xffffffff), uint64(d>>32)
	for len(p) > 0 {
		var q []byte
		if len(p) > nmax {
			p, q = p[:nmax], p[nmax:]
		}
		for _, x := range p {
			s1 += uint64(x)
			s2 += s1
		}
		s1 %= mod
		s2 %= mod
		p = q
	}
	return digest(s2<<32 | s1)
}

func (d *digest) Write(p []byte) (nn int, err error) {
	*d = update(*d, p)
	return len(p), nil
}

// Add p to the running checksum d while removing r.
func roll(d digest, blocksize uint64, del byte, add byte) digest {
	s1, s2 := uint64(d&0xffffffff), uint64(d>>32)

	addi := uint64(add)
	deli := uint64(del)

	/*
		s1 -= o
		s2 -= blocksize * o

		s1 += i
		s2 += s1
	*/

	s1 += addi - deli
	s2 += s1 - blocksize*deli

	s1 %= mod
	s2 %= mod

	return digest(s2<<32 | s1)
}

func (d *digest) Roll(blocksize uint64, del byte, add byte) {
	*d = roll(*d, blocksize, del, add)
}

func (d *digest) Sum64() uint64 { return uint64(*d) }

func (d *digest) Sum(in []byte) []byte {
	s := uint32(*d)
	return append(in, byte(s>>24), byte(s>>16), byte(s>>8), byte(s))
}

func Checksum(data []byte) uint64 { return uint64(update(1, data)) }
