package util

import (
	"crypto/rand"
	"fmt"
)

type UUID [16]byte

// NewUUID generates a random v4 UUID
func NewUUID() UUID {
	u := [16]byte{}
	_, err := rand.Read(u[:])
	if err != nil {
		panic(err)
	}

	u[8] = (u[8] | 0x80) & 0xBF
	u[6] = (u[6] | 0x40) & 0x4F

	return u
}

func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:16])
}
