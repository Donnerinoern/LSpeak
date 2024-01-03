package lib

const (
    HOST = "192.168.10.126"
    PORT = "25565"
    TYPE = "tcp"
    TERM_CHAR = '\x00'
)

const (
    SEND_MESSAGE = iota
    FETCH_MESSAGES
    REGISTER_USER
    ADM_DELETE_USER
    ADM_SAVE_MESSAGES
    WRITE
)

const (
    USER_ADDED = iota
    USER_EXISTS
)
