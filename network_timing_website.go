// network_timing_website.go: small utility to test the
// timing of a website. Reads the input json file, like example.json
// and makes the requests.
// For each request you can specify the type of the request (GET or POST)
// and the get or posts params.
// For each "link", the script measures the time to resolve the address ("resolv")
// time to establish the tcp connection ("conn"), time to send the data through
// the connection ("send data") and finally the time to receive the data ("receive data").
// Requests are made sequentially
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Conf defines the struct for decode the json
type Conf struct {
	Proto string `json:proto`
	Base  string `json:base`
	Port  string `json:port`
	Links []Link `json:link`
}

// Link defines a single request to the website
type Link struct {
	Path     string     `json:path`
	Type     string     `json:type`
	ArgsGet  url.Values `json:argsGet`
	ArgsPost url.Values `json:argsPost`
}

func main() {

	var err error
	var fileInput string
	flag.StringVar(&fileInput, "f", "", "input json file")
	flag.Parse()

	if fileInput == "" {
		log.Fatalf("no iput file")
	}

	file, err := os.Open(fileInput)
	if err != nil {
		log.Fatal("error open file:", err)
	}
	defer file.Close()

	contentFile, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal("read file error:", err)
	}

	p := &Conf{}
	err = json.Unmarshal(contentFile, p)
	if err != nil {
		log.Fatal("parse file error:", err)
	}

	var t0 time.Time

	for _, testURL := range p.Links {

		// get the raw http
		rawHTTP, err := RawHTTP(testURL.Type, p.Proto+p.Base+testURL.Path, testURL.ArgsGet, testURL.ArgsPost)
		if err != nil {
			log.Fatal(err)
		}

		// resolve the ip
		t0 = time.Now()
		ip, err := net.ResolveIPAddr("ip", p.Base)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("resolv:", time.Since(t0))

		// get the conn
		t0 = time.Now()
		conn, err := net.Dial("tcp", ip.String()+":"+p.Port)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("conn:", time.Since(t0))

		// send data to conn
		t0 = time.Now()
		fmt.Fprintln(conn, rawHTTP)
		fmt.Println("send data:", time.Since(t0))

		// read data
		t0 = time.Now()

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			log.Fatalln("error creating response: " + err.Error())
		}

		_, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
		}

		fmt.Println("receive data:", time.Since(t0))
		fmt.Println()

		// close connection
		conn.Close()

	}

}

// RawHTTP brings an url and returns a raw http request based from https://tools.ietf.org/html/rfc2068
// It accepts the conn type, and the data to send in body.
func RawHTTP(connType, base string, dataGet, dataPost url.Values) (string, error) {

	encode := dataGet.Encode()
	if encode != "" {
		encode = "?" + encode
	}

	url, err := url.Parse(base + encode)

	if err != nil {
		return "", errors.New("unable to create url: " + err.Error())
	}

	// prepare the header
	body := bytes.NewBufferString(dataPost.Encode())
	req, err := http.NewRequest(connType, url.String(), body)
	if err != nil {
		return "", errors.New("preparing header: " + err.Error())
	}
	var header bytes.Buffer
	if connType == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	err = req.Write(&header)
	if err != nil {
		return "", errors.New("reading header: " + err.Error())
	}

	return header.String(), nil
}
