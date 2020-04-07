package vterm

import "log"

func (v *VTerm) handlePrivateSequence(next rune, parameterCode string) {
	fullCode := parameterCode + string(next)

	switch next {
	case 'h': // generally enables features
		switch parameterCode {
		case "1": // application arrow keys (DECCKM)
		case "7": // Auto-wrap Mode (DECAWM)
		case "12": // start blinking Cursor
		case "25": // show Cursor
		case "1049", "1047", "47": // switch to alt screen buffer
			if !v.usingAltScreen {
				v.screenBackup = v.Screen
			}
		case "2004": // enable bracketed paste mode
		default:
			log.Printf("Unrecognized private H code: %v", fullCode)
		}
	case 'l': // generally disables features
		switch parameterCode {
		case "1": // Normal Cursor keys (DECCKM)
		case "7": // No Auto-wrap Mode (DECAWM)
		case "12": // stop blinking Cursor
		case "25": // hide Cursor
		case "1049", "1047", "47": // switch to normal screen buffer
			if v.usingAltScreen {
				v.Screen = v.screenBackup
			}
		case "2004": // disable bracketed paste mode
		default:
			log.Printf("Unrecognized private L code: %v", fullCode)
		}
	default:
		log.Printf("Unrecognized private code: %v", fullCode)
	}
}
