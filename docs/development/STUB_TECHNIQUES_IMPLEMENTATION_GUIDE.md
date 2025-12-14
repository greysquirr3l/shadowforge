# Stub Techniques Implementation Guide

> **Status**: Phase Encoding and Echo Hiding are currently stub implementations
> **Created**: December 13, 2025
> **Purpose**: Roadmap for completing audio steganography techniques

---

## Overview

Shadowforge currently has **2 stub audio steganography techniques** that require full implementation:

1. **Phase Encoding** (248 lines, 10 tests, ❌ round-trip FAILS)
2. **Echo Hiding** (285 lines, 10 tests, ❌ round-trip FAILS)

Both techniques have:
- ✅ Complete configuration structures
- ✅ Basic API interfaces implemented
- ✅ Comprehensive test scaffolding
- ❌ **Missing**: Actual signal processing algorithms (FFT/IFFT for Phase, Autocorrelation for Echo)

**Why stubs exist**: These techniques require complex digital signal processing (DSP) libraries that need careful evaluation before integration into the production codebase.

---

## 1. Phase Encoding Technique

### Current Implementation Status

**File**: `internal/infrastructure/stego/phase.go` (265 lines)
**Tests**: `internal/infrastructure/stego/phase_test.go` (10 tests)

**What Works:**
- ✅ Configuration structure (`PhaseEmbeddingConfig`)
- ✅ Format validation (WAV files only)
- ✅ API interface compliance (`Embed`, `Extract`, `CalculateCapacity`)
- ✅ Logging integration

**What's Missing:**
```go
// TODO: Implement phase encoding (line 85)
// 1. Parse WAV file to get audio samples
// 2. Convert payload to bits
// 3. Divide audio into overlapping segments
// 4. For each segment: FFT -> modify phase -> IFFT
// 5. Reconstruct audio with embedded data

// TODO: Implement phase decoding (line 115)
// 1. Parse audio segments
// 2. Extract phase modifications
// 3. Decode bits from phase differences
// 4. Reconstruct payload

// TODO: Calculate actual capacity (line 149)
// Based on audio parameters (sample rate, channels, bit depth)
```

### Mathematical Foundation

**Phase Encoding Principle:**
- Human ear is sensitive to magnitude spectrum, less sensitive to phase spectrum
- Modify phase while preserving magnitude → imperceptible changes

**Required Mathematics:**

1. **Discrete Fourier Transform (DFT)**:
   ```
   X[k] = Σ(n=0 to N-1) x[n] * e^(-j*2π*k*n/N)
   ```

2. **Polar Form**:
   ```
   X[k] = |X[k]| * e^(jφ[k])
   where |X[k]| = magnitude, φ[k] = phase
   ```

3. **Phase Modification**:
   ```
   φ'[k] = φ[k] + Δφ[k]
   where Δφ[k] = bit_value * π/8
   ```

4. **Inverse DFT**:
   ```
   x'[n] = (1/N) * Σ(k=0 to N-1) X'[k] * e^(j*2π*k*n/N)
   ```

### Configuration Parameters

```go
type PhaseEmbeddingConfig struct {
    SegmentSize    int     // Default: 1024 samples
    PhaseThreshold float64 // Default: π/8 (22.5°)
    MinFrequency   int     // Default: 10 bins
    MaxFrequency   int     // Default: 512 bins
    OverlapFactor  float64 // Default: 0.5 (50% overlap)
}
```

**Parameter Guidelines:**
- **SegmentSize**: Trade-off between capacity and imperceptibility (512-2048)
- **PhaseThreshold**: Maximum phase shift per bit (π/16 to π/4)
- **MinFrequency**: Skip low frequencies to avoid audible artifacts (5-20 bins)
- **MaxFrequency**: Limit to mid-range frequencies (256-1024 bins)
- **OverlapFactor**: Overlap prevents discontinuities (0.25-0.75)

### Required FFT Library

**Candidates:**

1. **gonum/fourier** (✅ Recommended)
   - Package: `gonum.org/v1/gonum/dsp/fourier`
   - Pros: Well-maintained, part of gonum ecosystem, comprehensive
   - Cons: Larger dependency tree
   - License: BSD-3-Clause (✅ Compatible)

2. **mjibson/go-dsp/fft**
   - Package: `github.com/mjibson/go-dsp/fft`
   - Pros: Lightweight, simple API
   - Cons: Less maintained, smaller community
   - License: MIT (✅ Compatible)

3. **mjibson/go-dsp** (Full DSP suite)
   - Package: `github.com/mjibson/go-dsp`
   - Pros: Includes FFT, filters, windows
   - Cons: Less maintained
   - License: MIT (✅ Compatible)

**Recommendation**: Use **gonum/fourier** for production quality.

### Implementation Roadmap

**Phase 1: FFT Integration (2-3 days)**
```bash
go get gonum.org/v1/gonum/dsp/fourier
```

```go
import "gonum.org/v1/gonum/dsp/fourier"

// Example usage:
fft := fourier.NewFFT(segmentSize)
coefficients := fft.Coefficients(nil, audioSegment)
// Modify phase spectrum...
reconstructed := fft.Sequence(nil, modifiedCoefficients)
```

**Phase 2: Embedding Implementation (3-4 days)**
1. Parse WAV file with existing `internal/infrastructure/media/audio_service.go`
2. Convert payload to bit stream
3. Segment audio with overlap
4. Apply FFT to each segment
5. Modify phase spectrum based on bits
6. Apply IFFT to reconstruct
7. Reassemble audio file

**Phase 3: Extraction Implementation (2-3 days)**
1. Segment stego audio
2. Apply FFT to extract phase spectrum
3. Decode bits from phase differences
4. Reconstruct payload
5. Verify integrity

**Phase 4: Testing & Optimization (2-3 days)**
1. Round-trip tests with all 10 existing test cases
2. Imperceptibility testing (SNR analysis)
3. Capacity optimization
4. Performance benchmarking

**Total Estimate**: 9-13 days for full implementation

### Test Strategy

**Existing Test Cases** (phase_test.go):
- ✅ Default configuration
- ✅ Custom configuration
- ✅ Format support validation
- ✅ Empty payload handling
- ✅ Insufficient capacity detection
- ✅ Name() identifier
- ✅ Basic embedding
- ✅ Basic extraction
- ✅ Capacity calculation
- ✅ Context cancellation

**Additional Tests Needed:**
- Round-trip integrity (embed → extract → verify)
- Imperceptibility (signal-to-noise ratio > 40dB)
- Robustness (MP3 compression, resampling)
- Edge cases (very short audio, mono vs stereo)
- Performance (1MB audio should process < 500ms)

### Code Example (Full Implementation)

```go
func (p *PhaseTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
    // 1. Parse WAV file
    wavData, err := parseWAV(carrier)
    if err != nil {
        return nil, fmt.Errorf("parse WAV: %w", err)
    }

    // 2. Convert payload to bits
    bits := bytesToBits(payload)
    if len(bits) > p.calculateMaxBits(wavData) {
        return nil, stego.ErrInsufficientCapacity
    }

    // 3. Initialize FFT
    fft := fourier.NewFFT(p.config.SegmentSize)

    // 4. Segment and embed
    samples := wavData.Samples
    segmentCount := len(samples) / p.config.SegmentSize
    bitIndex := 0

    for i := 0; i < segmentCount && bitIndex < len(bits); i++ {
        // Extract segment with overlap
        start := i * int(float64(p.config.SegmentSize) * (1 - p.config.OverlapFactor))
        end := start + p.config.SegmentSize
        if end > len(samples) {
            end = len(samples)
        }
        segment := samples[start:end]

        // FFT
        coeffs := fft.Coefficients(nil, segment)

        // Modify phase spectrum
        for k := p.config.MinFrequency; k < p.config.MaxFrequency && bitIndex < len(bits); k++ {
            magnitude := cmplx.Abs(coeffs[k])
            phase := cmplx.Phase(coeffs[k])

            // Embed bit in phase
            if bits[bitIndex] == 1 {
                phase += p.config.PhaseThreshold
            } else {
                phase -= p.config.PhaseThreshold
            }

            // Reconstruct complex number
            coeffs[k] = cmplx.Rect(magnitude, phase)
            bitIndex++
        }

        // IFFT
        modifiedSegment := fft.Sequence(nil, coeffs)
        copy(samples[start:end], modifiedSegment)
    }

    // 5. Reconstruct WAV
    return buildWAV(wavData), nil
}
```

---

## 2. Echo Hiding Technique

### Current Implementation Status

**File**: `internal/infrastructure/stego/echo.go` (318 lines)
**Tests**: `internal/infrastructure/stego/echo_test.go` (10 tests)

**What Works:**
- ✅ Configuration structure (`EchoEmbeddingConfig`)
- ✅ Format validation (WAV files only)
- ✅ API interface compliance
- ✅ Logging integration

**What's Missing:**
```go
// TODO: Implement echo hiding (line 83)
// 1. Parse WAV file to get audio samples
// 2. Convert payload to bits
// 3. Divide audio into segments
// 4. For each segment and bit: add echo with appropriate delay
// 5. Reconstruct audio with embedded echoes

// TODO: Implement echo detection (line 117)
// 1. Parse audio segments
// 2. Detect echo delays using autocorrelation
// 3. Classify delays as bit 0 or bit 1
// 4. Reconstruct payload from detected bits

// TODO: Calculate actual capacity (line 171)
// Based on segment length and audio duration
```

### Mathematical Foundation

**Echo Hiding Principle:**
- Add imperceptible echoes with different delays to represent binary data
- Delay difference (~1 sample at 44.1kHz ≈ 0.023ms) is below human perception threshold

**Required Mathematics:**

1. **Echo Addition**:
   ```
   y[n] = α * x[n] + (1-α) * x[n-d]
   where d = delay (Delay0 or Delay1)
         α = mix ratio (0.8 default)
   ```

2. **Autocorrelation Function**:
   ```
   R[τ] = Σ(n=0 to N-1) x[n] * x[n+τ]
   ```

3. **Echo Detection**:
   ```
   Peak detection in R[τ] around expected delays
   If peak near Delay0 → bit = 0
   If peak near Delay1 → bit = 1
   ```

### Configuration Parameters

```go
type EchoEmbeddingConfig struct {
    Delay0     int     // Default: 100 samples (~2.3ms @ 44.1kHz)
    Delay1     int     // Default: 101 samples (~2.3ms @ 44.1kHz)
    Amplitude  float64 // Default: 0.1 (10% echo strength)
    SegmentLen int     // Default: 8192 samples (~185ms)
    MixRatio   float64 // Default: 0.8 (80% original, 20% echo)
}
```

**Parameter Guidelines:**
- **Delay0/Delay1**: Should differ by 1-2 samples (below perception threshold)
- **Amplitude**: 0.05-0.15 (too high = audible, too low = undetectable)
- **SegmentLen**: 4096-16384 (trade-off: capacity vs robustness)
- **MixRatio**: 0.7-0.9 (preserve original audio quality)

### Required DSP Functions

**Autocorrelation Implementation Options:**

1. **Custom Implementation**:
   ```go
   func autocorrelation(signal []float64, maxLag int) []float64 {
       result := make([]float64, maxLag+1)
       n := len(signal)

       for lag := 0; lag <= maxLag; lag++ {
           sum := 0.0
           for i := 0; i < n-lag; i++ {
               sum += signal[i] * signal[i+lag]
           }
           result[lag] = sum / float64(n-lag)
       }

       return result
   }
   ```

2. **FFT-based (faster)**:
   ```go
   func autocorrelationFFT(signal []float64) []float64 {
       fft := fourier.NewFFT(len(signal))

       // FFT of signal
       coeffs := fft.Coefficients(nil, signal)

       // Power spectrum (|X[k]|²)
       for i := range coeffs {
           coeffs[i] = coeffs[i] * cmplx.Conj(coeffs[i])
       }

       // IFFT to get autocorrelation
       return fft.Sequence(nil, coeffs)
   }
   ```

**Recommendation**: Use custom implementation for clarity, optimize with FFT if performance issues arise.

### Implementation Roadmap

**Phase 1: Echo Addition (2 days)**
1. Parse WAV with existing audio service
2. Convert payload to bit stream
3. Segment audio
4. Add echoes based on bit values
5. Reconstruct WAV

**Phase 2: Autocorrelation Implementation (2 days)**
1. Implement autocorrelation function
2. Add peak detection algorithm
3. Test with synthetic echoes

**Phase 3: Echo Detection (2-3 days)**
1. Segment stego audio
2. Apply autocorrelation to each segment
3. Detect echo delays
4. Classify as bit 0 or 1
5. Reconstruct payload

**Phase 4: Testing & Optimization (2 days)**
1. Round-trip tests with all 10 existing test cases
2. Imperceptibility testing
3. Robustness to noise/compression
4. Performance benchmarking

**Total Estimate**: 8-9 days for full implementation

### Test Strategy

**Existing Test Cases** (echo_test.go):
- ✅ Default configuration
- ✅ Custom configuration
- ✅ Format support validation
- ✅ Empty payload handling
- ✅ Insufficient capacity detection
- ✅ Name() identifier
- ✅ Basic embedding
- ✅ Basic extraction
- ✅ Capacity calculation
- ✅ Context cancellation

**Additional Tests Needed:**
- Round-trip integrity verification
- Imperceptibility (SNR > 35dB)
- Robustness (noise, MP3 compression)
- Edge cases (very short audio, silence)
- Performance (1MB audio < 300ms)

### Code Example (Full Implementation)

```go
func (e *EchoTechnique) Embed(ctx context.Context, carrier, payload []byte) ([]byte, error) {
    // 1. Parse WAV
    wavData, err := parseWAV(carrier)
    if err != nil {
        return nil, fmt.Errorf("parse WAV: %w", err)
    }

    // 2. Convert payload to bits
    bits := bytesToBits(payload)
    if len(bits) > e.calculateMaxBits(wavData) {
        return nil, stego.ErrInsufficientCapacity
    }

    // 3. Segment and embed
    samples := wavData.Samples
    segmentCount := len(samples) / e.config.SegmentLen
    bitIndex := 0

    for i := 0; i < segmentCount && bitIndex < len(bits); i++ {
        start := i * e.config.SegmentLen
        end := start + e.config.SegmentLen
        if end > len(samples) {
            end = len(samples)
        }
        segment := samples[start:end]

        // Choose delay based on bit value
        delay := e.config.Delay0
        if bits[bitIndex] == 1 {
            delay = e.config.Delay1
        }

        // Add echo: y[n] = α*x[n] + (1-α)*x[n-delay]
        for n := delay; n < len(segment); n++ {
            echo := e.config.Amplitude * segment[n-delay]
            segment[n] = e.config.MixRatio*segment[n] + (1-e.config.MixRatio)*echo
        }

        bitIndex++
    }

    // 4. Reconstruct WAV
    return buildWAV(wavData), nil
}

func (e *EchoTechnique) Extract(ctx context.Context, carrier []byte) ([]byte, error) {
    // 1. Parse WAV
    wavData, err := parseWAV(carrier)
    if err != nil {
        return nil, fmt.Errorf("parse WAV: %w", err)
    }

    // 2. Extract bits from each segment
    samples := wavData.Samples
    segmentCount := len(samples) / e.config.SegmentLen
    bits := make([]byte, 0, segmentCount)

    for i := 0; i < segmentCount; i++ {
        start := i * e.config.SegmentLen
        end := start + e.config.SegmentLen
        if end > len(samples) {
            end = len(samples)
        }
        segment := samples[start:end]

        // Compute autocorrelation
        maxLag := max(e.config.Delay0, e.config.Delay1) + 10
        acf := autocorrelation(segment, maxLag)

        // Find peak near expected delays
        peak0 := acf[e.config.Delay0]
        peak1 := acf[e.config.Delay1]

        // Classify based on which peak is stronger
        if peak0 > peak1 {
            bits = append(bits, 0)
        } else {
            bits = append(bits, 1)
        }
    }

    // 3. Convert bits to bytes
    return bitsToBytes(bits), nil
}
```

---

## Priority & Sequencing

### Recommended Implementation Order

1. **Phase Encoding First** (Easier, more educational)
   - Reason: FFT is well-understood, libraries mature
   - Learning curve: Moderate
   - Estimated time: 9-13 days

2. **Echo Hiding Second** (Simpler algorithm, trickier detection)
   - Reason: Can reuse WAV parsing infrastructure
   - Learning curve: Lower
   - Estimated time: 8-9 days

### Dependencies

**Both techniques require**:
- ✅ WAV parsing (already exists in `internal/infrastructure/media/audio_service.go`)
- ✅ Test infrastructure (already exists)
- ✅ Domain interfaces (already defined)
- ⚠️ FFT library (need to add gonum/fourier)
- ⚠️ Autocorrelation function (need to implement)

---

## Success Criteria

### Phase Encoding
- ✅ All 10 existing tests pass
- ✅ Round-trip integrity: 100% (embed → extract → diff === 0 bytes)
- ✅ Imperceptibility: SNR > 40dB
- ✅ Performance: 1MB WAV processes in < 500ms
- ✅ Robustness: Survives MP3 compression at 192kbps
- ✅ Capacity: At least 1 bit per 1024 samples

### Echo Hiding
- ✅ All 10 existing tests pass
- ✅ Round-trip integrity: 100%
- ✅ Imperceptibility: SNR > 35dB
- ✅ Performance: 1MB WAV processes in < 300ms
- ✅ Robustness: Survives noise up to -20dB SNR
- ✅ Capacity: At least 1 bit per 8192 samples

---

## References

### Academic Papers
1. **Bender et al.** (1996) - "Techniques for Data Hiding"
2. **Cvejic & Seppänen** (2002) - "Spread Spectrum Audio Watermarking"
3. **Gruhl et al.** (1996) - "Echo Hiding"

### Go DSP Resources
1. **gonum/fourier docs**: https://pkg.go.dev/gonum.org/v1/gonum/dsp/fourier
2. **mjibson/go-dsp**: https://github.com/mjibson/go-dsp
3. **DSP Guide**: https://www.dspguide.com/

### Shadowforge Internal References
1. **Audio Service**: `internal/infrastructure/media/audio_service.go` (WAV I/O)
2. **Stego Domain**: `internal/domain/stego/` (Interfaces)
3. **LSB Audio**: `internal/infrastructure/stego/lsb_audio.go` (Working reference)

---

**Next Steps:**
1. ✅ Review this documentation
2. ⚠️ Evaluate and add gonum/fourier dependency
3. ⚠️ Implement Phase Encoding (9-13 days)
4. ⚠️ Implement Echo Hiding (8-9 days)
5. ⚠️ Complete validation and performance testing

**Total Estimated Time**: ~17-22 days for both techniques

---

*This guide will be updated as implementation progresses.*
