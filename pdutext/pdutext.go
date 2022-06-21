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
	GSM7 = pdutext.GSM7
	// GSM7Packed is GSM 7-bit coding (packed).
	GSM7Packed = pdutext.GSM7 // FIXME: should be pdutext.GSM7Packed but it is not supported by Kannel.
	// UCS2 is UCS2 coding (UTF-16BE).
	UCS2 = pdutext.UCS2
)

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
	size := Size(message)
	single := SizeUCS2Single
	multipart := SizeUCS2Multipart
	if IsGSM7(message) {
		single = SizeGSM7Single
		multipart = SizeGSM7Multipart
	}

	if size <= single {
		return 1
	}

	segments := size / multipart
	if size%multipart != 0 {
		segments++
	}

	return segments
}

// Split in valid UTF-8 sequences.
func Split(message string, size int) []string {
	var i, n, m int
	segment := make([]rune, size)
	var segments []string

	for _, r := range message {
		n = i % size
		segment[n] = r

		m = n + 1
		if m == size {
			segments = append(segments, string(segment[:m]))
			m = 0
		}

		i++ // The index from range is the byte index and not the rune index.
	}

	if m > 0 {
		segments = append(segments, string(segment[:m]))
	}

	return segments
}
