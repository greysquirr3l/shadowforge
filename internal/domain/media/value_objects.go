// Package media value objects.
package media

import (
"fmt"

"github.com/google/uuid"
)

// AssetID is a unique identifier for media assets.
type AssetID struct {
value string
}

// NewAssetID creates a new AssetID with validation.
func NewAssetID() AssetID {
return AssetID{value: uuid.New().String()}
}

// NewAssetIDFromString creates AssetID from string with validation.
func NewAssetIDFromString(id string) (AssetID, error) {
if id == "" {
return AssetID{}, ErrInvalidAssetID
}

// Validate UUID format
if _, err := uuid.Parse(id); err != nil {
return AssetID{}, fmt.Errorf("%w: invalid UUID format", ErrInvalidAssetID)
}

return AssetID{value: id}, nil
}

// String returns the string representation of AssetID.
func (a AssetID) String() string {
return a.value
}

// IsZero returns true if the AssetID is zero value.
func (a AssetID) IsZero() bool {
return a.value == ""
}

// Equals compares two AssetIDs for equality.
func (a AssetID) Equals(other AssetID) bool {
return a.value == other.value
}

// MediaType represents the type of media (Image, Audio, Text).
type MediaType string

const (
MediaTypeImage MediaType = "image"
MediaTypeAudio MediaType = "audio"
MediaTypeText  MediaType = "text"
)

// IsValid returns true if the MediaType is valid.
func (m MediaType) IsValid() bool {
switch m {
case MediaTypeImage, MediaTypeAudio, MediaTypeText:
return true
default:
return false
}
}

// String returns the string representation of MediaType.
func (m MediaType) String() string {
return string(m)
}

// MediaFormat represents the format of media files.
type MediaFormat string

const (
// Image formats
FormatPNG  MediaFormat = "png"
FormatJPEG MediaFormat = "jpeg"
FormatBMP  MediaFormat = "bmp"
FormatGIF  MediaFormat = "gif"

// Audio formats
FormatWAV  MediaFormat = "wav"
FormatFLAC MediaFormat = "flac"
FormatMP3  MediaFormat = "mp3"

// Text formats
FormatTXT      MediaFormat = "txt"
FormatMarkdown MediaFormat = "md"
)

// IsValid returns true if the MediaFormat is valid.
func (m MediaFormat) IsValid() bool {
switch m {
case FormatPNG, FormatJPEG, FormatBMP, FormatGIF,
FormatWAV, FormatFLAC, FormatMP3,
FormatTXT, FormatMarkdown:
return true
default:
return false
}
}

// String returns the string representation of MediaFormat.
func (m MediaFormat) String() string {
return string(m)
}

// IsImage returns true if the format is an image format.
func (m MediaFormat) IsImage() bool {
return m == FormatPNG || m == FormatJPEG || m == FormatBMP || m == FormatGIF
}

// IsAudio returns true if the format is an audio format.
func (m MediaFormat) IsAudio() bool {
return m == FormatWAV || m == FormatFLAC || m == FormatMP3
}

// IsText returns true if the format is a text format.
func (m MediaFormat) IsText() bool {
return m == FormatTXT || m == FormatMarkdown
}

// Dimensions represents image dimensions (width and height).
type Dimensions struct {
Width  int
Height int
}

// NewDimensions creates Dimensions with validation.
func NewDimensions(width, height int) (*Dimensions, error) {
if width <= 0 || height <= 0 {
return nil, ErrInvalidDimensions
}

return &Dimensions{
Width:  width,
Height: height,
}, nil
}

// IsValid returns true if dimensions are valid.
func (d *Dimensions) IsValid() bool {
return d != nil && d.Width > 0 && d.Height > 0
}

// Area returns the total area (width * height).
func (d *Dimensions) Area() int {
return d.Width * d.Height
}

// AspectRatio returns the aspect ratio (width / height).
func (d *Dimensions) AspectRatio() float64 {
if d.Height == 0 {
return 0.0
}
return float64(d.Width) / float64(d.Height)
}

// Resolution represents image resolution (DPI or PPI).
type Resolution struct {
Horizontal int // DPI/PPI horizontal
Vertical   int // DPI/PPI vertical
}

// NewResolution creates Resolution with validation.
func NewResolution(horizontal, vertical int) (*Resolution, error) {
if horizontal <= 0 || vertical <= 0 {
return nil, ErrInvalidResolution
}

return &Resolution{
Horizontal: horizontal,
Vertical:   vertical,
}, nil
}

// IsValid returns true if resolution is valid.
func (r *Resolution) IsValid() bool {
return r != nil && r.Horizontal > 0 && r.Vertical > 0
}

// SampleRate represents audio sample rate in Hz.
type SampleRate struct {
Hz       int // Sample rate in Hz
BitDepth int // Bit depth (8, 16, 24, 32)
Channels int // Number of channels (1=mono, 2=stereo)
}

// Common sample rates
const (
SampleRate8kHz   = 8000
SampleRate16kHz  = 16000
SampleRate22kHz  = 22050
SampleRate44kHz  = 44100
SampleRate48kHz  = 48000
SampleRate96kHz  = 96000
SampleRate192kHz = 192000
)

// NewSampleRate creates SampleRate with validation.
func NewSampleRate(hz, bitDepth, channels int) (*SampleRate, error) {
if hz <= 0 {
return nil, ErrInvalidSampleRate
}

if bitDepth != 8 && bitDepth != 16 && bitDepth != 24 && bitDepth != 32 {
return nil, ErrInvalidSampleRate
}

if channels < 1 || channels > 8 {
return nil, ErrInvalidSampleRate
}

return &SampleRate{
Hz:       hz,
BitDepth: bitDepth,
Channels: channels,
}, nil
}

// IsValid returns true if sample rate is valid.
func (s *SampleRate) IsValid() bool {
return s != nil &&
s.Hz > 0 &&
(s.BitDepth == 8 || s.BitDepth == 16 || s.BitDepth == 24 || s.BitDepth == 32) &&
s.Channels >= 1 && s.Channels <= 8
}

// ColorSpace represents image color space.
type ColorSpace string

const (
ColorSpaceRGB   ColorSpace = "rgb"
ColorSpaceRGBA  ColorSpace = "rgba"
ColorSpaceGray  ColorSpace = "gray"
ColorSpaceCMYK  ColorSpace = "cmyk"
ColorSpaceYCbCr ColorSpace = "ycbcr"
)

// IsValid returns true if the color space is valid.
func (c ColorSpace) IsValid() bool {
switch c {
case ColorSpaceRGB, ColorSpaceRGBA, ColorSpaceGray, ColorSpaceCMYK, ColorSpaceYCbCr:
return true
default:
return false
}
}

// String returns the string representation of ColorSpace.
func (c ColorSpace) String() string {
return string(c)
}

// HasAlpha returns true if the color space includes an alpha channel.
func (c ColorSpace) HasAlpha() bool {
return c == ColorSpaceRGBA
}
