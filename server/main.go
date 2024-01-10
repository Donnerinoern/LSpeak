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
    defer conn.Close()
    var opCode int16
    _ = binary.Read(conn, binary.LittleEndian, &opCode)
    reader := bufio.NewReader(conn)
    switch opCode {
    case lib.SEND_MESSAGE:
        recieveMessage(*reader, conn)
    case lib.FETCH_MESSAGES:
        sendMessages(*reader, conn)
    case lib.FETCH_USERS:
        sendUsers(conn)
    case lib.SIGN_UP_USER:
        signUpUser(*reader, conn)
    case lib.SIGN_IN_USER:
        signInUser(*reader, conn)
    case lib.DELETE_USER:
        deleteUser(*reader, conn)
    // case lib.ADM_DELETE_USER:
    //     removeUser(*reader, conn)
    case lib.ADM_SAVE_MESSAGES:
        saveMessages()
    case lib.ADM_RETRIEVE_MESSAGES:
        retrieveMessages()    
    }
}

func signUpUser(reader bufio.Reader, conn net.Conn) {
    username, _ := reader.ReadString(lib.TERM_CHAR)
    username = lib.RemoveTermChar(username)
    password, _ := reader.ReadString(lib.TERM_CHAR) // TODO: Make a function for this..?
    password = lib.RemoveTermChar(password)
    userExists, _ := checkIfUserExists(username)
    if userExists {
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_FAILURE))
    } else {
        file, _ := os.OpenFile(".users.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(username+"\n")
        file.Close()
        _, err := os.Stat("secrets") // Check if directory exists // TODO: Make a function for this..?
        if err != nil {
            os.Mkdir("secrets", os.ModePerm) // If it doesn't, create it
        }
        file, _ = os.OpenFile("secrets/."+username, os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(password+"\n")
        fmt.Println("Added user:", username)
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_SUCCESS))
        retrieveUsers() // TODO: Maybe append userBuffers instead. Unnecessary read from file.
    }
}

func signInUser(reader bufio.Reader, conn net.Conn) {
    username, _ := reader.ReadString(lib.TERM_CHAR)
    username = lib.RemoveTermChar(username)
    password, _ := reader.ReadString(lib.TERM_CHAR) // TODO: Make a function for this..?
    password = lib.RemoveTermChar(password)
    userExists, _ := checkIfUserExists(username)
    if userExists {
        file, _ := os.Open("secrets/."+username)
        scanner := bufio.NewScanner(file)
        scanner.Scan()
        hash := scanner.Text()
        if hash == password {
            _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_SUCCESS))
        } else {
            _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_FAILURE))
        }
    } else {
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_FAILURE))
    }
}

func deleteUser(reader bufio.Reader, conn net.Conn) {
    
}

func retrieveUsers() {
    file, _ := os.OpenFile(".users.txt", os.O_APPEND | os.O_CREATE | os.O_RDONLY, os.ModePerm)
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
        conn.Write([]byte(userBuffer[0] + string(lib.TERM_CHAR)))
    }
}

func recieveMessage(reader bufio.Reader, conn net.Conn) {
    message, _ := reader.ReadString(lib.TERM_CHAR) // Message from client, in format: DATETIME|AUTHOR|RECIPIENT|MESSAGE
    splitMessage := strings.Split(message, "|")    // Should I create a struct/type for messages?
    success := false
    for i := 0; i < len(userBuffers); i++ {
        if string(userBuffers[i][0]) == splitMessage[2] { // If recipient equals the username of userBuffer
            userBuffers[i] = append(userBuffers[i], message)
            success = true
        }
    }
    if success {
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_SUCCESS))
    } else {
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.OP_FAILURE))
    }
}

func sendMessages(reader bufio.Reader, conn net.Conn) { // TODO: This is complicated and can probably be done a better way.
    recipient, _ := reader.ReadString(lib.TERM_CHAR) // The user who fetched messages
    recipient = lib.RemoveTermChar(recipient) // Remove TERM_CHAR from recipient
    authors := make([]string, 0)
    var index int // Save index of userBuffer for later
    for i, userBuffer := range userBuffers { // Loop for finding index and adding authors to []authors
        if recipient == userBuffer[0] {
            index = i
            for i, message := range userBuffer { // TODO: Maybe do a normal loop (without range) here, so skipping is not required
                if i == 0 { // Skip iteration for recipient username
                    continue
                }
                splitMessage := strings.Split(message, string(lib.SEP_CHAR))
                if !slices.Contains(authors, splitMessage[1]) {
                    authors = append(authors, splitMessage[1]) // If author is not in authors, add author to []authors
                }
            }
        }
    }
    _ = binary.Write(conn, binary.LittleEndian, uint32(len(authors))) // Write number of authors to connection
    if len(authors) == 0 { // If there are no messages (no authors counted), return
        return
    }
    messageBuffers := make([][]string, len(authors))
    for i, author := range authors { // Add author username as header (index 0) for messageBuffers
        messageBuffers[i] = append(messageBuffers[i], author)
    }
    for i := 1; i < len(userBuffers[index]); i++ { // Loop through messages in userBuffer, using saved index
        splitMessage := strings.Split(userBuffers[index][i], string(lib.SEP_CHAR))
        for j := 0; j < len(messageBuffers); j++ { // Loop through messageBuffers for every message
            if messageBuffers[j][0] == splitMessage[1] { // If author (header) in messageBuffer == message author, add message to buffer
                messageBuffers[j] = append(messageBuffers[j], userBuffers[index][i])
            }
        }
    }
    for _, messageBuffer := range messageBuffers { // Loop through messageBuffers
        _ = binary.Write(conn, binary.LittleEndian, uint32(len(messageBuffer)-1)) // Write number of messages from author
        conn.Write([]byte(messageBuffer[0]+string(lib.TERM_CHAR))) // Write author to connection
        for i, message := range messageBuffer {
            if i == 0 { // TODO: Again, maybe change loop
                continue
            }
            conn.Write([]byte(message)) // Write message to connection
        }
    }
    userBuffers[index] = slices.Delete(userBuffers[index], 1, len(userBuffers[index])) // Clean up
}

func saveMessages() { // Should this respond with result?
    var numOfMessages int
    for i, userBuffer := range userBuffers { // TODO: slices.Index() may work instead if you can nest them
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

func retrieveMessages() { // TODO: Finish
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

func checkIfUserExists(username string) (bool, int) {
    var index int
    for i, userBuffer := range userBuffers {
        if username == userBuffer[0] {
            index = i
            return true, index
        }
    }
    return false, index
}
