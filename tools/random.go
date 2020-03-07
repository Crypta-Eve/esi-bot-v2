package tools

import (
	"math/rand"
	"time"
)

func UnsignedRandomIntWithMax(m int) uint {
	rand.Seed(time.Now().UnixNano())
	min := 0
	return uint(rand.Intn(m-min+1) + min)
}

func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
