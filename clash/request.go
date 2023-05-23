package clash

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/juju/errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Secrete = ""

func getSecrete() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", errors.Trace(err)
	}
	p := filepath.Join(dir, "secret.txt")
	f, err := ioutil.ReadFile(p)
	if err != nil {
		return "", errors.Trace(err)
	}
	return string(f), nil
}

func Request(method, route string, headers map[string]string, body io.Reader) (*http.Response, error) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	url := "http://127.0.0.1:9090" + route
	method = strings.ToUpper(method)

	client := &http.Client{}
	reqObj, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.Trace(err)
	}

	s := fmt.Sprintf("Bearer %s", Secrete)
	reqObj.Header.Add("Authorization", s)
	for key, value := range headers {
		reqObj.Header.Add(key, value)
	}

	resp, err := client.Do(reqObj)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return resp, nil
}

func EasyRequest(method, route string, headers map[string]string, body map[string]interface{}) (int, []byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return 0, []byte{}, errors.Trace(err)
	}
	_body := bytes.NewBuffer(data)

	resp, err := Request(method, route, headers, _body)
	if err != nil {
		return 0, []byte{}, errors.Trace(err)
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, result, errors.Trace(err)
	}
	return resp.StatusCode, result, nil
}

func UnmarshalRequest(method, route string, headers map[string]string, body map[string]interface{}, obj interface{}) error {
	_, content, err := EasyRequest(method, route, headers, body)
	if err != nil {
		return errors.Trace(err)
	}
	if err := json.Unmarshal(content, obj); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func HandleStreamResp(resp *http.Response, handler func(line []byte) (stop bool)) {
	go func() {
		buf := bufio.NewReader(resp.Body)
		defer resp.Body.Close()
		for {
			line, err := buf.ReadBytes('\n')
			if err != nil && err != io.EOF {
				return
			}
			if len(line) > 0 {
				line = bytes.TrimSpace(line)
				if stop := handler(line); stop {
					return
				}
			}
		}
	}()
}

func init() {
	var err error
	Secrete, err = getSecrete()
	if err != nil {
		panic(err)
	}
}