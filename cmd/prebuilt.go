package main

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/ops"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/sql"
	"github.com/licaonfee/selina/workers/text"
)

//GeneralOptions will not use jsonschema automatically because Type is determined in excution time
type GeneralOptions struct {
	Name  string                 `yaml:"name"`
	Type  string                 `yaml:"type"`
	Args  map[string]interface{} `yaml:"args"`
	Fetch []string               `yaml:"fetch"`
}

type NodeFacility interface {
	//Make create a selina.Worker, and wraps it in a selina.Node
	Make(name string) (*selina.Node, error)
}

const (
	splitLine     = "line"
	splitByte     = "byte"
	splitChar     = "char"
	fileOverwrite = "overwrite"
	fileAppend    = "append"
	fileFail      = "fail"
)

//ReadFile read data from a text file
type ReadFile struct {
	Filename  string `mapstructure:"filename" json:"filename" jsonschema:"minLength=1"`
	SplitMode string `mapstructure:"split" default:"line" json:"split,omitempty" jsonschema:"enum=line,enum=byte,enum=char"`
}

func (r *ReadFile) Make(name string) (*selina.Node, error) {
	f, err := os.Open(r.Filename)
	if err != nil {
		return nil, err
	}
	var split bufio.SplitFunc
	switch strings.ToLower(r.SplitMode) {
	case splitLine:
		split = bufio.ScanLines
	case splitByte:
		split = bufio.ScanBytes
	case splitChar:
		split = bufio.ScanRunes
	default:
		return nil, errors.New("invalid splitmode")
	}
	readOpts := text.ReaderOptions{Reader: f, SplitFunc: split, AutoClose: true}
	if err := readOpts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, text.NewReader(readOpts)), nil
}

var _ (NodeFacility) = (*WriteFile)(nil)

type WriteFile struct {
	Filename   string      `mapstructure:"filename" json:"filename" jsonschema:"minLength=1"`
	IfExists   string      `mapstructure:"ifexists" json:"ifexists" default:"fail" json:"ifexists,omitempty" jsonschema:"enum=fail,enum=overwrite,enum=append"`
	Mode       os.FileMode `mapstructure:"mode" default:"420" json:"mode,omitempty"` //0644
	BufferSize int         `mapstructure:"buffer" json:"buffer,omitempty" jsonschema_extras:"minimum=0"`
}

func (w *WriteFile) Make(name string) (*selina.Node, error) {
	flags := os.O_WRONLY | os.O_CREATE
	switch strings.ToLower(w.IfExists) {
	case fileAppend:
		flags |= os.O_APPEND
	case fileOverwrite:
		flags |= os.O_TRUNC
	case fileFail:
		flags |= os.O_EXCL
	default:
		return nil, errors.New("invalid value")
	}
	if w.Mode == 0 {
		w.Mode = 0600
	}
	f, err := os.OpenFile(w.Filename, flags, w.Mode)
	if err != nil {
		return nil, err
	}
	opts := text.WriterOptions{Writer: f, AutoClose: true, BufferSize: w.BufferSize}
	if err := opts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, text.NewWriter(opts)), nil
}

var _ (NodeFacility) = (*SQLQuery)(nil)

type SQLQuery struct {
	Driver string `mapstructure:"driver" json:"driver" jsonschema:"enum=mysql,enum=postgres,enum=clickhouse"`
	DSN    string `mapstructure:"dsn" json:"dsn" jsonschema:"minLength=1"`
	Query  string `mapstrcuture:"query" json:"query" jsonschema:"minLength=1"`
}

func (s *SQLQuery) Make(name string) (*selina.Node, error) {
	opts := sql.ReaderOptions{Driver: s.Driver,
		ConnStr: s.DSN,
		Query:   s.Query}
	if err := opts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, sql.NewReader(opts)), nil
}

var _ NodeFacility = (*Regexp)(nil)

type Regexp struct {
	Pattern string `mapstructure:"pattern" json:"pattern" jsonschema:"minLegth=1"`
}

func (r *Regexp) Make(name string) (*selina.Node, error) {
	opts := regex.FilterOptions{Pattern: r.Pattern}
	if err := opts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, regex.NewFilter(opts)), nil
}

var _ (NodeFacility) = (*CSV)(nil)

type CSV struct {
	Mode    string   `mapstructure:"mode" json:"mode" jsonschema:"enum=decode,enum=encode"`
	Header  []string `mapstructure:"header" json:"header,omitempty"`
	Comma   rune     `mapstructure:"comma" json:"comma,omitempty" jsonschema:"minLegth=1,maxLength=1"`
	UseCrlf bool     `mapstructure:"crlf" json:"crlf,omitempty" jsonschema:"minLegth=1,maxLength=1"`
	Comment rune     `mapstructure:"comment" json:"comment,omitempty" jsonschema:"minLegth=1,maxLength=1"`
}

func (c *CSV) Make(name string) (*selina.Node, error) {
	var w selina.Worker
	switch c.Mode {
	case "decode":
		opts := csv.DecoderOptions{Header: c.Header, Comma: c.Comma, Comment: c.Comment}
		w = csv.NewDecoder(opts)
	case "encode":
		opts := csv.EncoderOptions{Header: c.Header, Comma: c.Comma, UseCRLF: c.UseCrlf}
		w = csv.NewEncoder(opts)
	default:
		return nil, errors.New("CSV mode must be decode|encode")
	}
	return selina.NewNode(name, w), nil
}

var _ NodeFacility = (*Cron)(nil)

type Cron struct {
	Spec    string `mapstructure:"spec" json:"spec" jsonschema:"example=* * * * * *,example=@every 1h"`
	Message string `mapstructures:"message" json:"message,omitempty"`
}

func (c *Cron) Make(name string) (*selina.Node, error) {
	opts := ops.CronOptions{Spec: c.Spec, Message: []byte(c.Message)}
	if err := opts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, ops.NewCron(opts)), nil
}
