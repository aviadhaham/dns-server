package main

import "strings"

func ParseDomainName(labelSequence []byte) string {
	if len(labelSequence) == 0 {
		return ""
	}

	var parts []string
	i := 0

	for i < len(labelSequence) {
		labelLength := int(labelSequence[i])
		if labelLength == 0 {
			break // End of domain name
		}

		i++ // Move past the length byte
		if i+labelLength > len(labelSequence) {
			break // Prevent out of bounds
		}

		label := string(labelSequence[i : i+labelLength])
		parts = append(parts, label)
		i += labelLength
	}

	return strings.Join(parts, ".")
}
