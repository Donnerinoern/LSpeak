package main

import (
    "fmt"
    "net"
    "os"
    "bufio"
    "strings"
    "encoding/binary"
    "donnan/LSpeak/lib"
)

const (
    CMD_SEND = "send"
    CMD_FETCH = "fetch"
    CMD_WRITE = "write"
    CMD_REGISTER = "register"
    ADM_CMD_DELETE_USER = "DELETE"
    ADM_CMD_SAVE_MESSAGES = "SAVE"
)

var (
    USERNAME string
)

func main() {
    if os.Args[1] != CMD_REGISTER {
        logIn()
    }
    conn, err := net.Dial(lib.TYPE, net.JoinHostPort(lib.HOST, lib.PORT))
    if err != nil {
        fmt.Println("Error: ", err)
        return
    }
    defer conn.Close()
    // Connection has been established here
    switch os.Args[1] {
    case CMD_SEND:  
        sendMessage(conn)
    case CMD_FETCH:
        fetchMessages(conn)
    case CMD_REGISTER:
        registerUser(conn) 
    // case ADM_CMD_DELETE_USER:
    //     adminDeleteUser(conn)
    case ADM_CMD_SAVE_MESSAGES:
        _ = binary.Write(conn, binary.LittleEndian, int16(lib.ADM_SAVE_MESSAGES))
    case CMD_WRITE:
        _ = binary.Write(conn, binary.LittleEndian, int16(lib.WRITE))
    }
}

func sendMessage(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SEND_MESSAGE)) // Write opcode to connection
    var sb strings.Builder
    sb.WriteString(USERNAME) // Write author to stringbuilder
    sb.WriteRune('|')
    sb.WriteString(os.Args[2]) // Write recipient to stringbuilder
    sb.WriteRune('|')
    sb.WriteString(os.Args[3]) // Write message to stringbuilder
    sb.WriteRune(lib.TERM_CHAR)    // Write TERM_CHAR to stringbuilder
    _, err := conn.Write([]byte(sb.String())) // Write reciever and message to the connection
    fmt.Println("Sent:", sb.String())
    if err != nil {
        fmt.Println("Error: ", err)
    }
}

func fetchMessages(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.FETCH_MESSAGES)) // Write OpCode to connection
    var sb strings.Builder
    sb.WriteString(USERNAME)
    sb.WriteRune(lib.TERM_CHAR)
    conn.Write([]byte(sb.String())) // Write recipient to connection
    var numberOfMessages uint16
    _ = binary.Read(conn, binary.LittleEndian, &numberOfMessages) // Read number of messages
    fmt.Println("Messages fetched:", numberOfMessages)
    reader := bufio.NewReader(conn)
    messageBuffer := make([]string, numberOfMessages)
    for i := 0; i < int(numberOfMessages); i++ {
        fetchedMessage, _ := reader.ReadString(lib.TERM_CHAR)
        messageBuffer = append(messageBuffer, fetchedMessage)
        fmt.Println(fetchedMessage)
    }
}

func registerUser(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.REGISTER_USER))
    var sb strings.Builder
    sb.WriteString(os.Args[2])
    sb.WriteRune(lib.TERM_CHAR)
    conn.Write([]byte(sb.String()))
    var response int16
    _ = binary.Read(conn, binary.LittleEndian, &response)
    if response == lib.USER_ADDED {
        file, _ := os.OpenFile("session.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(os.Args[2]+"\n")
        fmt.Printf("User \"%s\" successfully registered!\n", os.Args[2])
        file.Close()
    } else if response == lib.USER_EXISTS {
        fmt.Println("User already registered...")
    }
}

func logIn() {
    file, err := os.Open("session.txt")
    if err != nil {
        fmt.Println("You need to register a user!")
        os.Exit(1)
    }
    scanner := bufio.NewScanner(file)
    scanner.Scan()
    USERNAME = scanner.Text()
    file.Close()
    fmt.Println("Logged in as:", USERNAME)
}

// func adminDeleteUser(conn net.Conn) {
//     // os.Remove("session.txt")
//     _ = binary.Write(conn, binary.LittleEndian, int16(lib.ADM_DELETE_USER))
// }
