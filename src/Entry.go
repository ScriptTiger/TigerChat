//go:build ecmascript

package main

import (
	"syscall/js"

	"github.com/ScriptTiger/jsGo"
)

const (
	// App name used for titling various things
	appName = "TigerChat"

	// Maximum connections per room
	roomMax = 8
)

var (
	// Location elements
	urlRaw, urlClean, room, name, password, turnUrl, turnUser, turnCred, policy string

	// Status tracking of the signalling server
	connected, destroyed bool

	// Global JS objects
	app, peer,
	chatArea, textArea, sendButton, fileButton,
	leaveButton, shareButton, qrCode js.Value

	// Room management
	roomID, nextTry int

	// Connection tracking
	conns [roomMax]js.Value
	verified [roomMax]bool
)

// Set up app, making sure it has everything it needs to proceed
func main() {

	// Get app location in DOM
	app = jsGo.GetElementById("app")

	// Check URL query strings
	hasRoom := jsGo.Params.Call("has", "room").Bool()
	hasName := jsGo.Params.Call("has", "name").Bool()
	hasPassword := jsGo.Params.Call("has", "password").Bool()
	hasTurnUrl := jsGo.Params.Call("has", "turnurl").Bool()
	hasTurnUser := jsGo.Params.Call("has", "turnuser").Bool()
	hasTurnCred := jsGo.Params.Call("has", "turncred").Bool()
	hasPolicy  := jsGo.Params.Call("has", "policy").Bool()

	// Decode base64 URL-safe strings back to regular strings
	if hasRoom {room = urlToString(jsGo.Params.Call("get", "room").String())}
	if hasName {name = urlToString(jsGo.Params.Call("get", "name").String())}
	if hasPassword {password = urlToString(jsGo.Params.Call("get", "password").String())}
	if hasTurnUrl {turnUrl = urlToString(jsGo.Params.Call("get", "turnurl").String())}
	if hasTurnUser {turnUser = urlToString(jsGo.Params.Call("get", "turnuser").String())}
	if hasTurnCred {turnCred = urlToString(jsGo.Params.Call("get", "turncred").String())}
	if hasPolicy {policy = urlToString(jsGo.Params.Call("get", "policy").String())}

	// Capture URL
	url := jsGo.URL.New(jsGo.Location.Get("href"))
	urlRaw = url.Call("toString").String()
	url.Set("search", "")
	urlClean = url.Call("toString").String()

	// Wipe current query strings without reloading
	jsGo.History.Call("replaceState", nil, nil, urlClean)

	// If the room, name, and password are not all given, present the user with input fields to input the needed information
	if !(hasRoom && hasName && hasPassword) {

		// Prepare page
		jsGo.Document.Set("title", appName+" - Log in")
		app.Set("innerHTML", nil)

		// TURN settings container
		var turnSettingsVisible bool
		turnSettings := jsGo.CreateElement("div")
		turnSettings.Set("hidden", true)

		// TURN URL input
		turnUrlLabel := jsGo.CreateElement("label")
		turnUrlLabel.Set("textContent", "URL:")
		turnSettings.Call("appendChild", turnUrlLabel)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))
		turnUrlField := jsGo.CreateElement("input")
		if hasTurnUrl {turnUrlField.Set("value", turnUrl)}
		turnSettings.Call("appendChild", turnUrlField)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))

		// TURN user input
		turnUserLabel := jsGo.CreateElement("label")
		turnUserLabel.Set("textContent", "User Name:")
		turnSettings.Call("appendChild", turnUserLabel)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))
		turnUserField := jsGo.CreateElement("input")
		if hasTurnUser {turnUserField.Set("value", turnUser)}
		turnSettings.Call("appendChild", turnUserField)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))

		// TURN credential input
		turnCredLabel := jsGo.CreateElement("label")
		turnCredLabel.Set("textContent", "Credential:")
		turnSettings.Call("appendChild", turnCredLabel)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))
		turnCredField := jsGo.CreateElement("input")
		turnCredField.Set("type", "password")
		if hasTurnCred {turnCredField.Set("value", turnCred)}
		turnSettings.Call("appendChild", turnCredField)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))

		// TURN policy selection
		policyLabel := jsGo.CreateElement("label")
		policyLabel.Set("textContent", "ICE Transport Policy:")
		turnSettings.Call("appendChild", policyLabel)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))
		policySelect := jsGo.CreateElement("select")
		policyAll := jsGo.CreateElement("option")
		policyAll.Set("value", "all")
		policyAll.Set("textContent", "All")
		policyAll.Set("title", "Select the best network path of all candidates, even if it is not through TURN")
		policySelect.Call("appendChild", policyAll)
		policyRelay := jsGo.CreateElement("option")
		policyRelay.Set("value", "relay")
		policyRelay.Set("textContent", "Relay")
		policyRelay.Set("title", "Select the best network path of only relay candidates to ensure only TURN is used")
		policySelect.Call("appendChild", policyRelay)
		if hasPolicy {policySelect.Set("value", policy)
		} else {policySelect.Set("value", "all")}
		turnSettings.Call("appendChild", policySelect)
		turnSettings.Call("appendChild", jsGo.CreateElement("br"))

		// Append the TURN settings container to the app container
		appAppendChild(turnSettings)

		// Toggle TURN settings
		turnButton := jsGo.CreateButton("TURN Settings", func() {
			if turnSettingsVisible{
				turnSettings.Set("hidden", true)
				turnSettingsVisible = false
			} else {
				turnSettings.Set("hidden", false)
				turnSettingsVisible = true
			}
		})
		appAppendChild(turnButton)

		// Form to input room, name, and password
		appAppendChild(jsGo.CreateElement("br"))
		appAppendChild(jsGo.CreateElement("br"))
		appAppendChild(jsGo.CreateElement("br"))
		form := jsGo.CreateElement("form")

		// Room input
		var roomLabel, roomField js.Value
		if !hasRoom {
			roomLabel = jsGo.CreateElement("label")
			roomLabel.Set("textContent", "Chat Room Name:")
			form.Call("appendChild", roomLabel)
			form.Call("appendChild", jsGo.CreateElement("br"))
			roomField = jsGo.CreateElement("input")
			roomField.Set("type", "text")
			roomField.Set("maxLength", "24")
			roomField.Set("pattern", "[\\x00-\\x7f]*")
			roomField.Set("title", "1 to 24 ANSI characters")
			roomField.Set("required", true)
			form.Call("appendChild", roomField)
			form.Call("appendChild", jsGo.CreateElement("br"))
		}

		// Name input
		nameLabel := jsGo.CreateElement("label")
		nameLabel.Set("textContent", "User Name:")
		form.Call("appendChild", nameLabel)
		form.Call("appendChild", jsGo.CreateElement("br"))
		nameField := jsGo.CreateElement("input")
		nameField.Set("required", true)
		form.Call("appendChild", nameField)
		form.Call("appendChild", jsGo.CreateElement("br"))

		// Password input
		var passwordLabel, passwordField js.Value
		if !hasPassword {
			passwordLabel = jsGo.CreateElement("label")
			passwordLabel.Set("textContent", "Chat Room Password:")
			form.Call("appendChild", passwordLabel)
			form.Call("appendChild", jsGo.CreateElement("br"))
			passwordField = jsGo.CreateElement("input")
			passwordField.Set("type", "password")
			passwordField.Set("required", true)
			form.Call("appendChild", passwordField)
			form.Call("appendChild", jsGo.CreateElement("br"))
		}

		// Submit button
		submit := jsGo.CreateElement("input")
		submit.Set("type", "submit")
		submit.Set("value", "Enter chat")
		form.Call("appendChild", submit)
		form.Call("addEventListener", "submit", jsGo.ProcOf(func(event []js.Value) {
			event[0].Call("preventDefault")
			if !hasRoom {room = stringToUrl(roomField.Get("value").String())
			} else {room = stringToUrl(room)}
			name = stringToUrl(nameField.Get("value").String())
			if !hasPassword {password = stringToUrl(passwordField.Get("value").String())
			} else {password = stringToUrl(password)}
			turnUrl = stringToUrl(turnUrlField.Get("value").String())
			turnUser = stringToUrl(turnUserField.Get("value").String())
			turnCred = stringToUrl(turnCredField.Get("value").String())
			policy = stringToUrl(policySelect.Get("value").String())
			var turnSettingsStr string
			if turnUrl != "T" && turnUser != "T" && turnCred != "T" {
				turnSettingsStr = "&turnurl="+turnUrl+"&turnuser="+turnUser+"&turncred="+turnCred+"&policy="+policy
			}
			jsGo.Location.Set(
				"href",
				urlClean+"?room="+room+"&name="+name+"&password="+password+turnSettingsStr,
			)
		}))
		appAppendChild(form)

		// Return from main()
		return
	}

	// Load required JS libraries and begin attempting to connect to the signalling server
	jsGo.LoadJS("https://cdn.jsdelivr.net/npm/peerjs@1.5.5/dist/peerjs.min.js", func() {
		jsGo.LoadJS("https://cdn.jsdelivr.net/npm/argon2-browser@1.18.0/dist/argon2-bundled.min.js", func() {
			jsGo.LoadJS("https://cdn.jsdelivr.net/npm/qrcodejs@1.0.0/qrcode.min.js", func() {

				// Attempt to connect to signalling server
				peer = jsGo.Get("Peer").New(stringToUrl(room)+"0", getOptions())

				// Create listeners to make the rest of the application logic event-driven
				addPeerListeners()
			})
		})
	})
}
