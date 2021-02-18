package main

import (
	"bufio"
	"io"
	"strings"
)

// StringIterator used to scan domains
type StringIterator interface {
	getNext() string
	hasNext() bool
}

// DomainFromArray - simple iterator
type DomainFromArray struct {
	index int
	array []string
}

func (d *DomainFromArray) getNext() string {
	d.index = d.index + 1;
	return strings.ToLower(d.array[d.index - 1]);
}

func (d *DomainFromArray) hasNext() bool {
	return d.index < len(d.array);
}

// CreateArrayIterator returns iterator of strings
func CreateArrayIterator(array []string) StringIterator {
	xd := new(DomainFromArray)
	xd.index = 0;
	xd.array = array;
	
	return xd;
}

// ReaderFromFile - simple iterator
type ReaderFromFile struct {
	reader *bufio.Reader
	str string
}


func (reader *ReaderFromFile) hasNext() bool {
	text, err := reader.reader.ReadString('\n')
	if (err != nil) {
		return false
	}
	reader.str = strings.Trim(strings.TrimRight(text, "\r\n"), " ");
	return len(reader.str) > 0
}

func (reader *ReaderFromFile) getNext() string {
	return strings.ToLower(reader.str);
}

// CreateReaderIterator - scan domains from io.Reader
func CreateReaderIterator(reader io.Reader) StringIterator {
	xd := new(ReaderFromFile)
	xd.reader = bufio.NewReader(reader)
	return xd
}