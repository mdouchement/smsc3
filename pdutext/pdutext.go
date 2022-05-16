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
	// GSM7 is GSM 7-bit coding.
	GSM7 = pdutext.GSM7
	// UCS2 is UCS2 coding (UTF-16BE).
	UCS2 = pdutext.UCS2
)

// SelectCodec selects the right codec and computes details.
func SelectCodec(message string) (c Codec, size int, segments int) {
	if IsGSM7(message) {
		return GSM7(message), Size(message), Segments(message)
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
