package selina

//SliceAsChannel return a channel that read from an slice
// if autoClose is true , then channel is closed after last message is consummed
func SliceAsChannel(data []string, autoClose bool) chan []byte {
	retc := make(chan []byte, len(data))
	go func() {
		for _, d := range data {
			retc <- []byte(d)
		}
		if autoClose {
			close(retc)
		}
	}()
	return retc
}

//SliceAsChannelRaw same as SliceAsChannel
func SliceAsChannelRaw(data [][]byte, autoClose bool) chan []byte {
	retc := make(chan []byte, len(data))
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

//ChannelAsSlice read from in channel until is closed
// return an slice with all messages received
func ChannelAsSlice(in <-chan []byte) []string {
	ret := []string{}
	for value := range in {
		ret = append(ret, string(value))
	}
	return ret
}
