package device

import (
	"crypto/rand"
	"strconv"
	"unicode"
)

const chars62 = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func newRandCharObf(val string) (obf, error) {
	length, err := strconv.Atoi(val)
	if err != nil {
		return nil, err
	}

	return &randCharObf{
		length: length,
	}, nil
}

type randCharObf struct {
	length int
}

func (o *randCharObf) Obfuscate(dst, src []byte) {
	rand.Read(dst[:o.length])
	for i := range dst[:o.length] {
		dst[i] = chars62[dst[i]%62]
	}
}

func (o *randCharObf) Deobfuscate(dst, src []byte) bool {
	for _, b := range src[:o.length] {
		r := rune(b)
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (o *randCharObf) ObfuscatedLen(n int) int {
	return o.length
}

func (o *randCharObf) DeobfuscatedLen(n int) int {
	return 0
}
