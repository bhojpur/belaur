package flag_test

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"os"
	"syscall"
	"testing"
	"time"

	. "github.com/bhojpur/belaur/pkg/utils/flag"
)

// Test parsing a environment variables
func TestParseEnv(t *testing.T) {

	syscall.Setenv("BOOL", "")
	syscall.Setenv("BOOL2", "true")
	syscall.Setenv("INT", "22")
	syscall.Setenv("INT64", "0x23")
	syscall.Setenv("UINT", "24")
	syscall.Setenv("UINT64", "25")
	syscall.Setenv("STRING", "hello")
	syscall.Setenv("FLOAT64", "2718e28")
	syscall.Setenv("DURATION", "2m")

	f := NewFlagSet(os.Args[0], ContinueOnError)

	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")

	err := f.ParseEnv(os.Environ())
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
}

// Test parsing a configuration file
func TestParseFile(t *testing.T) {

	f := NewFlagSet(os.Args[0], ContinueOnError)

	boolFlag := f.Bool("bool", false, "bool value")
	bool2Flag := f.Bool("bool2", false, "bool2 value")
	intFlag := f.Int("int", 0, "int value")
	int64Flag := f.Int64("int64", 0, "int64 value")
	uintFlag := f.Uint("uint", 0, "uint value")
	uint64Flag := f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	float64Flag := f.Float64("float64", 0, "float64 value")
	durationFlag := f.Duration("duration", 5*time.Second, "time.Duration value")

	err := f.ParseFile("./testdata/test.conf")
	if err != nil {
		t.Fatal("expected no error; got ", err)
	}
	if *boolFlag != true {
		t.Error("bool flag should be true, is ", *boolFlag)
	}
	if *bool2Flag != true {
		t.Error("bool2 flag should be true, is ", *bool2Flag)
	}
	if *intFlag != 22 {
		t.Error("int flag should be 22, is ", *intFlag)
	}
	if *int64Flag != 0x23 {
		t.Error("int64 flag should be 0x23, is ", *int64Flag)
	}
	if *uintFlag != 24 {
		t.Error("uint flag should be 24, is ", *uintFlag)
	}
	if *uint64Flag != 25 {
		t.Error("uint64 flag should be 25, is ", *uint64Flag)
	}
	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is ", *stringFlag)
	}
	if *float64Flag != 2718e28 {
		t.Error("float64 flag should be 2718e28, is ", *float64Flag)
	}
	if *durationFlag != 2*time.Minute {
		t.Error("duration flag should be 2m, is ", *durationFlag)
	}
}

func TestParseFileUnknownFlag(t *testing.T) {
	f := NewFlagSet("test", ContinueOnError)
	if err := f.ParseFile("./testdata/bad_test.conf"); err == nil {
		t.Error("parse did not fail for unknown flag; ", err)
	}
}

func TestDefaultConfigFlagname(t *testing.T) {
	f := NewFlagSet("test", ContinueOnError)

	f.Bool("bool", false, "bool value")
	f.Bool("bool2", false, "bool2 value")
	f.Int("int", 0, "int value")
	f.Int64("int64", 0, "int64 value")
	f.Uint("uint", 0, "uint value")
	f.Uint64("uint64", 0, "uint64 value")
	stringFlag := f.String("string", "0", "string value")
	f.Float64("float64", 0, "float64 value")
	f.Duration("duration", 5*time.Second, "time.Duration value")

	f.String(DefaultConfigFlagname, "./testdata/test.conf", "config path")

	if err := os.Unsetenv("STRING"); err != nil {
		t.Error(err)
	}

	if err := f.Parse([]string{}); err != nil {
		t.Error("parse failed; ", err)
	}

	if *stringFlag != "hello" {
		t.Error("string flag should be `hello`, is", *stringFlag)
	}
}

func TestDefaultConfigFlagnameMissingFile(t *testing.T) {
	f := NewFlagSet("test", ContinueOnError)
	f.String(DefaultConfigFlagname, "./testdata/missing", "config path")

	if err := os.Unsetenv("STRING"); err != nil {
		t.Error(err)
	}
	if err := f.Parse([]string{}); err == nil {
		t.Error("expected error of missing config file, got nil")
	}
}

func TestFlagSetParseErrors(t *testing.T) {
	fs := NewFlagSet("test", ContinueOnError)
	fs.Int("int", 0, "int value")

	args := []string{"-int", "bad"}
	expected := `invalid value "bad" for flag -int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse(args); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from args, got: %v", expected, err)
	}

	if err := os.Setenv("INT", "bad"); err != nil {
		t.Fatalf("error setting env: %s", err.Error())
	}
	expected = `invalid value "bad" for environment variable int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse([]string{}); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from env, got: %v", expected, err)
	}
	if err := os.Unsetenv("INT"); err != nil {
		t.Fatalf("error unsetting env: %s", err.Error())
	}

	fs.String("config", "", "config filename")
	args = []string{"-config", "testdata/bad_test.conf"}
	expected = `invalid value "bad" for configuration variable int: strconv.ParseInt: parsing "bad": invalid syntax`
	if err := fs.Parse(args); err == nil || err.Error() != expected {
		t.Errorf("expected error %q parsing from config, got: %v", expected, err)
	}
}

func TestTestingPackageFlags(t *testing.T) {
	f := NewFlagSet("test", ContinueOnError)
	if err := f.Parse([]string{"-test.v", "-test.count", "1"}); err != nil {
		t.Error(err)
	}
}
