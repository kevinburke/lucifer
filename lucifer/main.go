package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const usageStr = `The Lucifer binary makes requests to the Lucifer server.

Usage:

    lucifer command [arguments]

The commands are:

    invalidate      Invalidate the cache for a given file
    run             Run tests for a given file

Use "lucifer help [command]" for more information about a command.

`

const version = "0.1"
const userAgent = "lucifer-client/" + version

func usage() {
	fmt.Fprintf(os.Stderr, usageStr)
	flag.PrintDefaults()
	os.Exit(2)
}

const baseUri = "http://127.0.0.1:11666"
const timeout = time.Duration(5) * time.Second

type Filename string

type LuciferInvalidateRequest struct {
	Files []Filename `json:"files"`
}

type LuciferRunRequest struct {
	Bail  bool       `json:"bail"`
	Files []Filename `json:"files"`
	Grep  string     `json:"grep"`
}

func makeRequest(method string, uri string, body *bytes.Buffer) (*http.Response, error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json, q=0.8; application/problem+json, q=0.6; */*, q=0.3")
	client := http.Client{
		Timeout: timeout,
	}
	return client.Do(req)
}

func makeRunRequest(fnames []Filename, bail bool) (string, error) {
	body := LuciferRunRequest{
		Bail:  bail,
		Files: fnames,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(body)
	resp, err := makeRequest("POST", fmt.Sprintf("%s/v1/test_runs", baseUri), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		// XXX
		return "", errors.New(string(rbody))
	}
	return string(rbody), nil
}

func makeInvalidateRequest(fnames []Filename) (string, error) {
	body := LuciferInvalidateRequest{
		Files: fnames,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(body)
	resp, err := makeRequest("POST", fmt.Sprintf("%s/v1/cache/invalidate", baseUri), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 300 {
		// XXX
		return "", errors.New(string(rbody))
	}
	return string(rbody), nil
}

func handleError(err error, verbose bool) {
	if verbose {
		log.Fatal(err)
	} else {
		os.Exit(0)
	}
}

func init() {
	flag.Usage = usage
}

func doInvalidate(flags *flag.FlagSet, sync bool, verbose bool) {
	err := flags.Parse(os.Args[2:])
	if err != nil {
		handleError(err, verbose)
	}
	args := flags.Args()
	var fnames []Filename
	for i := 0; i < len(args); i++ {
		fnames = append(fnames, Filename(args[i]))
	}
	body, err := makeInvalidateRequest(fnames)
	if err != nil {
		handleError(err, verbose)
	}
	fmt.Println(body)
}

func doRun(flags *flag.FlagSet, bail bool, verbose bool) {
	args := flags.Args()
	var fnames []Filename
	for i := 0; i < len(args); i++ {
		fnames = append(fnames, Filename(args[i]))
	}
	body, err := makeRunRequest(fnames, bail)
	if err != nil {
		handleError(err, verbose)
	}
	fmt.Println(body)
}

func main() {
	invalidateflags := flag.NewFlagSet("invalidate", flag.ExitOnError)
	sync := invalidateflags.Bool("sync", true, "Make request synchronously")
	verbose := invalidateflags.Bool("verbose", false, "Verbose output")
	runflags := flag.NewFlagSet("run", flag.ExitOnError)
	bail := runflags.Bool("bail", false, "Bail after a single test failure")
	runverbose := runflags.Bool("verbose", false, "Verbose response output")
	switch os.Args[1] {
	case "invalidate":
		err := invalidateflags.Parse(os.Args[2:])
		if err != nil {
			handleError(err, true)
		}
		doInvalidate(invalidateflags, *sync, *verbose)
	case "run":
		err := runflags.Parse(os.Args[2:])
		if err != nil {
			handleError(err, true)
		}
		doRun(runflags, *bail, *runverbose)
	default:
		usage()
	}
}
