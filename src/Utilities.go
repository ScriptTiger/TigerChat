//go:build ecmascript

package main

import (
	"runtime"
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// DOM method handlers for the app container

// Append a child element to the app container
func appAppendChild(child js.Value) {app.Call("appendChild", child)}

// Prepend a child element to the app container
func appPrepend(child js.Value) {app.Call("prepend", child)}

// Messaging functions

// Append a message to the chat history
func chat(msg string) {
	chatArea.Call("append", msg)
	chatArea.Call("appendChild", jsGo.CreateElement("br"))
}

// Send a message to all peers
func sendAll(msg any) {
	for i, _ := range conns {
		if i != roomID && !conns[i].IsUndefined() && verified[i] {conns[i].Call("send", msg)}
	}
}

// Retrieve the user input text, send it to all peers, append it to your own chat history, and then clear the input text
func sendAllText() {
	msg := textArea.Get("value")
	if msg.String() != "" {
		sendAll(msg)
		chat("Me: "+msg.String())
		textArea.Set("value", nil)
	}
}

// Query string functions

// Encode a string as a base64 URL-safe string which also meets PeerJS peer ID naming constraints
func stringToUrl(str string) (string) {
	return "T"+jsGo.String.New(jsGo.String.New(jsGo.String.New(jsGo.Btoa(str)).Call("replaceAll", "=", "")).Call("replaceAll", "+", "-")).Call("replaceAll", "/", "_").String()
}

// Decade a base64 URL-safe string back to a regular string
func urlToString(str string) (string) {
	return jsGo.Atob(jsGo.String.New(jsGo.String.New(jsGo.String.New(str).Call("replaceAll", "-", "+")).Call("replaceAll", "_", "/")).Call("substring", 1)).String()
}

// Connection functions

// Return peer options, currently only used for passing ICE/TURN configuration if present
func getOptions() (map[string]any) {
	if turnUrl == "" || turnUser == "" || turnCred == "" {return map[string]any{}}
	if policy == "" {policy = "all"}
	return map[string]any{
		"config": map[string]any{
			"iceServers": []any{
				map[string]any{
					"urls": "turn:"+turnUrl,
					"username": turnUser,
					"credential": turnCred,
				},
			},
			"iceTransportPolicy": policy,
		},
	}
}

// Count the number of currently defined connections
func getConnCount() (count int) {
	for i, _ := range conns {
		if !conns[i].IsUndefined() {count++}
	}
	return
}

// Initiate a connection with a challenge
func connect(id int) {
	if !connected && destroyed {return}
	peerID := stringToUrl(room)+jsGo.String.Invoke(id).String()
	salt := generateSalt()
	conn := peer.Call(
		"connect",
		peerID,
		map[string]any{"metadata": map[string]any{"challenge": salt}},
	)
	addConnListeners(conn, id, true, salt)
}

// Scan for available peers in a room
func scan(reset bool) {
	if !connected || destroyed || getConnCount() > 0 {return}
	if reset {
		if roomID == 0 {nextTry = 1
		} else {nextTry = roomID-1}
	} else {
		if roomID == 0 {
			if nextTry == roomMax-1 {nextTry = 0}
			nextTry++
		} else if roomID == roomMax-1 {
			if nextTry == 0 {nextTry = roomID}
			nextTry--
		} else {
			if nextTry == 0 {nextTry = roomID+1
			} else if nextTry == roomMax-1 {nextTry = roomID-1
			} else if nextTry > roomID {nextTry++
			} else {nextTry--}
		}
	}
	connect(nextTry)
}

// Misc utility functions

// Hard exit
func exit(str string) {
	app.Set("innerHTML", str)
	runtime.Goexit()
}
