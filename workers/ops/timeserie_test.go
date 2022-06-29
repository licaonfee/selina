package ops_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers"
	"github.com/licaonfee/selina/workers/ops"
)

func TestTimeSerieProcessCancelation(t *testing.T) {
	c := ops.NewTimeSerie(ops.TimeSerieOptions{Start: time.Now().Add(-time.Hour),
		Stop:      time.Now(),
		Step:      time.Second,
		Generator: func(t time.Time) float64 { return 0.0 }})
	if err := workers.ATProcessCancel(c); err != nil {
		t.Fatal(err)
	}
}

func TestTimeSerieProcessCloseInput(t *testing.T) {
	c := ops.NewTimeSerie(ops.TimeSerieOptions{Start: time.Now().Add(-time.Hour),
		Stop:      time.Now(),
		Step:      time.Second,
		Generator: func(t time.Time) float64 { return 0.0 }})
	if err := workers.ATProcessCloseInput(c); err != nil {
		t.Fatal(err)
	}
}
func TestTimeSerieProcessCloseOutput(t *testing.T) {
	c := ops.NewTimeSerie(ops.TimeSerieOptions{Start: time.Now().Add(-time.Hour),
		Stop:      time.Now(),
		Step:      time.Minute,
		Generator: func(t time.Time) float64 { return 0.0 }})
	if err := workers.ATProcessCloseOutput(c); err != nil {
		t.Fatal(err)
	}
}

func TestTimeSerieProcess(t *testing.T) {
	var marshalError = errors.New("marshal error")
	test := []struct {
		name    string
		opts    ops.TimeSerieOptions
		want    []string
		wantErr error
	}{
		{
			name: "Two minutes",
			opts: ops.TimeSerieOptions{
				Start:       time.Date(2022, 1, 1, 1, 0, 0, 0, time.UTC),
				Stop:        time.Date(2022, 1, 1, 1, 2, 0, 0, time.UTC),
				Step:        time.Minute,
				Generator:   func(t time.Time) float64 { return 1.0 },
				WriteFormat: json.Marshal,
			},
			want: []string{`{"Time":"2022-01-01T01:01:00Z","Value":1}`,
				`{"Time":"2022-01-01T01:02:00Z","Value":1}`},
			wantErr: nil,
		},
		{
			name: "Error marshal",
			opts: ops.TimeSerieOptions{
				Start:       time.Date(2022, 1, 1, 1, 0, 0, 0, time.UTC),
				Stop:        time.Date(2022, 1, 1, 1, 2, 0, 0, time.UTC),
				Step:        time.Minute,
				Generator:   func(t time.Time) float64 { return 1.0 },
				WriteFormat: func(i interface{}) ([]byte, error) { return nil, marshalError },
			},
			want:    []string{},
			wantErr: marshalError,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			ts := ops.NewTimeSerie(tt.opts)
			input := make(chan []byte)
			output := make(chan []byte, len(tt.want))
			args := selina.ProcessArgs{Input: input, Output: output}

			gotErr := ts.Process(context.Background(), args)
			if !errors.Is(gotErr, tt.wantErr) {
				t.Errorf("Process() err = %v , wantErr = %v ", gotErr, tt.wantErr)
			}
			got := selina.ChannelAsSlice(output)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() got %v , want %v", got, tt.want)
			}
		})
	}
}
