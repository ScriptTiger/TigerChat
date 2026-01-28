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

// Hash the password and a salt together, returning a promise which will resolve to the hash
func argon2(salt, password string) (promise js.Value) {
	return jsGo.Get("argon2").Call("hash", map[string]any{
		"pass": password,
		"salt": jsGo.Uint8Array.Call("fromBase64", salt).String(),
	})
}
