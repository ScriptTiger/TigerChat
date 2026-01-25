[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

**DISCLAIMER!!!: THIS APP IS STILL IN ITS EARLY DEVELOPMENT AND HAS NOT BEEN AUDITED FOR SECURITY, SO USE AT YOUR OWN RISK!**

# TigerChat (https://scripttiger.github.io/tigerchat/)
TigerChat is a simple peer-to-peer/full-mesh chat app written in Go and transpiled to JavaScript via GopherJS. Every connection is inherently encrypted by WebRTC, and every peer must successfully perform a three-way cryptographic challenge-response handshake using Argon2 with each of the other peers in order to be authenticated with each other. However, while this provides reasonable security for fun chats amongst friends, it's not intended for the transmission of highly sensitive data.

If you are experiencing difficulties connecting to peers, TigerChat is also easily configurable for TURN in cases where your network may be preventing you from establishing peer-to-peer connections. However, while TURN may relay your traffic similarly to a proxy, it should be noted that using TURN does not make you anonymous. The point of TURN is to facilitate real-time communications in instances where peer-to-peer connections cannot be established, not to provide anonymity.

So, again, this is not intended to be a secure, anonymous chat app. It's just intended as a way to quickly and easily chat amongst friends.

If you don't already have access to a TURN server, you can sign up for ExpressTURN (https://www.expressturn.com) absloutely free, go to your dashboard, and you'll be presented with your TURN information which you can use to configure TigerChat for TURN. There are obviously a lot of different options out there for this, but I've found this to be the simplest for folks who are not too tech-savvy, as you can literally sign up and get your information for TURN access all in just a couple of minutes or less.

But...why written in Go?!?! Go is clearly a bit clunky when it comes to DOM manipulation. However, the advantages of being able to use Go to compile to native server-side apps (Go), to WASM (Go/TinyGo), and also transpile to JS (GopherJS) make it highly attractive for this portability and not having to context-switch when writing code for full-stack components which may be anywhere in that stack. This also means less of a learning curve and more efficiency with small teams all using the same language for everything.

To further smooth out some of that clunkiness, the `jsGo` package has also been used in order to simplify DOM manipulation slightly, but also to simplify interacting with the native JS API in general to avoid using the Go standard library wherever possible. Avoiding the Go standard library avoids the inevitable bloat which comes with the multi-layered abstractions needed to support the Go standard library within the JS context. And while large file sizes continue to be hailed as a pain point, file sizes can be dramatically reduced by both avoiding the standard library and also using an optimizer, such as Terser. However, if you do use Terser, it's recommended NOT to use the `-m`/`--minify` argument with GopherJS so that Terser can work more efficiently at reducing the output more than GopherJS is currently capable of doing.

For additional notes on `jsGo`, please refer to its documentation:  
https://github.com/ScriptTiger/jsGo

As an example of file size, the resulting `TigerChat.min.js` itself only amounts to around 80 KB, which is a reasonable size for such a simple app, and considering it also contains the GopherJS Go runtime. However, TigerChat does call `jsGo.LoadJS` to load external JS libraries at run time and no external libraries are bundled together with TigerChat, also reducing the file size. But if you do want to bundle external libraries together with your GopherJS apps, GopherJS makes that easy too by automatically including any files in the project directory ending with `.inc.js`. And when doing that, it's recommended to NOT use external libraries which have already been minified if you intend to run Terser on the resulting bundle so that it can work more efficiently to reduce the size of the bundle as a whole.

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/
