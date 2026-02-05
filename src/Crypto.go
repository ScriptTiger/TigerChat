//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Generate a salt, which will be used as a challenge and hashed together with the password using Argon2
func generateSalt() (string) {
	salt := jsGo.Uint8Array.New(16)
	jsGo.Crypto.Call("getRandomValues", salt)
	return salt.Call("toBase64").String()
}

// Argon2 handler, which takes a salt, the password, and a callback to asynchonrously handle the returned hash
func argon2(salt, password string, argon2Callback func(hash string)) {
	jsGo.ThenableChain(
		jsGo.Get("argon2").Call("hash", map[string]any{
			"pass": password,
			"salt": salt,
		}),
		func(hash js.Value) (any) {
			argon2Callback(hash.Get("hashHex").String())
			return nil
		},
	)
}
