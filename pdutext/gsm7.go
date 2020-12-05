package pdutext

// IsGSM7 returns true if the given string complies with GSM7 table.
func IsGSM7(message string) bool {
	i, ns := 0, len(message)

	for i < ns {
		if message[i] >= 32 && message[i] <= 95 { // ASCII table from ! to _
			i++
			continue
		}

		if message[i] >= 97 && message[i] <= 126 { // ASCII table from a to ~
			i++
			continue
		}

		switch message[i] {
		case 10, 13: // \n \r
			i++
			continue
		}

		// UTF-8 sequence
		if message[i] == 194 {
			if i+1 < ns {
				switch message[i+1] {
				case 161: // ¡
					fallthrough
				case 163: // £
					fallthrough
				case 164: // ¤
					fallthrough
				case 165: // ¥
					fallthrough
				case 167: // §
					fallthrough
				case 191: // ¿
					i += 2
					continue
				}
			}
		}
		if message[i] == 195 {
			if i+1 < ns {
				switch message[i+1] {
				case 132: // Ä
					fallthrough
				case 133: // Å
					fallthrough
				case 134: // Æ
					fallthrough
				case 135: // Ç
					fallthrough
				case 137: // É
					fallthrough
				case 145: // Ñ
					fallthrough
				case 150: // Ö
					fallthrough
				case 152: // Ø
					fallthrough
				case 156: // Ü
					fallthrough
				case 159: // ß
					fallthrough
				case 160: // à
					fallthrough
				case 164: // ä
					fallthrough
				case 165: // å
					fallthrough
				case 166: // æ
					fallthrough
				case 168: // è
					fallthrough
				case 169: // é
					fallthrough
				case 172: // ì
					fallthrough
				case 177: // ñ
					fallthrough
				case 178: // ò
					fallthrough
				case 182: // ö
					fallthrough
				case 184: // ø
					fallthrough
				case 185: // ù
					fallthrough
				case 188: // ü
					i += 2
					continue
				}
			}
		}
		if message[i] == 206 {
			if i+1 < ns {
				switch message[i+1] {
				case 147: // Γ
					fallthrough
				case 148: // Δ
					fallthrough
				case 152: // Θ
					fallthrough
				case 155: // Λ
					fallthrough
				case 158: // Ξ
					fallthrough
				case 160: // Π
					fallthrough
				case 163: // Σ
					fallthrough
				case 166: // Φ
					fallthrough
				case 168: // Ψ
					fallthrough
				case 169: // Ω
					i += 2
					continue
				}
			}
		}
		if message[i] == 226 {
			if i+2 < ns && message[i+1] == 130 && message[i+2] == 172 { // €
				i += 3
				continue
			}
		}

		return false
	}

	return true
}

// GSM7size computes the message's length based on the GSM 03.38 table from the given UTF-8 string.
func GSM7size(message string) (size int) {
	i, ns := 0, len(message)

	for i < ns {
		switch message[i] {
		case 91, 92, 93, 94, 123, 124, 125, 126: // [ \ ] ^ { | } ~
			size += 2 // These characters count for 2 characters in GSM7 table
			i++
			continue
		}

		// ASCII for both supported and unsupported by GSM7
		if message[i] < 128 {
			size++
			i++
			continue
		}

		// UTF-8 sequence
		if message[i] == 194 {
			if i+1 < ns {
				switch message[i+1] {
				case 161: // ¡
					fallthrough
				case 163: // £
					fallthrough
				case 164: // ¤
					fallthrough
				case 165: // ¥
					fallthrough
				case 167: // §
					fallthrough
				case 191: // ¿
					size++
					i += 2
					continue
				}
			}
		}
		if message[i] == 195 {
			if i+1 < ns {
				switch message[i+1] {
				case 132: // Ä
					fallthrough
				case 133: // Å
					fallthrough
				case 134: // Æ
					fallthrough
				case 135: // Ç
					fallthrough
				case 137: // É
					fallthrough
				case 145: // Ñ
					fallthrough
				case 150: // Ö
					fallthrough
				case 152: // Ø
					fallthrough
				case 156: // Ü
					fallthrough
				case 159: // ß
					fallthrough
				case 160: // à
					fallthrough
				case 164: // ä
					fallthrough
				case 165: // å
					fallthrough
				case 166: // æ
					fallthrough
				case 168: // è
					fallthrough
				case 169: // é
					fallthrough
				case 172: // ì
					fallthrough
				case 177: // ñ
					fallthrough
				case 178: // ò
					fallthrough
				case 182: // ö
					fallthrough
				case 184: // ø
					fallthrough
				case 185: // ù
					fallthrough
				case 188: // ü
					size++
					i += 2
					continue
				}
			}
		}
		if message[i] == 206 {
			if i+1 < ns {
				switch message[i+1] {
				case 147: // Γ
					fallthrough
				case 148: // Δ
					fallthrough
				case 152: // Θ
					fallthrough
				case 155: // Λ
					fallthrough
				case 158: // Ξ
					fallthrough
				case 160: // Π
					fallthrough
				case 163: // Σ
					fallthrough
				case 166: // Φ
					fallthrough
				case 168: // Ψ
					fallthrough
				case 169: // Ω
					size++
					i += 2
					continue
				}
			}
		}
		if message[i] == 226 {
			if i+2 < ns && message[i+1] == 130 && message[i+2] == 172 { // €
				size += 2 // Count for 2 characters in GSM7 table
				i += 3
				continue
			}
		}

		// Fallback for an unknown byte sequence
		size++
		i++
	}

	return size
}
