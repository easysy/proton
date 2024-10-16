package sgen

import (
	"math/rand"
	"strings"
	"time"
)

var (
	Nums       characterSet = []rune("0123456789")
	UpLetters  characterSet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LowLetters characterSet = []rune("abcdefghijklmnopqrstuvwxyz")
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// characterSet a set of characters to generate a random string.
type characterSet []rune

func (cs characterSet) Append(sets ...[]rune) []rune {
	length := len(cs)
	for _, set := range sets {
		length += len(set)
	}

	newSet := make([]rune, length)

	i := copy(newSet, cs)

	for _, set := range sets {
		i += copy(newSet[i:], set)
	}

	return newSet
}

type RandomString struct {
	prefix       string
	characterSet []rune
	length       int
}

// Configure sets configuration for generates random string.
func (s *RandomString) Configure(prefix string, charSet []rune, randLength int) {
	s.prefix = prefix
	s.characterSet = charSet
	s.length = len(prefix) + randLength
}

// Generate generates random string according to configuration.
func (s *RandomString) Generate() string {
	b := new(strings.Builder)
	b.Grow(s.length)
	b.WriteString(s.prefix)
	for n := len(s.prefix); n < s.length; n++ {
		b.WriteRune(s.characterSet[rand.Intn(len(s.characterSet))])
	}
	return b.String()
}
