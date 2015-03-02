Package minhash provides probabilistic data structures for computing
MinHash signatures for streaming data.

The reader should also confer https://github.com/dgryski/go-minhash and
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

For more information:

**MinHashing**
-  http://research.neustar.biz/2012/07/09/sketch-of-the-day-k-minimum-values/
-  http://infolab.stanford.edu/~ullman/mmds/ch3.pdf
-  https://en.wikipedia.org/wiki/MinHash

**BottomK**
-  http://www.math.tau.ac.il/~haimk/papers/p225-cohen.pdf
-  http://cohenwang.org/edith/Papers/metrics394-cohen.pdf
- http://www.mpi-inf.mpg.de/~rgemulla/publications/beyer07distinct.pdf

This package works best when provided with a strong 64-bit hash function,
such as CityHash, Spooky, Murmur3, or SipHash.


### MinWise

The **MinWise** data structure computes the MinHash for a set by
creating a parametric family of hash functions.  It is initialized
with two hash functions and an integer size parameter.  The two hash
functions h1 and h2 generate the family h1 + i\*h2 for i=1, ..., n,
where n is the size of the signature we wish to compute.

#### Usage

Let's explore an example of ingesting a stream of data, sketching a signature,
and computing the similarity between signatures.

```go
package main

import (
	"log"

	"github.com/dgryski/go-farm"
	"github.com/dgryski/go-spooky"
	"github.com/shawnohare/go-minhash"
)

func main() {
	// Some pre-existing sets to compute the signatures of.
	S1 := []string{"5", "1", "2", "3", "4"}
	S2 := []int{1, 2, 3, 4, 5}
	words := []string{"idempotent", "condensation", "is", "good"}

	// Specify two hash functions to initialize all our MinWise instances
  // with.  These two hash functions generate a parametric family
  // of hash functions of cardinality = size.
	h1 := spooky.Hash64
	h2 := farm.Hash64
	size := 3

	// Init a MinWise instance for each set above.
	wmw := minhash.NewMinWise(h1, h2, size) // handle words set
  mw1 := minhash.NewMinWise(h1, h2, size) // handle S1
	mw2 := minhash.NewMinWise(h1, h2, size) // handle S2

	// Ingest the words set one element at a time.
	for _, w := range words {
		wmw.Push(w)
	}


	// Repeat the above, but with string integer data S1 and integer data S2.
	
  // Ingest S1.
	for _, x := range S1 {
		mw1.PushStringInt(x) // Note the different push function.
	}

  // Ingest S2.
	for _, x := range S2 {
		mw2.Push(x) // we use Push for both integers and word strings.
	}
  
  // NOTE In the above we used the Push and PushStringInt methods.
  // If the data is already represented as a set of []byte elements,
  // then the PushBytes function is slightly more efficient.
  
	// Comparing signatures.
	var s float64
	// Using a helper function that accepts MinHash interfaces.
	s = minhash.Similarity(mw1, mw2)
	// or if we wish, we can call the MinWise method directly.
	// s = mw1.Similarity(mw2)
	log.Println("Similarity between signatures of S1 and S2:", s)

	// Output signatures for potential storage. 
  var wordSig []uint64 
  var sig1 []uint64 
  var sig2 []uint64 
  wordsSig = wmw.Signature()
	sig1 = mw1.Signature()
	sig2 = mw2.Signature()
	log.Println("Signature for words set:", wordsSig)
  
  // Computing similarity of signatures directly.
  // Suppose we store sig1 and sig2 above and retrieve them as []int.
	// We can directly compare the similarities as follows.
	usig1 := make([]uint64, len(sig1))
	usig2 := make([]uint64, len(sig2))
	// First, convert to []uint64
	for i, v := range sig1 {
		usig1[i] = uint64(v)
		usig2[i] = uint64(sig2[i])
	}
	// Calculate similarities using the now appropriately typed signatures.
	simFromSigs := minhash.MinWiseSimilarity(usig1, usig2)
	log.Println("Similarity calcuted directly from signatures: ", simFromSigs)
  
  // Construct a MinWise instance from a signature.
	// If we want to continue to stream elements into the set represented by
	// sig1, we can convert it into a MinWise instance via
	m := minhash.NewMinWiseFromSignature(h1, h2, usig1)
	// We can now stream in more elements an update the signature.
	for i := 20; i <= 50; i++ {
		m.Push(i)
	}
	// Calculate new similarity between updated S1 and S2
	newSim := m.Similarity(mw2)
	log.Println("New similarity between signatures for S1 and S2:", newSim)
}
```
