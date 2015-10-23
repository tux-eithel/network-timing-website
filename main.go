package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

func main() {

	var err error
	var t0 time.Time

	postParam := url.Values{}
	postParam.Add("action", "dummy_ajax_call")

	// prepare the raw http
	rawHTTP, err := RawHTTP("http://www.silvermuse.info/wp-admin/admin-ajax.php", postParam)
	if err != nil {
		log.Fatal(err)
	}

	// resolve the ip
	t0 = time.Now()
	ip, err := net.ResolveIPAddr("ip", "www.silversea.info")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("resolv:", time.Since(t0))

	// get the conn
	t0 = time.Now()
	conn, err := net.Dial("tcp", ip.String()+":80")
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

// RawHTTP brings an url and returns a raw http request based from https://tools.ietf.org/html/rfc2068
func RawHTTP(inURL string, data url.Values) (string, error) {

	_, err := url.Parse(inURL)
	if err != nil {
		return "", errors.New("input url ->" + err.Error())
	}

	// prepare the header
	req, err := http.NewRequest("POST", inURL, bytes.NewBufferString(data.Encode()))
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
