package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/ops"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/sql"
	"github.com/licaonfee/selina/workers/text"
)

var _ error = (*MakeError)(nil)

type MakeError struct {
	Facility string
	err      error
}

func (m *MakeError) Unwrap() error {
	return m.err
}

func (m *MakeError) Error() string {
	return fmt.Sprintf("failed to make [%s]  : %s", m.Facility, m.err.Error())
}

func newMakeError(f interface{}, err error) *MakeError {
	return &MakeError{Facility: reflect.TypeOf(f).String(), err: err}
}

//GeneralOptions will not use jsonschema automatically because Type is determined in excution time
type GeneralOptions struct {
	Name  string                 `yaml:"name"`
	Type  string                 `yaml:"type"`
	Args  map[string]interface{} `yaml:"args"`
	Fetch []string               `yaml:"fetch"`
}

type NewFacility func() NodeFacility

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

func NewReadFile() NodeFacility {
	return &ReadFile{SplitMode: splitLine}
}

//ReadFile read data from a text file
type ReadFile struct {
	Filename  string `mapstructure:"filename" json:"filename" jsonschema:"minLength=1"`
	SplitMode string `mapstructure:"split" json:"split,omitempty" jsonschema:"enum=line,enum=byte,enum=char"`
}

func (r *ReadFile) Make(name string) (*selina.Node, error) {
	f, err := os.Open(r.Filename)
	if err != nil {
		return nil, newMakeError(r, err)
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
		return nil, newMakeError(r, errors.New("invalid split mode "+r.SplitMode))
	}
	readOpts := text.ReaderOptions{Reader: f, SplitFunc: split, AutoClose: true}
	if err := readOpts.Check(); err != nil {
		return nil, newMakeError(r, err)
	}
	return selina.NewNode(name, text.NewReader(readOpts)), nil
}

var _ (NodeFacility) = (*WriteFile)(nil)

func NewWriteFile() NodeFacility {
	return &WriteFile{IfExists: fileFail,
		Mode: 0644}
}

type WriteFile struct {
	Filename   string      `mapstructure:"filename" json:"filename" jsonschema:"minLength=1"`
	IfExists   string      `mapstructure:"ifexists" json:"ifexists,omitempty" jsonschema:"enum=fail,enum=overwrite,enum=append"`
	Mode       os.FileMode `mapstructure:"mode" json:"mode,omitempty"` //0644
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
		return nil, newMakeError(w, errors.New("invalid ifexist value "+w.IfExists))
	}
	if w.Mode == 0 {
		w.Mode = 0600
	}
	f, err := os.OpenFile(w.Filename, flags, w.Mode)
	if err != nil {
		return nil, newMakeError(w, err)
	}
	opts := text.WriterOptions{Writer: f, AutoClose: true, BufferSize: w.BufferSize}
	if err := opts.Check(); err != nil {
		return nil, newMakeError(w, err)
	}
	return selina.NewNode(name, text.NewWriter(opts)), nil
}

var _ (NodeFacility) = (*SQLQuery)(nil)

func NewSQLQuery() NodeFacility {
	return &SQLQuery{}
}

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
		return nil, newMakeError(s, err)
	}
	return selina.NewNode(name, sql.NewReader(opts)), nil
}

var _ NodeFacility = (*SQLInsert)(nil)

func NewSQLInsert() NodeFacility {
	return &SQLInsert{}
}

type SQLInsert struct {
	Driver string `mapstructure:"driver" json:"driver" jsonschema:"enum=mysql,enum=postgres,enum=clickhouse"`
	DSN    string `mapstructure:"dsn" json:"dsn" jsonschema:"minLength=1"`
	Table  string `mapstructure:"table" json:"table" jsonschema:"minLength=1"`
}

func (s *SQLInsert) Make(name string) (*selina.Node, error) {
	opts := sql.WriterOptions{
		Driver:  s.Driver,
		ConnStr: s.DSN,
		Table:   s.Table,
	}
	if err := opts.Check(); err != nil {
		return nil, newMakeError(s, err)
	}
	return selina.NewNode(name, sql.NewWriter(opts)), nil
}

var _ NodeFacility = (*Regexp)(nil)

func NewRegexp() NodeFacility {
	return &Regexp{}
}

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

func NewCSV() NodeFacility {
	return &CSV{}
}

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
		return nil, newMakeError(c, errors.New("invalid mode value "+c.Mode))
	}
	return selina.NewNode(name, w), nil
}

var _ NodeFacility = (*Cron)(nil)

func NewCron() NodeFacility {
	return &Cron{}
}

type Cron struct {
	Spec    string `mapstructure:"spec" json:"spec" jsonschema:"example=* * * * * *,example=@every 1h"`
	Message string `mapstructures:"message" json:"message,omitempty"`
}

func (c *Cron) Make(name string) (*selina.Node, error) {
	opts := ops.CronOptions{Spec: c.Spec, Message: []byte(c.Message)}
	if err := opts.Check(); err != nil {
		return nil, newMakeError(c, err)
	}
	return selina.NewNode(name, ops.NewCron(opts)), nil
}
