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

type Conf struct {
	Proto string `json:proto`
	Base  string `json:base`
	Port  string `json:port`
	Links []Link `json:link`
}

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

		// prepare the url
		completeURL, err := PrepareURL(p.Proto+p.Base+testURL.Path, testURL.ArgsGet)
		if err != nil {
			log.Fatal(err)
		}
		// get the raw http
		rawHTTP, err := RawHTTP(testURL.Type, completeURL, testURL.ArgsPost)
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
		r := bufio.NewReader(conn)
		for {
			b, _, _ := r.ReadLine()
			if strings.TrimSpace(string(b)) == "" {
				break
			}
		}
		fmt.Println("recieve data:", time.Since(t0))

		// close connection
		conn.Close()
	}

}

// PrepareURL takes a base string (likes "http://site.com/") and a
// map of param-value and builds the url.
// This function have been made to use the "net/url" package
func PrepareURL(base string, params url.Values) (string, error) {

	encode := params.Encode()
	if encode != "" {
		encode = "?" + encode
	}

	url, err := url.Parse(base + encode)

	if err != nil {
		return "", errors.New("unable to create url: " + err.Error())
	}

	return url.String(), nil
}

// RawHTTP brings an url and returns a raw http request based from https://tools.ietf.org/html/rfc2068
// It accepts the conn type, and the data to send in body.
func RawHTTP(connType, inURL string, data url.Values) (string, error) {

	_, err := url.Parse(inURL)
	if err != nil {
		return "", errors.New("input url: " + err.Error())
	}

	// prepare the header
	req, err := http.NewRequest(connType, inURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", errors.New("preparing header: " + err.Error())
	}
	var header bytes.Buffer
	err = req.Write(&header)
	if err != nil {
		return "", errors.New("reading header: " + err.Error())
	}

	return header.String(), nil
}
