package csv_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/csv"
)

func TestEncoderProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    csv.EncoderOptions
		input   []string
		want    []string
		wantErr error
	}{
		{
			name:    "No upstream channel",
			opts:    csv.EncoderOptions{},
			input:   []string{},
			want:    []string{},
			wantErr: selina.ErrNilUpstream,
		},
		{
			name:    "Success",
			opts:    csv.EncoderOptions{Header: []string{"name", "id"}},
			input:   []string{`{"name": "Selina","id":0}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{`name,id` + "\n", `Selina,0` + "\n", `Lizbeth,1` + "\n"},
			wantErr: nil,
		},
		{
			name:    "Use pipe as comma",
			opts:    csv.EncoderOptions{Header: []string{"name", "id"}, Comma: '|'},
			input:   []string{`{"name": "Selina","id":0}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{`name|id` + "\n", `Selina|0` + "\n", `Lizbeth|1` + "\n"},
			wantErr: nil,
		},
		{
			name:    "filter header",
			opts:    csv.EncoderOptions{Header: []string{"name"}},
			input:   []string{`{"name": "Selina","id":0}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{`name` + "\n", `Selina` + "\n", `Lizbeth` + "\n"},
			wantErr: nil,
		},
		{
			name:    "Auto header",
			opts:    csv.EncoderOptions{},
			input:   []string{`{"name": "Selina","id":0}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{`id,name` + "\n", `0,Selina` + "\n", `1,Lizbeth` + "\n"},
			wantErr: nil,
		},
		{
			name:    "Missing data",
			opts:    csv.EncoderOptions{Header: []string{"name", "id", "color"}},
			input:   []string{`{"name": "Selina","id":0, "color":"yellow"}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{`name,id,color` + "\n", `Selina,0,yellow` + "\n", `Lizbeth,1,` + "\n"},
			wantErr: nil,
		},
		{
			name:    "Invalid JSON",
			opts:    csv.EncoderOptions{Header: []string{"name", "id", "color"}},
			input:   []string{`{"name": "Selina","id"0, "color":"yellow"}`, `{"name":"Lizbeth","id":1}`},
			want:    []string{},
			wantErr: &json.SyntaxError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := csv.NewEncoder(tt.opts)
			var input chan []byte
			if len(tt.input) > 0 {
				input = selina.SliceAsChannel(tt.input, true)
			}
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			if err := c.Process(context.Background(), args); err != tt.wantErr && errors.Is(err, tt.wantErr) {
				t.Fatalf("Process() err =%v", err)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestEncoderProcessCancelation(t *testing.T) {
	c := csv.NewEncoder(csv.EncoderOptions{})
	if err := workers.ATProcessCancel(c); err != nil {
		t.Fatal(err)
	}
}

func TestEncoderProcessCloseInput(t *testing.T) {
	c := csv.NewEncoder(csv.EncoderOptions{})
	if err := workers.ATProcessCloseInput(c); err != nil {
		t.Fatal(err)
	}
}
func TestEncoderProcessCloseOutput(t *testing.T) {
	c := csv.NewEncoder(csv.EncoderOptions{})
	if err := workers.ATProcessCloseOutput(c); err != nil {
		t.Fatal(err)
	}
}

func TestDecoderProcess(t *testing.T) {
	tests := []struct {
		name    string
		opts    csv.DecoderOptions
		input   []string
		want    []string
		wantErr error
	}{
		{
			name:    "no upstream",
			opts:    csv.DecoderOptions{},
			input:   []string{},
			want:    []string{},
			wantErr: selina.ErrNilUpstream,
		},
		{
			name:    "Success",
			opts:    csv.DecoderOptions{Header: []string{"id", "name"}},
			input:   []string{`6,Selina`, `7,Lizbeth`},
			want:    []string{`{"id":"6","name":"Selina"}`, `{"id":"7","name":"Lizbeth"}`},
			wantErr: nil,
		},
		{
			name:    "pipe is comma",
			opts:    csv.DecoderOptions{Header: []string{"id", "name"}, Comma: '|'},
			input:   []string{`6|Selina`, `7|Lizbeth`},
			want:    []string{`{"id":"6","name":"Selina"}`, `{"id":"7","name":"Lizbeth"}`},
			wantErr: nil,
		},
		{
			name:    "skip comments",
			opts:    csv.DecoderOptions{Header: []string{"id", "name"}, Comment: '#'},
			input:   []string{`6,Selina`, `#8,Maria`, `7,Lizbeth`},
			want:    []string{`{"id":"6","name":"Selina"}`, `{"id":"7","name":"Lizbeth"}`},
			wantErr: nil,
		},
		{
			name:  "Empty fields",
			opts:  csv.DecoderOptions{Header: []string{"id", "name", "color", "pet"}},
			input: []string{`6,Selina,pink,`, `7,Lizbeth,,cat`},
			want: []string{`{"color":"pink","id":"6","name":"Selina","pet":""}`,
				`{"color":"","id":"7","name":"Lizbeth","pet":"cat"}`},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := csv.NewDecoder(tt.opts)
			var input chan []byte
			if len(tt.input) > 0 {
				input = selina.SliceAsChannel(tt.input, true)
			}
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}
			if err := d.Process(context.Background(), args); err != tt.wantErr {
				t.Fatalf("Process() err = %v", err)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Process() got = %v , want = %v", got, tt.want)
			}

		})
	}
}

func TestDecoderProcessCancelation(t *testing.T) {
	c := csv.NewDecoder(csv.DecoderOptions{})
	if err := workers.ATProcessCancel(c); err != nil {
		t.Fatal(err)
	}
}

func TestDecoderProcessCloseInput(t *testing.T) {
	c := csv.NewDecoder(csv.DecoderOptions{})
	if err := workers.ATProcessCloseInput(c); err != nil {
		t.Fatal(err)
	}
}
func TestDecoderProcessCloseOutput(t *testing.T) {
	c := csv.NewDecoder(csv.DecoderOptions{})
	if err := workers.ATProcessCloseOutput(c); err != nil {
		t.Fatal(err)
	}
}
