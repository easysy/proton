package sgen

import (
	"crypto/rand"
	"math/big"
	"strings"
)

var (
	Nums       characterSet = []rune("0123456789")
	UpLetters  characterSet = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	LowLetters characterSet = []rune("abcdefghijklmnopqrstuvwxyz")
)

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

// Configure sets configuration for generating random strings.
func (s *RandomString) Configure(prefix string, charSet []rune, randLength int) {
	s.prefix = prefix
	s.characterSet = charSet
	s.length = len(prefix) + randLength
}

// Generate generates a cryptographically secure random string.
func (s *RandomString) Generate() string {
	b := new(strings.Builder)
	b.Grow(s.length)
	_, _ = b.WriteString(s.prefix)

	for n := len(s.prefix); n < s.length; n++ {
		r, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.characterSet))))
		_, _ = b.WriteRune(s.characterSet[r.Int64()])
	}

	return b.String()
}
