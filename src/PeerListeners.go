//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Add listeners for events from signalling server
func addPeerListeners() {

	// Update status
	connected = false
	destroyed = false

	// Open
	peer.Call("on", "open", jsGo.ProcOf(func(id []js.Value) {

		// Update status
		connected = true
		destroyed = false

		// Begin scanning for peeers
		scan(true)

		// Set up page
		jsGo.Document.Set("title", appName+" - "+room)
		app.Set("innerHTML", nil)

		// Leave button to disconnect from everything and destroy peer object
		if leaveButton.IsUndefined() {
			leaveButton = jsGo.CreateButton("Leave chat", func() {
				for i, _ := range conns {
					if !conns[i].IsUndefined() {conns[i].Call("close")}
				}
				connected = false
				destroyed = true
				conns = [roomMax]js.Value{}
				verified = [roomMax]bool{}
				peer.Call("disconnect")
				peer.Call("destroy")
				app.Set("innerHTML", nil)
				appAppendChild(jsGo.CreateButton("Re-enter chat", func() {jsGo.Location.Call("reload")}))
			})
			appAppendChild(leaveButton)
		}

		// Share link for share button and QR code
		url := jsGo.URL.New(jsGo.Location.Get("href"))
		url.Set("search", "")
		shareLink := url.Call("toString").String()+"?room="+stringToUrl(room)+"&password="+stringToUrl(password)

		// Share button to copy share link to clipboard
		if shareButton.IsUndefined() {
			shareButton = jsGo.CreateButton("Copy share link", func() {
				jsGo.Get("navigator").Get("clipboard").Call(
					"writeText",
					shareLink,
				)
			})
			appAppendChild(shareButton)
		}

		// Display QR code with share link
		if qrCode.IsUndefined() {
			qrCode = jsGo.CreateElement("div")
			appAppendChild(qrCode)
			jsGo.Get("QRCode").New(qrCode, shareLink)
		}
	}))
	

	// Connection
	peer.Call("on", "connection", jsGo.ProcOf(func(conn []js.Value) {
		metadata := conn[0].Get("metadata")

		// Verify the incomming connection request is from a peer in the same room before accepting request, and close if not
		if jsGo.String.New(conn[0].Get("peer")).Call("substring", 0, len(stringToUrl(room))).String() == stringToUrl(room) {
			connID := jsGo.ParseInt(jsGo.String.New(conn[0].Get("peer")).Call("substring", len(stringToUrl(room)))).Int()
			if !connected {conn[0].Call("close")
			} else {addConnListeners(conn[0], connID, false, metadata.Get("challenge").String())}
		} else {conn[0].Call("close")}
	}))

	// Close
	peer.Call("on", "close", jsGo.SimpleProcOf(func() {
		connected = false
		destroyed = true
	}))

	// Disconnected
	peer.Call("on", "disconnected", jsGo.SimpleProcOf(func() {
		if !destroyed {peer.Call("reconnect")
		} else {connected = false}
	}))

	// Error
	peer.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
		errType := err[0].Get("type").String()
		switch errType {

			// If requested ID in room is taken, incremenet ID and request new ID
			case "unavailable-id":
				peer.Call("disconnect")
				peer.Call("destroy")
				connected = false
				destroyed = true
				if roomID == roomMax-1 {exit("This room is already full!")}
				roomID++
				peer = jsGo.Get("Peer").New(stringToUrl(room)+jsGo.String.Invoke(roomID).String(), getOptions())
				addPeerListeners()

			// If no peer is found in the room at the given ID, wait a period of time before trying to connect to the next ID
			case "peer-unavailable":
				if connected && getConnCount() == 0 {
					jsGo.SetTimeout(jsGo.SimpleProcOf(func() {
						if connected && getConnCount() == 0 {scan(false)}
					}), 5000)
				}

			// Outout error to console by default
			default:
				jsGo.Log(errType)
		}
	}))
}
