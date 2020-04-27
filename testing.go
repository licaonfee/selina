package selina

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

func ChannelAsSlice(in <-chan []byte) []string {
	ret := []string{}
	for value := range in {
		ret = append(ret, string(value))
	}
	return ret
}
