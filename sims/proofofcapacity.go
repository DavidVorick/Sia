package sims

import (
	"fmt"
	"math"

	"github.com/NebulousLabs/Sia/siacrypto"
)

// bucketSim helps try to answer the question of variance between results when
// trying to determine what percentage of a data set someone is storing.
func bucketSim() {
	fmt.Println("Bucket Variance Sim")

	// Create N buckets and fill them with M objects.
	n := 1000000
	m := 1000
	buckets0 := make([][]struct{}, n)
	for i := 0; i < m; i++ {
		// Put the object in a random bucket.
		index := siacrypto.RandomInt(n)
		buckets0[index] = append(buckets0[index], struct{}{})
		// fmt.Println("Object in bucket", index)
	}

	// Find the closest object at the back --> makes sim more efficient.
	backClosest := 0
	for i := n - 1; backClosest == 0 && i != 0; i-- {
		if len(buckets0[i]) != 0 {
			backClosest = i
		}
	}
	// fmt.Println("backClosest:", backClosest)

	// Find the front closest object.
	frontClosest := n
	for i := 0; frontClosest == n && i != n; i++ {
		if len(buckets0[i]) != 0 {
			frontClosest = i
		}
	}
	// fmt.Println("frontClosest:", frontClosest)

	// Mean can be assumed (n / (2*m)) for large datasets.
	mean := float64(n) / (2 * float64(m))
	fmt.Println("Assumed Mean:", mean)

	// Find the variance of the closest object.
	sumVariance := float64(0)
	for i := 0; i < n; i++ {
		frontDistance := frontClosest - i
		if frontDistance < 0 {
			frontDistance = i - frontClosest
		}

		backDistance := i - backClosest
		if backDistance < 0 {
			backDistance = backClosest - i
		}

		if frontDistance < backDistance {
			sumVariance += (float64(frontDistance) - mean) * (float64(frontDistance) - mean)
		} else {
			sumVariance += (float64(backDistance) - mean) * (float64(backDistance) - mean)
		}

		// Update frontClosest and backClosest if the closest object is the current bucket.
		if frontDistance == 0 {
			// After we step forward, the back closest will be this index.
			backClosest = i

			// Search forward for the next front closest.
			for j := i + 1; frontClosest == i && j != i; j++ {
				// Wrap around to the front if needed.
				if j == n {
					j = 0
				}

				if len(buckets0[j]) != 0 {
					frontClosest = j
				}
			}
			// fmt.Println("New front closest:", frontClosest)
		}
	}

	variance := sumVariance / float64(n)
	sd := math.Sqrt(variance)
	// fmt.Println("Variance:", variance)
	fmt.Println("Standard Deviation:", sd)


	sigma := float64(4)
	// Calculate the bad-luck data represented on 1 trial.
	rm0 := float64(n) / (2 * ((sigma * mean)+mean))
	bld0 := rm0 / float64(m)
	fmt.Println(sigma, "Sigma Bad Luck, 1 Trial:", bld0)

	// Calculate the bad-luck data represented on 10 trials.
	rm1 := float64(n) / (2 * ((sigma * mean / math.Sqrt(10)+mean)))
	bld1 := rm1 / float64(m)
	fmt.Println(sigma, "Sigma Bad Luck, 10 Trial:", bld1)

	// Calculate the bad-luck data represented on 100 trials.
	rm2 := float64(n) / (2 * ((sigma * mean / math.Sqrt(100)+mean)))
	bld2 := rm2 / float64(m)
	fmt.Println(sigma, "Sigma Bad Luck, 10 Trial:", bld2)
}
