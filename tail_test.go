// Copyright 2011 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestTail(t *testing.T) {
	if testing.Short() {
		return
	}
	d, err := ioutil.TempDir("", "tail_test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(d)

	logfile := d + "/log"
	f, err := os.Create(logfile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	lines := make(chan string)
	ta := NewTailer(lines)
	if ta == nil {
		t.Fatalf("Couldn't make a tailer.")
	}
	defer ta.Stop()
	ta.Tail(logfile)

	if _, ok := ta.files[logfile]; !ok {
		t.Error("path not found in files map")
	}
}

func TestHandleLogChange(t *testing.T) {
	if testing.Short() {
		return
	}
	d, err := ioutil.TempDir("", "tail_test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(d)

	logfile := d + "/log"
	f, err := os.Create(logfile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	lines := make(chan string)
	ta := NewTailer(lines)
	if ta == nil {
		t.Fatalf("Couldn't make a tailer.")
	}
	defer ta.Stop()
	ta.Tail(logfile)

	_, err = f.WriteString("a\nb\nc\nd\n")
	if err != nil {
		t.Error(err)
	}
	go ta.handleLogUpdate(logfile)

	for _, expected := range []string{"a", "b", "c", "d"} {
		// Run as a goroutine because it's going to emit lines via output channel
		line := <-ta.lines
		if line != expected {
			t.Errorf("line doesn't match:\n\texpected: %s\n\tgot: %s", expected, line)
			continue
		}
	}
}

func TestHandleLogChangePartialLine(t *testing.T) {
	if testing.Short() {
		return
	}
	d, err := ioutil.TempDir("", "tail_test")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(d)

	logfile := d + "/log"
	f, err := os.Create(logfile)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	lines := make(chan string)
	ta := NewTailer(lines)
	if ta == nil {
		t.Fatalf("Couldn't make a tailer.")
	}
	defer ta.Stop()
	ta.Tail(logfile)

	_, err = f.WriteString("a")
	if err != nil {
		t.Error(err)
	}
	go ta.handleLogUpdate(logfile)
	select {
	case line := <-ta.lines:
		t.Errorf("unexpected line found: %s", line)
	default:
	}

	_, err = f.WriteString("b")
	if err != nil {
		t.Error(err)
	}
	go ta.handleLogUpdate(logfile)

	select {
	case line := <-ta.lines:
		t.Errorf("unexpected line found: %s", line)
	default:
	}

	_, err = f.WriteString("\n")
	if err != nil {
		t.Error(err)
	}
	go ta.handleLogUpdate(logfile)
	line := <-ta.lines
	if line != "ab" {
		t.Errorf("line doesn't match: expected 'ab' vs %s", line)
	}
}
