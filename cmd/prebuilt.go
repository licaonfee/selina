package main

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/licaonfee/selina"
	"github.com/licaonfee/selina/workers/csv"
	"github.com/licaonfee/selina/workers/regex"
	"github.com/licaonfee/selina/workers/sql"
	"github.com/licaonfee/selina/workers/text"
)

type GeneralOptions struct {
	Name string                 `yaml:"name"`
	Type string                 `yaml:"type"`
	Args map[string]interface{} `yaml:"args"`
}

type NodeFacility interface {
	//Make create a selina.Worker, and wraps it in a selina.Node
	//is mandatory that empty values give a default behavior
	Make(name string) (*selina.Node, error)
}

var _ NodeFacility = (*ReadFile)(nil)

//ReadFile read data from a text file
type ReadFile struct {
	Filename  string `yaml:"filename",mapstructure:"filename"`
	SplitMode string `yaml:"split",mapstructure:"split"`
}

func (r *ReadFile) Make(name string) (*selina.Node, error) {
	f, err := os.Open(r.Filename)
	if err != nil {
		return nil, err
	}
	var split bufio.SplitFunc
	switch strings.ToLower(r.SplitMode) {
	case "line", "":
		split = bufio.ScanLines
	case "byte":
		split = bufio.ScanBytes
	case "char":
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
	Filename string      `yaml:"filename",mapstructure:"filename"`
	IfExists string      `yaml:"ifexists","mapstructure:"ifexists"`
	Mode     os.FileMode `yaml:"mode",mapstructure:"mode"`
}

func (w *WriteFile) Make(name string) (*selina.Node, error) {
	flags := os.O_WRONLY | os.O_CREATE
	switch w.IfExists {
	case "append":
		flags |= os.O_APPEND
	case "overwrite":
		flags |= os.O_TRUNC
	case "fail", "":
		flags |= os.O_EXCL
	default:
		return nil, errors.New("invalid value")
	}
	f, err := os.OpenFile(w.Filename, flags, w.Mode)
	if err != nil {
		return nil, err
	}
	opts := text.WriterOptions{Writer: f, AutoClose: true}
	if err := opts.Check(); err != nil {
		return nil, err
	}
	return selina.NewNode(name, text.NewWriter(opts)), nil
}

var _ (NodeFacility) = (*SQLQuery)(nil)

type SQLQuery struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
	Query  string `yaml:"query"`
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
	Pattern string `yaml:"pattern",mapstructure:"pattern"`
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
	Mode    string   `yaml:"mode",mapstructure:"mode"`
	Header  []string `yaml:"header", mapstructure:header"`
	Comma   rune     `yaml:"comma",mapstructure:"comma"`
	UseCrlf bool     `yaml:"crlf",mapstructure:"crlf"`
	Comment rune     `yaml:"comment",mapstructure:"comment"`
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
