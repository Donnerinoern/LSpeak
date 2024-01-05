package lib

const (
    HOST = "192.168.10.126"
    PORT = "25565"
    TYPE = "tcp"
)

const (
    SEND_MESSAGE = iota
    FETCH_MESSAGES
    REGISTER_USER
    FETCH_USERS
    ADM_DELETE_USER
    ADM_SAVE_MESSAGES
    ADM_RETRIEVE_MESSAGES
)

const (
    OP_SUCCESS = iota
    OP_FAILURE
)

const (
    TERM_CHAR = '\x00'
    SEP_CHAR = '|'
)
