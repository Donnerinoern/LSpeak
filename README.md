# LSpeak

Simple and easy-to-use CLI client and TCP server for sending and recieving messages, written in Go.
I made this mainly to learn Go, and to have a unique way of sending/recieving messages to/from friends.

Coming from Java, writing this in Go was a pleasure. I know the code isn't good. I know this isn't "optimal" Go code. This is a hobby project, for learning. As of now, the project is full of out-commented code, deprecated code, lacking comments/documentation, and is not in a state where you'd want to use it.
I'll probably keep working on this until there is no more to change or to add. Feel free to use this for whatever you'd like.

### Commands/args
---

- send (takes in a string in quotation marks)
- fetch (none)
- register (string)
- users (none)

##### Examples:

`lsc send linus "Linux <3"`

`lsc fetch`

`lsc register`

`lsc users`

### OpCodes
---

I don't know if calling these op codes is "right" or "correct". Let me know if you have a more correct alternative.

OpCodes lets the server know what operation is requested. For every connection with a client, the server first reads an integer from the connection. It will then call the function corresponding to this integer/OpCode.

| Name | Integer value |
| ----------- | ----------- |
| SEND_MESSAGE | 0 |
| FETCH_MESSAGES | 1 | 
| REGISTER_USER | 2 |
| FETCH_USERS | 3 |

WRITE is deprecated/removed and was only used for testing.

### Response
---

When a user tries to register or send a message to a unregistered user, the server will respond with an integer corresponding with the result of the operation. This is either a OP_SUCCESS, or a OP_FAILURE.

| Name | Integer value |
| ---- | ------------- |
| OP_SUCCESS | 0 |
| OP_FAILURE | 1 |
