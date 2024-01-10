# LSpeak

Simple and easy-to-use CLI client and TCP server for sending and recieving messages, written in Go.
I made this mainly to learn Go, and to have a unique way of sending/recieving messages to/from friends.

Coming from Java, writing this in Go was a pleasure. I know the code isn't good. I know this isn't "optimal" Go code. This is a hobby project, for learning. As of now, the project is full of out-commented code, deprecated code, lacking comments/documentation, and is not in a state where you'd want to use it.
I'll probably keep working on this until there is no more to change or to add. Feel free to use this for whatever you'd like.

### Commands/args
---

- send (string) (string)
- fetch
- signup
- signin
- signout
- users

##### Examples:

`./Client send foo "Hello, Foo. Good morning."`

`./Client fetch`

`./Client users`

`./Client signup`

`./Client signin`

`./Client signout`

`./Client delete`

### OpCodes
---

I don't know if calling these op codes is "right" or "correct". Let me know if you have a more correct alternative.

OpCodes lets the server know what operation is requested. For every connection with a client, the server first reads an integer from the connection. It will then call the function corresponding to this integer/OpCode.

| Name | Integer value |
| ----------- | ----------- |
| SEND_MESSAGE | 0 |
| FETCH_MESSAGES | 1 | 
| FETCH_USERS | 2 |
| SIGN_UP_USER | 3 |
| SIGN_IN_USER | 4 |
| SIGN_OUT_USER | 5 |
| DELETE_USER | 6 |

### Response
---

When a user performs an action with a result, the server will respond with an integer corresponding with the result of the operation. This is either a OP_SUCCESS, or a OP_FAILURE.

| Name | Integer value |
| ---- | ------------- |
| OP_SUCCESS | 0 |
| OP_FAILURE | 1 |
