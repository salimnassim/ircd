package ircd

func parseMask(mask string) ([]byte, error) {
	var result []byte

	for _, char := range mask {
		switch {
		case char == '?':
			result = append(result, 0x3F)
		case char == '*':
			result = append(result, 0x2A)
		case (char >= 0x01 && char <= 0x29) || (char >= 0x2B && char <= 0x3E) || (char >= 0x40 && char <= 0xFF):
			result = append(result, byte(char))
		case (char >= 0x01 && char <= 0x5B) || (char >= 0x5D && char <= 0xFF):
			result = append(result, byte(char))
		default:
			return nil, errorBadMaskCharadcter
		}
	}

	return result, nil
}

func matchMask(mask []byte, input string) bool {
	if len(mask) == 0 {
		return true
	}

	if len(mask) > len(input) {
		return false
	}

	for i := 0; i < len(mask); i++ {
		switch mask[i] {
		case '?':
			continue
		case '*':
			for j := 0; j <= len(input)-i; j++ {
				if matchMask(mask[i+1:], input[i+j:]) {
					return true
				}
			}
			return false
		default:
			if mask[i] != input[i] {
				return false
			}
		}
	}
	return true
}
