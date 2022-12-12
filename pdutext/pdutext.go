package pdutext

import (
	"unicode/utf8"

	"github.com/mdouchement/smpp/smpp/pdu/pdutext"
)

const (
	// SizeGSM7Single is the max number of characters allowed in one SMS.
	SizeGSM7Single = 160
	// SizeGSM7Multipart is the max number of characters allowed in each chunk of the SMS due to the UDH.
	SizeGSM7Multipart = 153
	// SizeUCS2Single is the max number of characters allowed in one SMS.
	SizeUCS2Single = 70
	// SizeUCS2Multipart is the max number of characters allowed in each chunk of the SMS due to the UDH.
	SizeUCS2Multipart = 67
)

type (
	// Codec to define text codec.
	Codec = pdutext.Codec
	// Raw text codec, no encoding.
	Raw = pdutext.Raw
	// GSM7 is GSM 7-bit coding (7-bit on 8-bit space).
	// Not the official format but still used by several tools/SMSC.
	GSM7 = pdutext.GSM7
	// GSM7Packed is GSM 7-bit coding (packed).
	// This coding allows 160 characters coded over 140 bytes (160 * 7-bit / 8-bit = 140 bytes)
	// which is the format described in GSM 03.38.
	// FIXME: should be pdutext.GSM7Packed but it is not supported by Kannel.
	GSM7Packed = pdutext.GSM7
	// UCS2 is UCS2 coding (UTF-16BE).
	UCS2 = pdutext.UCS2
)

// TODO: not optimized, refactor to avoid to use `IsGSM7' each time.

// SelectCodec selects the right codec and computes details.
func SelectCodec(message string) (c Codec, size int, segments int) {
	if IsGSM7(message) {
		return GSM7Packed(message), Size(message), Segments(message)
	}
	return UCS2(message), Size(message), Segments(message)
}

// Size returns the size of the message.
func Size(message string) int {
	if IsGSM7(message) {
		return GSM7size(message)
	}
	return utf8.RuneCountInString(message) // size of the rune array
}

// Segments returns the number of segments used to send the given message.
func Segments(message string) int {
	if IsGSM7(message) {
		var segments, s, n int
		for _, r := range message {
			n = GSM7size(string(r))
			s += n
			switch {
			case s < SizeGSM7Multipart:
				// Still have place in the current segment.
			case s == SizeGSM7Multipart:
				// Segment complete.
				segments++
				s = 0
			default:
				// Over segment size.
				segments++
				s = n
			}
		}

		if s > 0 {
			segments++
		}

		return segments
	}

	size := Size(message)
	if size <= SizeUCS2Single {
		return 1
	}

	segments := size / SizeUCS2Multipart
	if size%SizeUCS2Multipart != 0 {
		segments++
	}

	return segments
}

// Split in valid UTF-8 sequences.
func Split(message string, size int) []string {
	var s, n int
	segment := make([]rune, 0, size+20)
	var segments []string

	for _, r := range message {
		n = Size(string(r))
		s += n
		switch {
		case s < size:
			// Still have place in the current segment.
			segment = append(segment, r)
		case s == size:
			// Segment complete.
			segment = append(segment, r)
			segments = append(segments, string(segment))
			segment = segment[:0]
			s = 0
		default:
			// Over segment size.
			segments = append(segments, string(segment))
			segment[0] = r
			segment = segment[:1]
			s = n
		}
	}

	if len(segment) > 0 {
		segments = append(segments, string(segment))
	}

	return segments
}
