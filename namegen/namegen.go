package namegen

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	// ADJ is a list of adjectives used in the name generator.
	ADJ = []string{"adaptable", "adventurous", "ambitious", "amusing",
		"agreeable", "brave", "bright", "calm", "charming", "considerate",
		"courageous", "creative", "decisive", "diligent", "diplomatic",
		"discreet", "dynamic", "enthusiastic", "exuberant", "faithful",
		"fearless", "friendly", "funny", "generous", "gentle", "gregarious",
		"helpful", "honest", "humorous", "imaginative", "impartial",
		"idependent", "intellectual", "kind", "loving", "loyal", "neat",
		"nice", "passionate", "persistent", "polite", "powerful", "quiet",
		"rational", "reliable", "romantic", "thoughtful", "tidy"}

	// NOUN is a list of nouns used in the name generator.
	NOUN = []string{"Ada", "Alef", "ALGOL", "Asm", "Awk", "Bash", "BASIC",
		"BCPL", "C", "C++", "C#", "Chicken", "Chapel", "COBOL", "CoffeeScript",
		"Crystal", "D", "Dart", "Eiffel", "Elixer", "Elm", "Elisp", "Erlang",
		"F#", "Forth", "Fortran", "Guile", "Go", "Hack", "Haskell", "Haxe", "HolyC",
		"Io", "Idris", "Java", "JavaScript", "Julia", "Kotlin", "Limbo", "Lisp", "Lua",
		"MATLAB", "Nim", "OCaml", "Octave", "Pascal", "Perl", "PHP", "PowerShell", "Prolog",
		"Python", "R", "ReasonML", "Ruby", "Rust", "SAS", "Scala", "Scheme", "Scratch",
		"Smalltalk", "SML", "Swift", "Tcl", "TeX", "TypeScript", "Vala"}
)

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// GetName returns a randomly generated name which is formatted.
func GetName() (name string) {
	name = fmt.Sprintf("%s %s", ADJ[rnd.Intn(len(ADJ))],
		NOUN[rnd.Intn(len(NOUN))])
	name = strings.Title(name)
	return
}
