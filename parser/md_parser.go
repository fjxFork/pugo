package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const (
	MD_PARSER_PREFIX = "```"
)

type (
	MdParser struct {
		blocks []Block
	}
)

func NewMdParser() *MdParser {
	return &MdParser{
		blocks: []Block{new(IniBlock), new(MarkdownBlock)},
	}
}

func (mp *MdParser) Is(data []byte) bool {
	data = bytes.TrimLeft(data, "\n")
	return bytes.HasPrefix(data, []byte(MD_PARSER_PREFIX))
}

func (mp *MdParser) Detect(mark []byte) Block {
	for _, b := range mp.blocks {
		if b.Is(mark) {
			return b.New()
		}
	}
	return nil
}

func (mp *MdParser) Parse(src []byte) ([]Block, error) {
	buf := bytes.NewBuffer(src)
	return mp.ParseReader(buf)
}

func (mp *MdParser) ParseReader(r io.Reader) ([]Block, error) {
	var (
		currentBlock Block   = nil
		blocks       []Block = nil
		reader               = bufio.NewReader(r)
	)
	for {
		lineData, _, err := reader.ReadLine()
		if currentBlock == nil {
			if len(lineData) == 0 {
				continue
			}
			if len(blocks) > 0 {
				// the second block must be markdown block
				currentBlock = new(MarkdownBlock).New()
			} else {
				if currentBlock = mp.Detect(bytes.TrimLeft(lineData, MD_PARSER_PREFIX)); currentBlock == nil {
					return nil, errors.New("block-parse-first-error")
				}
			}
			continue
		}

		// when parsing first block, check end mark to close the block.
		if bytes.Equal(lineData, []byte(MD_PARSER_PREFIX)) && len(blocks) == 0 {
			blocks = append(blocks, currentBlock)
			currentBlock = nil
			continue
		}

		// write block
		if err := currentBlock.Write(append(lineData, []byte("\n")...)); err != nil {
			return nil, err
		}

		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
	}

	if currentBlock != nil {
		blocks = append(blocks, currentBlock)
	}
	return blocks, nil
}
