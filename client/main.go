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
)

const (
    IDENTITY = "donnan"
)

func main() {
    serverAddr := net.JoinHostPort(lib.HOST, lib.PORT)
    conn, err := net.Dial(lib.TYPE, serverAddr)
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
    case CMD_WRITE:
        _ = binary.Write(conn, binary.LittleEndian, int16(lib.WRITE))
    }
}

func sendMessage(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SEND_MESSAGE)) // Write opcode to connection
    var sb strings.Builder
    sb.WriteString(os.Args[2]) // Write reciever to stringbuilder
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
    sb.WriteString(IDENTITY)
    sb.WriteRune(lib.TERM_CHAR)
    conn.Write([]byte(sb.String())) // Write reciever to connection
    var numberOfMessages uint16
    _ = binary.Read(conn, binary.LittleEndian, &numberOfMessages) // Read number of messages
    fmt.Println("Messages:", numberOfMessages)
    reader := bufio.NewReader(conn)
    for i := 0; i < int(numberOfMessages); i++ {
        fetchedMessage, _ := reader.ReadString(lib.TERM_CHAR)
        fmt.Println(fetchedMessage)
    }
}

func registerUser(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.REGISTER_USER))
    var sb strings.Builder
    sb.WriteString(IDENTITY)
    sb.WriteRune(lib.TERM_CHAR)
    conn.Write([]byte(sb.String()))
    var response int16
    _ = binary.Read(conn, binary.LittleEndian, &response)
    if response == lib.USER_ADDED {
        fmt.Println("User successfully registered!")
    } else if response == lib.USER_EXISTS {
        fmt.Println("User already registered...")
    }
}
