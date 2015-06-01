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

type Filename string

type LuciferInvalidateRequest struct {
	Files []Filename `json:"files"`
}

func makeInvalidateRequest(fnames []Filename) error {
	body := LuciferInvalidateRequest{
		Files: fnames,
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(body)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/cache/invalidate", baseUri), &buf)
	if err != nil {
		return err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json, q=0.8; application/problem+json, q=0.6; */*, q=0.3")
	client := http.Client{
		Timeout: time.Duration(5) * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		// XXX
		return errors.New(string(rbody))
	}
	fmt.Println(string(rbody))
	return nil
}

func main() {
	flag.Usage = usage
	invalidateflags := flag.NewFlagSet("invalidate", flag.ExitOnError)
	sync := invalidateflags.Bool("sync", true, "Make request synchronously")
	verbose := invalidateflags.Bool("verbose", false, "Verbose output")
	runflags := flag.NewFlagSet("run", flag.ExitOnError)
	switch os.Args[1] {
	case "invalidate":
		err := invalidateflags.Parse(os.Args[2:])
		if err != nil {
			if *verbose {
				log.Fatal(err)
			} else {
				os.Exit(0)
			}
		}
		fmt.Println(*sync)
		args := invalidateflags.Args()
		var fnames []Filename
		for i := 0; i < len(args); i++ {
			fnames = append(fnames, Filename(args[i]))
		}
		err = makeInvalidateRequest(fnames)
		if err != nil {
			if *verbose {
				log.Fatal(err)
			} else {
				os.Exit(0)
			}
		}
	case "run":
		err := runflags.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
	default:
		usage()
	}
}
