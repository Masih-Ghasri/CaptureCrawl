package model

import "time"

type Config struct {
	URL          string
	OutFile      string
	MaxPages     int
	MaxDepth     int
	Workers      int
	Delay        time.Duration
	PageTimeout  time.Duration
	IgnoreRobots bool
	DumpJSON     bool
}

type BlockKind int

const (
	BlockHeading BlockKind = iota
	BlockParagraph
	BlockList
	BlockTable
	BlockCode
)

type Block struct {
	Kind  BlockKind
	Level int
	Text  string
	Rows  [][]string
}

type Page struct {
	URL     string
	Title   string
	Depth   int
	Parent  string
	Content []Block
	Links   []string
	Error   string
}
