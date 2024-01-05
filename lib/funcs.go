package lib

import (
    "time"
    "strings"
    "slices"
)

func FormatMessage(author, recipient, message string) string {
    var sb strings.Builder
    sb.WriteString(time.Now().Format(time.Stamp))
    sb.WriteRune(SEP_CHAR)
    sb.WriteString(author) // Write author to stringbuilder
    sb.WriteRune(SEP_CHAR)
    sb.WriteString(recipient) // Write recipient to stringbuilder
    sb.WriteRune(SEP_CHAR)
    sb.WriteString(message) // Write message to stringbuilder
    sb.WriteRune(TERM_CHAR)    // Write TERM_CHAR to stringbuilder
    return sb.String()
}

func RemoveTermChar(input string) string {
    inputSlice := []byte(input)
    inputSlice = slices.Delete(inputSlice, len(inputSlice)-1, len(inputSlice))
    return string(inputSlice)
}
