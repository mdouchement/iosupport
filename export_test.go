package iosupport

// This file is only for test purpose and is only loaded by test framework

// SetToken sets the token of the given TsvParser
func SetToken(tp *TsvParser, token []byte) {
	tp.token = token
}

// ParseFields parses fileds for the current token defined in the given TsvParser
func ParseFields(tp *TsvParser) [][]byte {
	return tp.parseFields()
}
