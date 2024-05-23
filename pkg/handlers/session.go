package handlers

import (
	"time"
	"unsafe"
)

type session struct {
	id        uint64
	entryTime time.Time
}

const (
	sessionBinarySize = unsafe.Sizeof(session{})
)

// NOTE: due to use of unsafe cast, prisme instances should all share the same
// endianness.

func unsafeSessionToBytesCast(session *session) []byte {
	bytes := (*[sessionBinarySize]byte)(unsafe.Pointer(session))
	return (*bytes)[:]
}

func unsafeBytesToSessionCast(rawSession []byte) *session {
	bytes := [sessionBinarySize]byte(rawSession)
	return (*session)(unsafe.Pointer(&bytes))
}
