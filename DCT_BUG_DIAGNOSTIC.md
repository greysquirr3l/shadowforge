# DCT Implementation Bug Analysis

## Issue Discovered

DCT steganography implementation has systematic bit manipulation corruption causing 6/17 test failures.

## Corruption Pattern

Input: [0x00, 0xff, 0xaa, 0x55] → Output: [0x01, 0xfe, 0x2a, 0x5e]

### Binary Analysis

- 0x00 (00000000) → 0x01 (00000001) : bit 0 incorrectly set
- 0xff (11111111) → 0xfe (11111110) : bit 0 incorrectly cleared
- 0xaa (10101010) → 0x2a (00101010) : high bits corrupted, bit 1 cleared
- 0x55 (01010101) → 0x5e (01011110) : multiple bits corrupted

## Root Cause Hypothesis

**BIT POSITION MISMATCH**: Code uses bit 1 for embedding but extraction has inconsistent bit position handling.

### Code Analysis

**Embedding** (line 217):

```go
// Modify bit 1 (not LSB) for better JPEG compression robustness
// Clear bit 1, then set it: (value & 0xFD) | (dataBit << 1)
*channelValue = (*channelValue & 0xFD) | (dataBit << 1)
```

**Extraction** (line 259):

```go
// Extract bit 1 (not LSB/bit 0) - matches embedding
dataBit := (channelValue >> 1) & 1
```

The embedding and extraction LOOK correct for bit 1, but there may be:

1. JPEG quality issues not preserving bit 1 modifications
2. Buffer index calculation errors causing wrong bytes to be modified
3. Bit ordering issues in the byte reconstruction logic

## Immediate Fix Strategy

1. Create minimal diagnostic test with single byte [0x00]
2. Add extensive logging to track every bit manipulation
3. Verify bit 1 is actually preserved by JPEG encoding at quality 100
4. Test hypothesis: temporarily change to bit 0 (LSB) to match working LSB implementation

## Length Header Corruption

Additional issue: Binary length encoding causing:

- 500 bytes reading as 2,147,487,720
- Suggests binary.BigEndian corruption or buffer overflow

## Status

CRITICAL BLOCKER - All DCT functionality non-functional until resolved.
