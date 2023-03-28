package common

import "errors"

var (
	ErrNotFound = errors.New("not found")
)

// 假设{}只用于json分界
// FindEndForPeerSepInByteArray 返回分隔符position
func FindEndForPeerSepInByteArray(bytes []byte, startPos int, leftSep byte, rightSep byte) (int, error) {
	stack := make([]byte, 0, len(bytes))
	for i := startPos; i < len(bytes); i++ {
		// 若是{，则直接放入
		if bytes[i] == leftSep {
			stack = append(stack, leftSep)
			continue
		}

		// 若是}, 则去掉左位的
		if bytes[i] == rightSep {
			// 若stack为空，则直接返回i
			if len(stack) == 1 {
				return i, nil
			}
			stack = stack[:len(stack)-1]
		}
	}
	// 返回-1
	return -1, ErrNotFound
}

const (
	NilFlag   = byte(0x00)
	EmptyFlag = byte(0xFF)

	StringSep = byte(0xFF)
	// TODO: 假设分隔符只用于分隔符
	JsonLeftSep   = byte('{')
	JsonRightSep  = byte('}')
	SliceLeftSep  = byte('[')
	SliceRightSep = byte(']')
)

// FindEndForSepInByteArray 字符串以FFFF为边界
// FindEndForSepInByteArray 返回string分隔符position
func FindEndForSepInByteArray(bytes []byte, startPos int, sep byte) int {
	for i := startPos + 1; i < len(bytes); i++ {
		if bytes[i] == sep {
			return i
		}
	}
	return -1
}
