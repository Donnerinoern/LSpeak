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

const ( // Commands/args
    CMD_SEND = "send"   // TODO: Error check all commands that take two or more args
    CMD_FETCH = "fetch" // Should fetch take arg[2] as a subcommand for fetching messages/users?
    CMD_REGISTER = "register"
    CMD_USERS = "users"
    ADM_CMD_DELETE_USER = "DELETE"
    ADM_CMD_SAVE_MESSAGES = "SAVE"  
    ADM_CMD_RETRIEVE_MESSAGES = "RETRIEVE"
)

var (
    USERNAME string
)

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Please provide a command.")
        return
    }
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
    case CMD_USERS:
        fetchUsers(conn)
    // case ADM_CMD_DELETE_USER:
    //     adminDeleteUser(conn)
    case ADM_CMD_SAVE_MESSAGES:
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.ADM_SAVE_MESSAGES))
    case ADM_CMD_RETRIEVE_MESSAGES:
        _ = binary.Write(conn, binary.LittleEndian, uint8(lib.ADM_RETRIEVE_MESSAGES))
    }
}

func sendMessage(conn net.Conn) { 
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SEND_MESSAGE)) // Write opcode to connection
    formattedMessage := lib.FormatMessage(USERNAME, os.Args[2], os.Args[3])
    _, err := conn.Write([]byte(formattedMessage)) // Write formatted message (DATETIME|AUTHOR|RECIPIENT|MESSAGE) to connection
    var response uint8                             // Use a different seperation character? Currently uses pipe (|)
    _ = binary.Read(conn, binary.LittleEndian, &response) // Get a response from the server
    if response == lib.OP_SUCCESS {
        fmt.Printf("Sent to %s: %s\n", os.Args[2], os.Args[3])
        fmt.Println(formattedMessage) // TODO: Remove
    } else {
        fmt.Println("User", os.Args[2], "does not exist!")
    }
    if err != nil {
        fmt.Println("Error: ", err)
    }
}

func fetchMessages(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.FETCH_MESSAGES)) // Write OpCode to connection
    conn.Write([]byte(USERNAME + string(lib.TERM_CHAR))) // Write recipient (client's username) and TERM_CHAR to connection
    var numberOfMessages uint32
    _ = binary.Read(conn, binary.LittleEndian, &numberOfMessages) // Read number of messages
    fmt.Println("Messages fetched:", numberOfMessages)
    reader := bufio.NewReader(conn)
    var sb strings.Builder
    messageBuffer := make([]string, numberOfMessages)
    for i := 0; i < int(numberOfMessages); i++ { // For each message
        fetchedMessage, _ := reader.ReadString(lib.TERM_CHAR)
        splitMessage := strings.Split(fetchedMessage, "|")
        sb.WriteString(splitMessage[0])
        sb.WriteString(" | ")
        sb.WriteString(splitMessage[1])
        sb.WriteString(": ")
        sb.WriteString(splitMessage[3])
        fmt.Println(sb.String())
        messageBuffer = append(messageBuffer, sb.String()) // TODO: Sort messages by author and display them sorted
        sb.Reset()
    }
}

func registerUser(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.REGISTER_USER))
    var sb strings.Builder
    sb.WriteString(os.Args[2])
    sb.WriteRune(lib.TERM_CHAR)
    conn.Write([]byte(sb.String()))
    var response uint8
    _ = binary.Read(conn, binary.LittleEndian, &response)
    if response == lib.OP_SUCCESS {
        file, _ := os.OpenFile("session.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(os.Args[2]+"\n")
        fmt.Printf("User \"%s\" successfully registered!\n", os.Args[2])
        file.Close()
    } else if response == lib.OP_SUCCESS {
        fmt.Println("User already registered...")
    }
}

func fetchUsers(conn net.Conn) {
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.FETCH_USERS))
    var numberOfUsers uint32
    _ = binary.Read(conn, binary.LittleEndian, &numberOfUsers)
    if numberOfUsers == 0 {
        fmt.Println("No users registered...")
        return
    }
    var sb strings.Builder
    sb.WriteString("Registered users:\n")
    reader := bufio.NewReader(conn)
    for i := 0; i < int(numberOfUsers); i++ {
        fetchedUser, _ := reader.ReadString(lib.TERM_CHAR)
        sb.WriteString(fetchedUser)
        sb.WriteRune('\n')
    }
    fmt.Print(sb.String())
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
