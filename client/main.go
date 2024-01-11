package main

import (
    "fmt"
    "net"
    "os"
    "log"
    "bufio"
    "strings"
    "encoding/binary"
    "encoding/hex"
    "crypto/sha512"
    "donnan/LSpeak/lib"
)

const ( // Commands/args
    CMD_SEND = "send" 
    CMD_FETCH = "fetch" // TODO: Should fetch take arg[2] as a subcommand for fetching messages/users?
    CMD_USERS = "users"
    CMD_SIGN_IN = "signin"
    CMD_SIGN_UP = "signup"
    CMD_SIGN_OUT = "signout"
    CMD_DELETE_USER = "delete"
    ADM_CMD_SAVE_MESSAGES = "SAVE"         // Works 
    ADM_CMD_RETRIEVE_MESSAGES = "RETRIEVE" // Doesn't work
)

const (
    SESSION_FILE = ".session.txt"
)

var (
    USERNAME string
    PASSWORD string
)

func main() {
    if len(os.Args) <= 1 { // If no arg/command
        fmt.Println("Please provide a command.")
        return
    }
    // if os.Args[1] == CMD_SIGN_OUT {
    //     notLoggedIn := signOut()
    //     if notLoggedIn {
    //         fmt.Println("Couldn't sign out as you're not signed in.")
    //     } else {
    //         fmt.Println("Successfully signed out.")
    //     }
    //     return
    // } else if os.Args[1] == CMD_SIGN_IN || os.Args[1] ==  CMD_SIGN_UP {
    //     _, err := os.Open(SESSION_FILE)
    //     if err == nil {
    //         fmt.Println("You're already signed in.")
    //         return
    //     } else if os.Args[1] == CMD_SIGN_IN {
    //         conn := makeConnection()
    //         defer conn.Close()
    //         signIn()
    //         return
    //     } else if os.Args[1] == CMD_SIGN_UP {
    //         conn := makeConnection()
    //         defer conn.Close()
    //         signUp()
    //         return
    //     }
    // }
    switch os.Args[1] {
    case CMD_SIGN_IN:
        signIn()
        return
    case CMD_SIGN_UP:
        signUp()
        return
    case CMD_SIGN_OUT:
        signOut()
        return
    }
    isSignedIn := isSignedIn(os.Args[1] == CMD_SIGN_IN)
    if isSignedIn {
        switch os.Args[1] {
        case CMD_SEND: 
            if len(os.Args) < 4 {
                fmt.Println("Please provide a username and a message.")
            } else {
                sendMessage()
            }
        case CMD_FETCH:
            fetchMessages()
        case CMD_USERS:
            fetchUsers()
        case CMD_DELETE_USER:
            deleteUser()
        default:
            fmt.Println("Please provide a valid command.")
            fmt.Printf("Commands:\n%s\n%s\n%s\n%s\n%s\n%s\n%s", CMD_SEND, CMD_FETCH, CMD_USERS, CMD_SIGN_UP, CMD_SIGN_IN, CMD_SIGN_OUT, CMD_DELETE_USER)
        }
    }
}

func sendMessage() {
    if len(os.Args) < 3 {
        fmt.Println("Please provide a user and a message.")
        return
    }
    conn := makeConnection()
    defer conn.Close()
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SEND_MESSAGE)) // Write opcode to connection
    formattedMessage := lib.FormatMessage(USERNAME, os.Args[2], os.Args[3])
    _, err := conn.Write([]byte(formattedMessage)) // Write formatted message (DATETIME|AUTHOR|RECIPIENT|MESSAGE) to connection
    var response uint8                             // Use a different seperation character? Currently uses pipe (|)
    _ = binary.Read(conn, binary.LittleEndian, &response) // Get a response from the server
    if response == lib.OP_SUCCESS {
        fmt.Printf("Sent to %s: %s\n", os.Args[2], os.Args[3])
    } else {
        fmt.Println("User", os.Args[2], "does not exist!")
    }
    if err != nil {
        fmt.Println("Error:", err)
    }
}

func fetchMessages() {
    conn := makeConnection()
    defer conn.Close()
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.FETCH_MESSAGES)) // Write OpCode to connection
    conn.Write([]byte(USERNAME + string(lib.TERM_CHAR))) // Write recipient (client's username) and TERM_CHAR to connection
    var numberOfAuthors uint32
    _ = binary.Read(conn, binary.LittleEndian, &numberOfAuthors) // Read number of authors
    if numberOfAuthors == 0 {
        fmt.Println("No new messages...")
        return
    } else {
        reader := bufio.NewReader(conn)
        for i := 0; i < int(numberOfAuthors); i++ {
            var numberOfMessages uint32
            _ = binary.Read(conn, binary.LittleEndian, &numberOfMessages)
            author, _ := reader.ReadString(lib.TERM_CHAR)
            fmt.Printf("Messages from %s:\n", author)
            for i := 0; i < int(numberOfMessages); i++ {
                fetchedMessage, _ := reader.ReadString(lib.TERM_CHAR)
                formattedMessage := formatIncomingMessage(fetchedMessage)
                fmt.Println(formattedMessage)
            }
        }
    }
}

func fetchUsers() { // TODO: Should users be able to hide?
    conn := makeConnection()
    defer conn.Close()
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
        sb.WriteString(fetchedUser+"\n")
    }
    fmt.Print(sb.String())
}

func signUp() bool {
    _, err := os.Open(SESSION_FILE)
    if err == nil {
        fmt.Println("You're already signed in.")
        return false
    }
    var username string
    var password string
    fmt.Print("Username: ")
    fmt.Scanln(&username)
    for {
        fmt.Print("Password: ")
        fmt.Scanln(&password)
        fmt.Print("Confirm password: ")
        var confirmPassword string
        fmt.Scanln(&confirmPassword)
        if password == confirmPassword {
            break
        } else {
            fmt.Println("Passwords did not match...") // TODO: Maybe clear here, and print "Username: *username*"
        }
    }
    hashedPassword := sha512.Sum512_256([]byte(password))
    hexHash := hex.EncodeToString(hashedPassword[:])
    conn := makeConnection()
    defer conn.Close()
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SIGN_UP_USER))
    conn.Write([]byte(username+string(lib.TERM_CHAR)))
    conn.Write([]byte(hexHash+string(lib.TERM_CHAR)))
    var response uint8
    _ = binary.Read(conn, binary.LittleEndian, &response)
    if response == lib.OP_SUCCESS {
        file, _ := os.OpenFile(SESSION_FILE, os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(username+"\n"+hexHash+"\n")
        fmt.Printf("User \"%s\" successfully signed up!\n", username)
        fmt.Println("You are now logged in as:", username)
        file.Close()
        return true
    } else {
        fmt.Println("User already registered...")
        return false
    }
}

func isSignedIn(cmdIsLogIn bool) bool {
    file, err := os.Open(SESSION_FILE)
    if err != nil && !cmdIsLogIn { // If SESSION_FILE does not exists and command is not "login"
        fmt.Println("You are not logged in.")
        for {
            fmt.Print("Do you want to sign in or sign up now? sign(in)/sign(up)/(C)ancel ")
            var input string
            fmt.Scanln(&input)
            input = strings.ToLower(input)
            if input == "in" || input == "signin" {
                return signIn()
            } else if input == "up" || input == "signup" {
                return signUp()
            } else if input == "c" || input == "" {
                return false
            } else {
                fmt.Println("Please provide a valid choice.")
            }
        }
    } else if cmdIsLogIn {
        return signIn()
    }
    scanner := bufio.NewScanner(file)
    scanner.Scan()
    USERNAME = scanner.Text()
    scanner.Scan()
    PASSWORD = scanner.Text()
    file.Close()
    fmt.Println("Signed in as:", USERNAME)
    return true
}

func signIn() bool {
    _, err := os.Open(SESSION_FILE)
    if err == nil {
        fmt.Println("You're already signed in.")
        return false
    }
    var username string
    fmt.Print("Username: ")
    fmt.Scanln(&username)
    var password string
    fmt.Print("Password: ")
    fmt.Scanln(&password)
    hashedPassword := sha512.Sum512_256([]byte(password))
    hexHash := hex.EncodeToString(hashedPassword[:])
    conn := makeConnection()
    defer conn.Close()
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.SIGN_IN_USER))
    conn.Write([]byte(username+string(lib.TERM_CHAR)))
    conn.Write([]byte(hexHash+string(lib.TERM_CHAR)))
    var response uint8
    _ = binary.Read(conn, binary.LittleEndian, &response)
    if response == lib.OP_SUCCESS {
        file, _ := os.OpenFile(".session.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
        file.WriteString(username+"\n"+hexHash+"\n")
        file.Close()
        fmt.Println("Successfully signed in!")
        return true
    } else {
        fmt.Println("Could not log in.")
        return false
    }
}

func signOut() bool {
    err := os.Remove(SESSION_FILE)
    if err != nil {
        fmt.Println("You're not signed in.")
        return false
    } else {
        fmt.Println("Successfully signed out.")
        return true
    }
}

func deleteUser() {
    conn := makeConnection()
    defer conn.Close()
    _ = binary.Write(conn, binary.LittleEndian, lib.DELETE_USER)
    conn.Write([]byte(USERNAME+string(lib.TERM_CHAR)))
    conn.Write([]byte(PASSWORD+string(lib.TERM_CHAR)))
    var result uint8
    _ = binary.Read(conn, binary.LittleEndian, &result)
    if result == lib.OP_SUCCESS {
        signOut()
    } else {
        fmt.Println("Something went wrong.")
    }
}

func makeConnection() net.Conn {
    conn, err := net.Dial(lib.TYPE, net.JoinHostPort(lib.HOST, lib.PORT))
    if err != nil {
        log.Fatal("Error:", err)
    }
    return conn
}

func formatIncomingMessage(message string) string {
    splitMessage := strings.Split(message, "|")
    var sb strings.Builder
    sb.WriteString(splitMessage[0])
    sb.WriteString(" | ")
    sb.WriteString(splitMessage[1])
    sb.WriteString(": ")
    sb.WriteString(splitMessage[3])
    return sb.String()
}
