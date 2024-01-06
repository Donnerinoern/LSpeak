package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
    "strings"
    "os"
    "slices"
    "donnan/LSpeak/lib"
)

var (
    userBuffers [][]string // A slice of string slices for buffering messages to recipients
)

func main() {
    retrieveUsers()
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
    case lib.FETCH_USERS:
        sendUsers(conn)
    // case lib.ADM_DELETE_USER:
    //     removeUser(*reader, conn)
    case lib.ADM_SAVE_MESSAGES:
        saveMessages()
    case lib.ADM_RETRIEVE_MESSAGES:
        retrieveMessages()    
    }
}

func addUser(reader bufio.Reader, conn net.Conn) {
    username, _ := reader.ReadString(lib.TERM_CHAR)
    username = lib.RemoveTermChar(username)
    file, _ := os.OpenFile("users.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
    for _, userBuffer := range userBuffers {
        if username == userBuffer[0] {
            file.Close()
            _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_FAILURE))
            return
        }
    }
    file.WriteString(username + string('\n'))
    file.Close()
    fmt.Println("Added user:", username)
    _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_SUCCESS))
    retrieveUsers()
}

// func removeUser(reader bufio.Reader, conn net.Conn) {
//     
// }

func retrieveUsers() {
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
    fmt.Println(i, "users registered...")
}

func sendUsers(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, uint32(len(userBuffers)))
    for _, userBuffer := range userBuffers {
        conn.Write([]byte(userBuffer[0]))
    }
}

func recieveMessage(reader bufio.Reader, conn net.Conn) {
    message, _ := reader.ReadString(lib.TERM_CHAR) // Message from client, in format: DATETIME|AUTHOR|RECIPIENT|MESSAGE
    splitMessage := strings.Split(message, "|")    // Should I create a struct/type for messages?
    // var sb strings.Builder
    // sb.WriteString(splitString[0])
    // sb.WriteString(": ")
    // sb.WriteString(splitString[2])
    // var sbForNullTerm strings.Builder
    // sbForNullTerm.WriteString(splitString[1]) // TODO: Should maybe cleanup username so this isn't needed
    // sbForNullTerm.WriteRune(lib.TERM_CHAR)    // stringbuilder for comparing username with recipient (because of null char)
    success := false
    for i := 0; i < len(userBuffers); i++ {
        if string(userBuffers[i][0]) == splitMessage[2] { // If recipient equals the username of userBuffer
            userBuffers[i] = append(userBuffers[i], message)
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
    recipient = lib.RemoveTermChar(recipient)
    for i := 0; i < len(userBuffers); i++ {
        if recipient == userBuffers[i][0] { // If recipient equals name of user in userBuffer
            _ = binary.Write(conn, binary.LittleEndian, uint32(len(userBuffers[i])-1)) // Write number of messages in userBuffer to the connection
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

func saveMessages() { // Should this respond with result?
    var numOfMessages int
    for i, userBuffer := range userBuffers {
        if len(userBuffer) <= 1 { // Skip iteration if there are no messages
            continue
        }
        _, err := os.Stat("messages") // Check if directory exists
        if err != nil {
            os.Mkdir("messages", os.ModePerm) // If it doesn't, create it
        }
        file, err := os.OpenFile("messages/" + string(userBuffer[0]) + ".txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
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
        userBuffers[i] = slices.Delete(userBuffers[i], 1, len(userBuffers[i])) // Empty message buffer/slice of messages
    }
    fmt.Println(numOfMessages, "messages saved...")
}

func retrieveMessages() { // TODO: Format messages either when client recieves, or when server sends
    dirSlice, err := os.ReadDir("messages")
    if err != nil {
        fmt.Println("Error:", err)
    }
    for _, file := range dirSlice {
        openedFile, _ := os.Open("messages/" + file.Name())
        scanner := bufio.NewScanner(openedFile)
        for scanner.Scan() {
            fmt.Println(scanner.Text())
        }
    }
}
