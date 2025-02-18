package textutil

import (
	"bytes"
	"unicode"
)

// ToSnakeCase converts given name to "snake_text_format".
func ToSnakeCase(name string) string {
	multipleUpper := false

	var (
		ret         bytes.Buffer
		lastUpper   rune
		beforeUpper rune
	)

	for _, c := range name {
		// Non-lowercase character after uppercase is considered to be uppercase too.
		isUpper := (unicode.IsUpper(c) || (lastUpper != 0 && !unicode.IsLower(c)))

		// Output a delimiter if last character was either the first uppercase character
		// in a row, or the last one in a row (e.g. 'S' in "HTTPServer").
		// Do not output a delimiter at the beginning of the name.
		if lastUpper != 0 {
			firstInRow := !multipleUpper
			lastInRow := !isUpper

			if ret.Len() > 0 && (firstInRow || lastInRow) && beforeUpper != '_' {
				ret.WriteByte('_')
			}

			ret.WriteRune(unicode.ToLower(lastUpper))
		}

		// Buffer uppercase char, do not output it yet as a delimiter may be required if the
		// next character is lowercase.
		if isUpper {
			multipleUpper = (lastUpper != 0)
			lastUpper = c

			continue
		}

		ret.WriteRune(c)

		lastUpper = 0
		beforeUpper = c
		multipleUpper = false
	}

	if lastUpper != 0 {
		ret.WriteRune(unicode.ToLower(lastUpper))
	}

	return ret.String()
}
