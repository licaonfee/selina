package selina_test

import (
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
)

func TestSliceAsChannel(t *testing.T) {
	data := []string{"foo", "bar", "baz"}
	c := selina.SliceAsChannel(data, true)
	got := []string{}
	for v := range c {
		got = append(got, string(v))
	}
	if !reflect.DeepEqual(got, data) {
		t.Fatalf("SliceAsChannel() got = %v , want = %v", got, data)
	}
}

func TestSliceAsChannelRaw(t *testing.T) {
	data := [][]byte{[]byte("foo"), []byte("bar"), []byte("baz")}
	c := selina.SliceAsChannelRaw(data, true)
	got := [][]byte{}
	for v := range c {
		got = append(got, v)
	}
	if !reflect.DeepEqual(got, data) {
		t.Fatalf("SliceAsChannelRaw() got = %v , want = %v", got, data)
	}
}

func TestChannelAsSlice(t *testing.T) {
	want := []string{"foo", "bar", "baz"}
	input := make(chan []byte, 3)
	for _, v := range want {
		input <- []byte(v)
	}
	close(input)
	got := selina.ChannelAsSlice(input)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ChannelAsSlice() got = %v, want = %v", got, want)
	}
}
