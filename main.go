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
	Path     string            `json:path`
	Type     string            `json:type`
	ArgsGet  map[string]string `json:argsGet`
	ArgsPost map[string]string `json:argsPost`
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

		postParam := url.Values{}
		for key, param := range testURL.ArgsPost {
			postParam.Add(key, param)
		}
		getParam := url.Values{}
		for key, param := range testURL.ArgsGet {
			getParam.Add(key, param)
		}

		completeURL, err := PrepareURL(p.Proto+p.Base+testURL.Path, getParam)
		if err != nil {
			log.Fatal(err)
		}
		rawHTTP, err := RawHTTP(testURL.Type, completeURL, postParam)
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
			log.Fatal("error creating response: " + err.Error())
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("error reading response: " + err.Error())
		}
		resp.Body.Close()

		if len(body) <= 0 && err == nil {
			log.Fatal("empty response")
		}
		fmt.Println("recive data:", time.Since(t0))
	}

}

// PrepareURL takes a base string (likes "http://site.com/") and a map of param-value and builds the url.
// This functions are been made to use the "net/url" package
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
func RawHTTP(connType, inURL string, data url.Values) (string, error) {

	_, err := url.Parse(inURL)
	if err != nil {
		return "", errors.New("input url ->" + err.Error())
	}

	// prepare the header
	req, err := http.NewRequest(connType, inURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", errors.New("preparing header -> " + err.Error())
	}
	var header bytes.Buffer
	err = req.Write(&header)
	if err != nil {
		return "", errors.New("reading header -> " + err.Error())
	}

	return header.String(), nil
}
