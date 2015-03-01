/*
Package minhash provides probabilistic data structures for computing
MinHash signatures for streaming data.

The reader should also conf https://github.com/dgryski/go-minhash and
https://github.com/tylertreat/BoomFilters.  In fact, most of the
implementation details of this package are based off the former.

MinHash signatures can be used to estimate the Jaccard index
J(A, B) := |A & B| / |A || B| of two sets that are subsets
of some finite, totally ordered set U.  If s is a permutation
of U chosen uniformly at random, then x := argmin s(A || B) is
just a random element chosen uniformly from A || B.  It's
clear that P(x in A & B) = J(A, B).  Since min s(A) = min s(B)
if and only if x is in A & B, we have just shown that
P(min s(A) = min S(B)) = J(A, B).

The central idea of minhash signatures is to repeatedly perform
the above experiment with different permutations as a way to estimate
the underlying probability
J(A, B) = P(an element x in A || B is also  in A & B).

A length k minhash signature S(A) is theoretically generated by randomly
choosing k permutations si (i=1, ..., k) in the symmetric group of U
(group of bijective endofunctions on U) and computing  hi(A) := min si(A)
for each permutation.  We take S(A) := [h1(A), ..., hk(A)].
Since each permutation is a bijection, min si(A) = min si(B) if and
only if argmin si(A) = argmin si(B) and so we could just as
well use these argmins, which is sometimes how the signature S(A) is
defined.

Specifying permutations for large U is not efficient, and so we often
take a family of integer-valued hash functions that are minwise
independent, in the sense that for most sets A,
min h(A) ! = min g(A) for two distinct hash functions in the family.
Frequently this family is parametrically  generated.

For more information,
    http://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/

    MinHashing:
    http://infolab.stanford.edu/~ullman/mmds/ch3.pdf
    https://en.wikipedia.org/wiki/MinHash

    BottomK:
    http://www.math.tau.ac.il/~haimk/papers/p225-cohen.pdf
    http://cohenwang.org/edith/Papers/metrics394-cohen.pdf

    http://www.mpi-inf.mpg.de/~rgemulla/publications/beyer07distinct.pdf

This package works best when provided with a strong 64-bit hash function,
such as CityHash, Spooky, Murmur3, or SipHash.

*/

package minhash

import (
	"encoding/binary"
	"log"
	"math"
	"strconv"
)

const (
	MaxUint  uint   = ^uint(0)
	MaxInt   int    = int(MaxUint >> 1)
	infinity uint64 = math.MaxUint64
)

type HashFunc func([]byte) uint64

// Signature is an array representing the signature of a set.
type Signature []uint64

// MinHash is an a probabilistic data structure used to
// compute a similarity preserving signature for a set.  It ingests
// a stream of the set's elements and continuously updates the signature.
type MinHash interface {
	// Push ingests a set element, hashes it, and updates the signature.
	Push(interface{})

	// Merge updates the signature of the instance with the signature
	// of the input.  This results in the signature of the union of the
	// two sets, which is stored in the original MinHash instance.
	Merge(*MinHash)

	// Cardinality estimates the size of the set from the signature.
	Cardinality() int

	// Signature returns the signature itself.
	Signature() []uint64

	// Similarity computes the similarity between two MinHash signatures.
	// The method for computing similarity depends on whether a MinWise
	// or Bottom-K implementation is used.
	Similarity(*MinHash) float64
}

// Similarity invokes the specific
func Similarity(m1, m2 *MinHash) float64 {
	return (*m1).Similarity(m2) // dereference m1
}

// defaultSignature will return an appropriately typed array
func defaultSignature(size int) Signature {
	s := make(Signature, size)
	for i := range s {
		s[i] = infinity
	}
	return s
}

// toBytes converts various types to a byte slice
// so they can be pushed into a MinHash instance.
func toBytes(x interface{}) []byte {
	b := make([]byte, 8)
	switch t := x.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	case uint, uint16, uint32, uint64:
		binary.LittleEndian.PutUint64(b, t.(uint64))
	case int, int16, int32, int64:
		binary.LittleEndian.PutUint64(b, uint64(t.(int64)))
	}
	return b
}

// stringIntToByte converts a string representation of an integer to a byte slice.
func stringIntToByte(s string) []byte {
	n, err := strconv.ParseUint(s, 0, 64)
	var b []byte
	if err != nil {
		log.Println("Could not convert string to uint64.")
		b = []byte(s)
	} else {
		b = toBytes(n)
	}
	return b
}

// Intersection estimates the cardinality of the intersection
// between two sets provided their sizes and Jaccard similarity are known.
//
// If n, m, i, u denote |A|, |B|, |A & B|, and |A || B| respectively,
// then given that A and B are not disjoint (i != 0), w have
// J := J(A, B) = i / u which is equivalent to
// 1/J = u / i = (n + m - i) / i = (n + m / i) - 1.
// Solving for i yields |A & B| = (n + m) / ((1/J)+ 1).
// Thus an estimate for the Jaccard index of A and B yields an estimate
// for the size of their intersection, provided we know the sizes of A and B.
func Intersection(js float64, size1, size2 int) int {
	var est int
	if js == 0 {
		est = 0
	} else {
		est = int(math.Floor(float64(size1+size2) / ((1.0 / js) + 1)))
	}
	return est
}
