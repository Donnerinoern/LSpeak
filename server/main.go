package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
    "strings"
    "time"
    "os"
    "donnan/LSpeak/lib"
)

var (
    userBuffers [][]string
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
    // Event-loop
    for {
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
        recieveMessage(*reader)
    case lib.FETCH_MESSAGES:
        sendMessages(*reader, conn)
    case lib.REGISTER_USER:
        addUser(*reader, conn)
    // case WRITE:
    //     writeToFile()
    }
}

// func writeToFile() { // FIXME: This is horrible. Do this in a better way.
//     // if != nil {
//     //     file, err := os.OpenFile(".txt", os.O_APPEND | os.O_CREATE, os.ModeAppend)
//     //     if err != nil {
//     //         fmt.Println("Error: ", err)
//     //     }
//     //     for _, message := range {
//     //         file.WriteString(message)
//     //     }
//     //     = make([]string, 0)
//     // }
//     if donnan != nil {
//         fmt.Println("Writing...")
//         file, err := os.OpenFile("donnan.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm) 
//         if err != nil {
//             fmt.Println("Error: ", err)
//         }
//         var sb strings.Builder
//         for _, message := range donnan {
//             sb.WriteString(message)
//             sb.WriteRune('\n')
//             file.WriteString(sb.String())
//             sb.Reset()
//         }
//         donnan = make([]string, 0)
//     }
// }

func addUser(reader bufio.Reader, conn net.Conn) {
    userText, _ := reader.ReadString(lib.TERM_CHAR)
    file, _ := os.OpenFile("users.txt", os.O_APPEND | os.O_CREATE | os.O_WRONLY, os.ModePerm)
    for _, userBuffer := range userBuffers {
        if userText == userBuffer[0] {
            file.Close()
            _ = binary.Write(conn, binary.LittleEndian, int16(lib.USER_EXISTS))
            return
        }
    }
    file.WriteString(userText + string('\n'))
    file.Close()
    fmt.Println("Added user:", userText)
    _ = binary.Write(conn, binary.LittleEndian, int16(lib.USER_ADDED))
    fetchUsers()
}

func fetchUsers() {
    file, _ := os.OpenFile("users.txt", os.O_APPEND | os.O_CREATE | os.O_RDONLY, os.ModePerm)
    scanner := bufio.NewScanner(file)
    users := make([]string, 0) // Make new slice for users
    i := 0                     // Index users
    for scanner.Scan() {
        fmt.Println("User:", scanner.Text())
        users = append(users, scanner.Text())
        i++
    }
    file.Close()
    userBuffers = make([][]string, i)
    for i, user := range users {
        userBuffers[i] = append(userBuffers[i], user)
    }
    // userBuffers[0] = append(userBuffers[0], "Test!")
    // userBuffers[0] = append(userBuffers[0], "Mer test!")
    // userBuffers[0] = append(userBuffers[0], "Test mer!")
    fmt.Println(userBuffers)
    fmt.Println("Users fetched...")
}

func recieveMessage(reader bufio.Reader) {
    inputText, _ := reader.ReadString(lib.TERM_CHAR)
    splitString := strings.Split(inputText, "|")
    var sb strings.Builder
    sb.WriteString(time.Now().Format(time.Stamp))
    sb.WriteString(" | ")
    sb.WriteString(splitString[1])
    var sbForNullTerm strings.Builder
    sbForNullTerm.WriteString(splitString[0])
    sbForNullTerm.WriteRune(lib.TERM_CHAR)
    for i := 0; i < len(userBuffers); i++ {
        if string(userBuffers[i][0]) == sbForNullTerm.String() {
            userBuffers[i] = append(userBuffers[i], sb.String())
        }
    }
}

func sendMessages(reader bufio.Reader, conn net.Conn) {
    reciever, _ := reader.ReadString(lib.TERM_CHAR)
    // var sb strings.Builder
    // if reciever == {
    //     _ = binary.Write(conn, binary.LittleEndian, uint16(len()))
    //     for _, element := range {
    //         sb.WriteString(element)
    //         conn.Write([]byte(sb.String()))
    //         sb.Reset()
    //     }
    //     = make([]string, 0)
    //     printSlices()
    // } else if reciever == DONNAN {
    //     _ = binary.Write(conn, binary.LittleEndian, uint16(len(donnan)))
    //     for _, element := range donnan {
    //         sb.WriteString(element)
    //         conn.Write([]byte(sb.String()))
    //         sb.Reset()
    //     }
    //     donnan = make([]string, 0)
    //     printSlices()
    // } else {
    //     fmt.Println("Unknown reciever. Exiting...")
    //     return
    // }
    for _, userBuffer := range userBuffers {
        if reciever == userBuffer[0] {
            _ = binary.Write(conn, binary.LittleEndian, int16(len(userBuffer)-1))
            // fmt.Println(userBuffer)
            // fmt.Println(len(userBuffer)-1)
            for i, message := range userBuffer {
                if i == 0 {  // If loop is on first index in userBuffer, skip to the next iteration
                    continue // TODO: Maybe do a normal loop (without range) here, so skipping is not required
                } else {
                    conn.Write([]byte(message))
                }
            }
        }
    }
}
