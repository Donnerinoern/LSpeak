package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
    "strings"
    "time"
    "os"
    "slices"
    "donnan/LSpeak/lib"
)

var (
    userBuffers [][]string // A slice with string slices for buffering the messages to recipients
)

func main() {
    fetchUsers()
    listener, err := net.Listen(lib.TYPE, lib.HOST+":"+lib.PORT) // Listen on port: PORT
    if err != nil {
        fmt.Println(err)
        return
    } else {
        fmt.Println("Listening on port: " + lib.PORT)
    }
    defer listener.Close()
    for { // Event-loop
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error: ", err)
            continue
        }
        go handleClient(conn)
    }
}

func handleClient(conn net.Conn) {
    var opCode int16
    _ = binary.Read(conn, binary.LittleEndian, &opCode)
    reader := bufio.NewReader(conn)
    switch opCode {
    case lib.SEND_MESSAGE:
        recieveMessage(*reader, conn)
    case lib.FETCH_MESSAGES:
        sendMessages(*reader, conn)
    case lib.REGISTER_USER:
        addUser(*reader, conn)
    case lib.ADM_SAVE_MESSAGES:
        saveMessages()
    // case lib.ADM_DELETE_USER:
    //     removeUser(*reader, conn)
    }
}

func addUser(reader bufio.Reader, conn net.Conn) {
    userText, _ := reader.ReadString(lib.TERM_CHAR)
    file, _ := os.OpenFile("users.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
    for _, userBuffer := range userBuffers {
        if userText == userBuffer[0] {
            file.Close()
            _ = binary.Write(conn, binary.LittleEndian, int16(lib.OP_FAILURE))
            return
        }
    }
    file.WriteString(userText + string('\n'))
    file.Close()
    fmt.Println("Added user:", userText)
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.OP_SUCCESS))
    fetchUsers()
}

// func removeUser(reader bufio.Reader, conn net.Conn) {
//     
// }

func fetchUsers() {
    file, _ := os.OpenFile("users.txt", os.O_APPEND | os.O_CREATE | os.O_RDONLY, os.ModePerm)
    scanner := bufio.NewScanner(file)
    users := make([]string, 0) // Make new slice for users
    i := 0                     // Index users
    for scanner.Scan() {
        users = append(users, scanner.Text())
        i++
    }
    file.Close()
    userBuffers = make([][]string, i)
    for i, user := range users {
        userBuffers[i] = append(userBuffers[i], user)
    }
    fmt.Println(userBuffers)
    fmt.Println(i, "users fetched...")
}

func recieveMessage(reader bufio.Reader, conn net.Conn) {
    inputText, _ := reader.ReadString(lib.TERM_CHAR) // Message from client, in format: AUTHOR|RECIPIENT|MESSAGE
    splitString := strings.Split(inputText, "|")
    var sb strings.Builder
    sb.WriteString(time.Now().Format(time.Stamp))
    sb.WriteString(" | ")
    sb.WriteString(splitString[0])
    sb.WriteString(": ")
    sb.WriteString(splitString[2])
    var sbForNullTerm strings.Builder
    sbForNullTerm.WriteString(splitString[1])
    sbForNullTerm.WriteRune(lib.TERM_CHAR)
    success := false
    for i := 0; i < len(userBuffers); i++ {
        if string(userBuffers[i][0]) == sbForNullTerm.String() { // If recipient equals the username of userBuffer
            userBuffers[i] = append(userBuffers[i], sb.String())
            success = true
        }
    }
    if success {
        _ = binary.Write(conn, binary.LittleEndian, int16(lib.OP_SUCCESS))
    } else {
        _ = binary.Write(conn, binary.LittleEndian, int16(lib.OP_FAILURE))
    }
}

func sendMessages(reader bufio.Reader, conn net.Conn) {
    recipient, _ := reader.ReadString(lib.TERM_CHAR) // The user who fetched messages
    for i := 0; i < len(userBuffers); i++ {
        if recipient == userBuffers[i][0] { // If recipient equals name of user in userBuffer
            _ = binary.Write(conn, binary.LittleEndian, int16(len(userBuffers[i])-1)) // Write amount of messages in userBuffer to the connection
            for i, message := range userBuffers[i] {
                if i == 0 {  // If loop is on first index in userBuffer (the name of the user), skip to the next iteration
                    continue // TODO: Maybe do a normal loop (without range) here, so skipping is not required
                } else {
                    conn.Write([]byte(message)) // Write messages to connection
                }
            }
            userBuffers[i] = slices.Delete(userBuffers[i], 1, len(userBuffers[i])) // Empty slice of messages
            return
        }
    }
    _ = binary.Write(conn, binary.LittleEndian, int16(0)) // If user is not registered, respond with amount of messages 0
}

func saveMessages() {
    var numOfMessages int
    for _, userBuffer := range userBuffers {
        if len(userBuffer) <= 1 { // Skip iteration if there are no messages
            continue
        }
        userNameSlice := []byte(userBuffer[0])                                                 // TODO: Figure out a better way to do this
        userNameSlice = slices.Delete(userNameSlice, len(userNameSlice)-1, len(userNameSlice)) // Removes the null character from the username
        _, err := os.Stat("messages") // Check if directory exists
        if err != nil {
            os.Mkdir("messages", os.ModePerm) // If it doesn't, create it
        }
        file, err := os.OpenFile("messages/" + string(userNameSlice) + ".txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
        for i, message := range userBuffer {
            if i == 0 { // Skip iteration on username
                continue
            }
            file.WriteString(message + "\n")
            numOfMessages++
        }
        file.Close()
    }
    fmt.Println(numOfMessages, "messages saved...")
}

// func retrieveMessages() {
// }
