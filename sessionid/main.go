package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func main() {
	id, err := genTerminalSessionId()
	fmt.Println(id, err)
}

func genTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	return string(id), nil
}
