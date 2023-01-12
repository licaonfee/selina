package selina

import (
	"bytes"
)

// SliceAsChannel return a channel that read from an slice
// if autoClose is true , then channel is closed after last message is consummed
func SliceAsChannel[T any](data []T, autoClose bool) chan T {
	retc := make(chan T, len(data))
	go func() {
		for _, d := range data {
			retc <- d
		}
		if autoClose {
			close(retc)
		}
	}()
	return retc
}

// SliceAsChannelOfBuffer return a channel that read from an slice
// if autoClose is true , then channel is closed after last message is consummed
func SliceAsChannelOfBuffer(data []string, autoClose bool) chan *bytes.Buffer {
	retc := make(chan *bytes.Buffer, len(data))
	go func() {
		for _, d := range data {
			buff := GetBuffer()
			buff.WriteString(d)
			retc <- buff
		}
		if autoClose {
			close(retc)
		}
	}()
	return retc
}

// SliceAsChannelRaw same as SliceAsChannel
func SliceAsChannelRaw[T any](data []T, autoClose bool) chan T {
	retc := make(chan T, len(data))
	go func() {
		for _, d := range data {
			retc <- d
		}
		if autoClose {
			close(retc)
		}
	}()
	return retc
}

// ChannelAsSlice read from in channel until is closed
// return an slice with all messages received
func ChannelAsSlice[T any](in <-chan T) []T {
	var ret []T
	for value := range in {
		ret = append(ret, value)
	}
	return ret
}
