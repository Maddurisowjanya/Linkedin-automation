package stealth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// TypeLikeHuman types text into the currently focused element on the page
// using variable per‑keystroke delays, occasional typos and backspace
// corrections to approximate a natural typing rhythm.
func TypeLikeHuman(page *rod.Page, text string) error {
	for _, r := range text {
		// Occasionally introduce a typo: press a neighbouring key, then
		// backspace and type the correct rune.
		if rand.Float64() < 0.03 {
			// Introduce a typo by briefly appending a neighbouring rune and
			// then "correcting" the value via JavaScript by trimming the
			// last character off before continuing.
			wrong := randomNearbyRune(r)
			if err := appendChar(page, wrong); err != nil {
				return err
			}
			time.Sleep(randomKeystrokeDelay())
			if err := backspaceJS(page); err != nil {
				return err
			}
			time.Sleep(randomKeystrokeDelay())
		}

		if err := appendChar(page, r); err != nil {
			return err
		}
		time.Sleep(randomKeystrokeDelay())
	}
	return nil
}

func randomKeystrokeDelay() time.Duration {
	// Base delay between 80‑220ms with a bit of jitter to form a rhythm.
	base := 80 + rand.Intn(140)
	return time.Duration(base) * time.Millisecond
}

func randomNearbyRune(r rune) rune {
	// Very small, approximate mapping of nearby characters on a QWERTY
	// keyboard to simulate realistic slips. For anything unknown, just
	// return the original rune.
	switch r {
	case 'a':
		return 's'
	case 's':
		return 'a'
	case 'e':
		return 'r'
	case 'r':
		return 'e'
	case 'n':
		return 'm'
	case 'm':
		return 'n'
	default:
		return r
	}
}

// appendChar appends a single rune to the currently focused input/textarea
// using a small JS helper. This avoids needing Rod's low‑level key codes and
// keeps the implementation portable across Rod versions.
func appendChar(page *rod.Page, r rune) error {
	script := fmt.Sprintf(`() => {
		const el = document.activeElement;
		if (!el) return;
		if (typeof el.value === "string") {
			el.value = (el.value || "") + %q;
		}
	}`, string(r))
	_, err := page.Eval(script, nil)
	return err
}

// backspaceJS simulates a backspace by trimming the last character from the
// active element's value via JS.
func backspaceJS(page *rod.Page) error {
	script := `() => {
		const el = document.activeElement;
		if (!el || typeof el.value !== "string") return;
		el.value = el.value.slice(0, -1);
	}`
	_, err := page.Eval(script, nil)
	return err
}



