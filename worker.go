package selina

import "context"

//Worker is standard interface implemented by proccessors, is used to build pipeline nodes
type Worker interface {
	//Process must close write only channel
	Process(context.Context, <-chan []byte, chan<- []byte) error
}
