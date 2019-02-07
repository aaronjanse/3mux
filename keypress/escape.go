package keypress

import (
	"strings"
	"unicode"
)

func handleEscapeCode() {
	data := []byte{27}

	k := next()
	data = append(data, byte(k))
	switch k {
	case 91:
	default:
		letter := rune(k)
		uppercase := strings.ToUpper(string(letter))
		if unicode.IsUpper(letter) {
			callback("Alt+Shift+"+uppercase, data)
		} else {
			callback("Alt+"+uppercase, data)
		}
		return
	}
	if k != 91 {
		callback("", data)
		return
	}

	k = next()
	data = append(data, byte(k))
	if k != 49 {
		callback("", data)
		return
	}

	k = next()
	data = append(data, byte(k))
	if k != 59 {
		callback("", data)
		return
	}

	out := "Alt+"

	k = next()
	data = append(data, byte(k))
	switch k {
	case 51:
	case 52:
		out += "Shift+"
	default:
		callback(out, data)
		return
	}

	k = next()
	data = append(data, byte(k))
	switch k {
	case 65:
		out += "Up"
	case 66:
		out += "Down"
	case 67:
		out += "Right"
	case 68:
		out += "Left"
	default:
		callback(out, data)
		return
	}

	callback(out, data)
}
