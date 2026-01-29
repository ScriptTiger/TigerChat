//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

// Add listeners for connection events
func connectionHandler(conn js.Value, connID int, initiator bool, salt string) {

	// Cancel if connection to signaling server is currently disconnected or destroyed
	if !connected || destroyed {return}

	// Connection name, given by peer
	var connName string

	// Add connection to global connections list
	conns[connID] = conn

	// Start timeout timer
	jsGo.SetTimeout(jsGo.SimpleProcOf(func() {
		if !connected || destroyed {return}
		if !verified[connID] {
			conn.Call("close")
			conns[connID] = js.Undefined()
			if connected && !destroyed && getConnCount() == 0 {scan(false)}
		}
	}), 5000)

	// Data
	conn.Call("on", "data", jsGo.ProcOf(func(data []js.Value) {

		// Check if the incoming data is a metadata message or not, and handle accordingly if it is
		if jsGo.HasOwn(data[0], "metadata") {
			metadata := data[0].Get("metadata")
			msg := metadata.Get("msg").String()

			// Check if the incoming metadata message is from a verified (authenticated) connection or not, and handle accordingly
			if !verified[connID] {

				// Handle metadata messages from unverified connections
				switch msg {

					// Verify the response given by the receiver, and close if it is incorrect
					case "response-challenge":
						jsGo.ThenableChain(argon2(salt, password), func(hash js.Value) (any) {
							hashStr := hash.Get("hashHex").String()
							if metadata.Get("response").String() == hashStr {
								verified[connID] = true

								// Reply with the response to the receiver's challenge once the receiver has been verified (authenticated)
								jsGo.ThenableChain(argon2(metadata.Get("challenge").String(), password), func(hash js.Value) (any) {
									hashStr := hash.Get("hashHex").String()
									conn.Call("send", map[string]any{"metadata": map[string]any{
										"msg": "response",
										"response": hashStr,
									}})
									return nil
								})
							} else {conn.Call("close")}
							return nil
						})

					// Verify the response given by the initiator, and close if it is incorrect
					case "response":
						jsGo.ThenableChain(argon2(salt, password), func(hash js.Value) (any) {
							hashStr := hash.Get("hashHex").String()
							if metadata.Get("response").String() == hashStr {
								verified[connID] = true

								// Sand name once the initiator has been verified
								conn.Call("send", map[string]any{"metadata": map[string]any{
									"msg": "name-request",
									"name": name,
								}})

								// Announce new verified connection to all peers
								sendAll(map[string]any{"metadata": map[string]any{
									"msg": "intro",
									"id": connID,
								}})

							// Close connection by default
							} else {conn.Call("close")}
							return nil
						})

					// Close connection by default
					default:
						conn.Call("close")
				}
			} else {

				// Handle metadata messages from verified connections
				switch msg {

					// Receive name from receiver, announce them in chat history, and reply with initiator's name
					case "name-request":
						connName = metadata.Get("name").String()
						chat(connName+" has entered the chat")
						conn.Call("send", map[string]any{"metadata": map[string]any{
							"msg": "name-response",
							"name": name,
						}})

					// Receive initiator's name and announce them in chat history
					case "name-response":
						connName = metadata.Get("name").String()
						chat(connName+" has entered the chat")

					// Receive announcement of new peer and initiate connection with that new peer
					case "intro":
						connID := metadata.Get("id").Int()
						if roomID != connID && conns[connID].IsUndefined() {connect(connID)}

					// Close connection by default
					default:
						conn.Call("close")
				}
			}

		// Handle data messages from verified connections
		} else if verified[connID] {

			// Append text messages to the chat history
			if jsGo.HasOwn(data[0], "text") {
				chat(connName+": "+data[0].Get("text").String())

			// Append received images to the chat history
			} else if jsGo.HasOwn(data[0], "image") {
				img := jsGo.CreateElement("img")
				img.Set("src", jsGo.URL.Call("createObjectURL", jsGo.Blob.New(data[0].Get("image"))))
				chat(connName+": ")
				chat(img)

			// Close connection by default
			} else {conn.Call("close")}

		// Close connection by default
		} else {conn.Call("close")}
	}))

	// Open
	conn.Call("on", "open", jsGo.SimpleProcOf(func() {

		// Trigger the receiver of the connection to reply with a response for the initial challenge, and also with their own challenge
		if !verified[connID] && !initiator {
			mySalt := generateSalt()
			jsGo.ThenableChain(argon2(salt, password), func(hash js.Value) (any) {
				hashStr := hash.Get("hashHex")
				conn.Call("send", map[string]any{"metadata": map[string]any{
					"msg": "response-challenge",
					"response": hashStr,
					"challenge": mySalt,
				}})
				salt = mySalt
				return nil
			})
		}

		// Set up chat elements if not set up already
		if fileButton.IsUndefined() {

			// Three-line break between chat elements and leave button
			appPrepend(jsGo.CreateElement("br"))
			appPrepend(jsGo.CreateElement("br"))
			appPrepend(jsGo.CreateElement("br"))

			// File button to send files
			fileButton = jsGo.CreateLoadFileButton("Image", ".jpg, .jpeg, image/jpeg, .png, image/png, .gif, image/gif", false, func(event js.Value) {
				file := event.Get("target").Get("files").Index(0)
				if jsGo.String.New(file.Get("type")).Call("split", "/", 1).Index(0).String() == "image" {
					sendAllImage(file)
				} else {
					jsGo.Alert("You may only send valid image files at this time!")
				}
			})
			appPrepend(fileButton)

			// Send button to trigger text being sent to all peers as well as chat history, and clearing text
			sendButton = jsGo.CreateButton("Send", func() {sendAllText()})
			appPrepend(sendButton)
			appPrepend(jsGo.CreateElement("br"))

			// Text area to type messages which will be sent to peers
			appAppendChild(jsGo.CreateElement("br"))
			textArea = jsGo.CreateElement("textarea")
			textArea.Set("style", "resize: none;")
			textArea.Call("addEventListener", "keydown", jsGo.ProcOf(func(event []js.Value) {
				if event[0].Get("key").String() == "Enter" {
					event[0].Call("preventDefault")
					sendAllText()
				}
			}))
			appPrepend(textArea)

			// Chat history
			chatArea = jsGo.CreateElement("div")
			appPrepend(chatArea)
		}

	}))

	// Close
	conn.Call("on", "close", jsGo.SimpleProcOf(func() {
		conns[connID] = js.Undefined()
		if verified[connID] {
			verified[connID] = false
			chat(connName+" has left the chat")
		}
		if connected && !destroyed && getConnCount() == 0 {
			jsGo.SetTimeout(jsGo.SimpleProcOf(func() {
				if connected && !destroyed && getConnCount() == 0 {scan(true)}
			}), 5000)
		}
	}))

	// Error
	conn.Call("on", "error", jsGo.ProcOf(func(err []js.Value) {
		jsGo.Log("Data connection error with "+conn.Get("peer").String()+": "+err[0].Get("type").String())
	}))

}
