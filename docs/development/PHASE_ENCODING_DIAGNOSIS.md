# Phase Encoding Diagnosis - Complete Analysis

**Date**: December 16, 2025
**Status**: ❌ Pairwise Differential Approach Failed
**Conclusion**: Phase encoding in WAV files requires time-domain techniques

---

## Problem Statement

Absolute phase encoding fails completely after WAV file processing (IFFT → int32 quantization → save → load → FFT). The phase values become essentially random noise.

## Root Cause Analysis

### Isolated FFT Test (✅ WORKS)
```go
// Direct FFT → modify phase → IFFT → FFT cycle
// NO WAV processing, NO quantization
spectrum[10] = cmplx.Rect(mag, +π/2)  // Embed bit=1
// After IFFT → FFT:
phase = +π/2  ✅ EXACTLY PRESERVED
```

### WAV Round-Trip Test (❌ FAILS)
```go
// FFT → modify phase → IFFT → int32 → WAV save → WAV load → int32→float64 → FFT
spectrum[10] = cmplx.Rect(mag, +π/2)  // Embed bit=1
// After WAV cycle:
phase = 2.763 (random)  ❌ COMPLETELY DESTROYED
```

**Conclusion**: The corruption happens during the **time-domain quantization** step (float64 → int32 conversion).

---

## Attempted Solutions

### Solution 1: Absolute Phase Encoding ❌ FAILED
- **Method**: Set phase[freq] = +3π/4 for bit=1, -3π/4 for bit=0
- **Extraction**: Compare phase > 0
- **Result**: Phases become random after WAV processing (2.763, -1.487, -0.627)
- **Failure Mode**: Quantization destroys absolute phase values

### Solution 2: Pairwise Differential Phase ❌ FAILED
- **Method**: Use adjacent frequencies as pairs
  - Bit 0: phase[freq] - phase[freq+1] = +π/2
  - Bit 1: phase[freq] - phase[freq+1] = -π/2
- **Extraction**: Compare phase difference > 0
- **Hypothesis**: Quantization errors on adjacent frequencies should be correlated
- **Result**: Phase differences are still random after WAV
  - Embedded: phase_diff = +1.571 (π/2) ✓
  - Extracted: phase_diff = 2.637, -1.844, -0.167 (random) ✗
- **Failure Mode**: Quantization errors are NOT correlated between frequencies

---

## Why Phase Encoding Fails in WAV Files

### The Quantization Problem

```
Time Domain Signal → FFT → Frequency Domain Phase
  |                                    ↑
  | int32 quantization                 |
  ↓ (LOSSY)                           |
int32 WAV samples → float64 → FFT ───┘
                    (NON-LINEAR CORRUPTION)
```

**Key insight**: Small time-domain quantization errors create LARGE, NON-LINEAR phase errors in frequency domain because:

1. **Magnitude vs Phase sensitivity**: FFT phase is extremely sensitive to time-domain perturbations
2. **Non-uniform corruption**: Different frequencies experience independent quantization effects
3. **No correlation**: Adjacent frequencies don't have correlated phase errors

### What Survives WAV Processing

| Feature | Survives? | Why? |
|---------|-----------|------|
| **Magnitude** | ✅ YES | Quantization preserves energy distribution |
| **Time-domain delays** | ✅ YES | Integer sample delays preserved exactly |
| **Autocorrelation** | ✅ YES | Statistical time-domain property |
| **Absolute phase** | ❌ NO | Extremely sensitive to quantization |
| **Phase differences** | ❌ NO | No correlation between freq bins |
| **Spectral envelope** | ✅ YES | High-level frequency structure |

---

## The Correct Approach: Time-Domain Phase Encoding

Phase encoding CAN work in WAV files if we encode information in **phase-DERIVED time-domain features** rather than raw phase values.

### Method: Phase-Modulated Group Delay

**Concept**: Use phase gradients to create time-domain shifts that survive quantization.

```go
// Embedding
for freq := minFreq; freq < maxFreq; freq++ {
    if bit == 0 {
        // Create upward phase ramp (positive group delay → time advance)
        phase[freq] = basePhase + α * freq
    } else {
        // Create downward phase ramp (negative group delay → time delay)
        phase[freq] = basePhase - α * freq
    }
}
// This creates a time-domain chirp/sweep
```

**Extraction**: Detect the time-domain chirp direction using autocorrelation or envelope analysis.

**Why this works**:
1. Phase gradients create **time-domain chirps** (frequency sweeps)
2. Chirp direction is a TIME-DOMAIN feature
3. Time-domain features survive int32 quantization
4. Detection uses robust autocorrelation (like echo hiding)

---

## Comparison with Echo Hiding

| Aspect | Echo Hiding (✅ WORKS) | Phase Encoding (❌ FAILS) |
|--------|----------------------|------------------------|
| **Domain** | Time (sample delays) | Frequency (phase values) |
| **Encoding** | Add delayed copies | Modify phase spectrum |
| **Quantization** | Delays preserved | Phase destroyed |
| **Detection** | Autocorrelation peaks | Phase comparison |
| **Robustness** | High (integer delays) | Low (fractional phases) |

**Lesson**: For WAV files, encode in TIME domain, detect in TIME domain. Don't cross domains through quantization.

---

## Recommended Implementation

### Option A: Hybrid Phase-Echo Technique
Combine frequency-domain analysis with time-domain encoding:

1. **Embedding**:
   - FFT to find dominant frequency band
   - Add frequency-selective echo to that band
   - Creates phase-modulated time-domain pattern

2. **Extraction**:
   - FFT to identify modified band
   - Autocorrelation detection within that band
   - Robust to WAV quantization

### Option B: Group Delay Encoding
Use linear phase gradients to create time shifts:

1. **Embedding**:
   - Bit 0: Positive phase slope → time advance
   - Bit 1: Negative phase slope → time delay

2. **Extraction**:
   - Measure time-domain envelope shift
   - Compare to baseline segment

### Option C: Magnitude-Phase Coupling
Encode data in the relationship between magnitude and phase:

1. **Embedding**:
   - Bit 0: Magnitude peak where phase = 0 (constructive)
   - Bit 1: Magnitude peak where phase = π (destructive)

2. **Extraction**:
   - Detect interference pattern in time domain
   - Robust because it uses magnitude + phase

---

## Conclusion

**Phase encoding in WAV files is fundamentally incompatible with direct phase value encoding.** The int32 quantization step destroys phase information through non-linear, frequency-dependent corruption.

**The solution**: Use phase to CREATE time-domain features (chirps, delays, interference patterns), then detect those time-domain features using robust methods like autocorrelation.

This is why echo hiding works - it's a time-domain technique. Phase encoding must also become a time-domain technique to work with WAV files.

---

## Next Steps

1. ✅ Document pairwise differential failure (this document)
2. ⏳ Implement Option A (hybrid phase-echo) or Option B (group delay)
3. ⏳ Test with real WAV files
4. ⏳ Compare with echo hiding performance
5. ⏳ Update architecture documentation

**Status**: Pivoting to time-domain phase encoding approach
