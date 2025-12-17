package main

import (
	"fmt"
	"math"
	"math/cmplx"
	"math/rand"

	"gonum.org/v1/gonum/dsp/fourier"
)

func main() {
	// Create test data
	testBits := []bool{true, false, true, true, false, false, true, false}

	rng := rand.New(rand.NewSource(42))
	segmentSize := 1024
	segment := make([]float64, segmentSize)
	for i := range segment {
		segment[i] = (rng.Float64()*2 - 1) * 0.7 // White noise
	}

	fft := fourier.NewFFT(segmentSize)
	freqDomain := fft.Coefficients(nil, segment)

	fmt.Println("=== EMBEDDING ===")
	minFreq := 10

	for i, bit := range testBits {
		freq := minFreq + i
		magnitude := cmplx.Abs(freqDomain[freq])
		origPhase := cmplx.Phase(freqDomain[freq])

		// Set phase to specific value
		var newPhase float64
		if bit {
			newPhase = math.Pi / 2 // +π/2 for bit 1
		} else {
			newPhase = -math.Pi / 2 // -π/2 for bit 0
		}

		freqDomain[freq] = cmplx.Rect(magnitude, newPhase)

		fmt.Printf("Freq %d: bit=%v, orig_phase=%.4f, new_phase=%.4f\n",
			freq, bit, origPhase, newPhase)
	}

	// Apply IFFT to create modified audio
	modifiedSegment := fft.Sequence(nil, freqDomain)

	// Now extract - apply FFT again
	extractFFT := fourier.NewFFT(segmentSize)
	extractFreqDomain := extractFFT.Coefficients(nil, modifiedSegment)

	fmt.Println("\n=== EXTRACTING ===")
	var extractedBits []bool
	for i := 0; i < len(testBits); i++ {
		freq := minFreq + i
		phase := cmplx.Phase(extractFreqDomain[freq])

		// Normalize to [0, 2π]
		normalizedPhase := phase
		if normalizedPhase < 0 {
			normalizedPhase += 2 * math.Pi
		}

		// Bit 1 if phase > 0, bit 0 if phase <= 0
		extractedBit := phase > 0

		extractedBits = append(extractedBits, extractedBit)

		fmt.Printf("Freq %d: phase=%.4f, normalized=%.4f, extracted_bit=%v, expected=%v ✓=%v\n",
			freq, phase, normalizedPhase, extractedBit, testBits[i], extractedBit == testBits[i])
	}

	// Check if all bits match
	allMatch := true
	for i := range testBits {
		if testBits[i] != extractedBits[i] {
			allMatch = false
			break
		}
	}

	fmt.Printf("\n=== RESULT: ")
	if allMatch {
		fmt.Println("✅ ALL BITS MATCH!")
	} else {
		fmt.Println("❌ BITS DON'T MATCH")
	}
}
