package japi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sudachen.xyz/pkg/errstr"
)

type Agent struct {
	verbose func(string, ...interface{})
	baseUrl string
	http    *http.Client
}

/*
Client describes node client options
*/
type Remote struct {
	Endpoint string
	Verbose  func(string, ...interface{})
}

/*
New creates new node Agent
*/
func (c Remote) New() *Agent {
	verbose := c.Verbose
	if verbose == nil {
		verbose = func(string, ...interface{}) {}
	}
	return &Agent{
		baseUrl: "http://" + c.Endpoint + "/v1",
		verbose: verbose,
		http:    &http.Client{},
	}
}

func (c *Agent) post(api string, in, out interface{}) (err error) {
	data, err := json.Marshal(in)
	if err != nil {
		return
	}
	url := c.baseUrl + api
	c.verbose("request: %v, body: %s", url, data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	resBody, _ := ioutil.ReadAll(res.Body)
	c.verbose("response body: %s", resBody)

	if res.StatusCode != http.StatusOK {
		rb := struct {
			Error string `json:"error"`
		}{}
		if json.Unmarshal(resBody, &rb) == nil && rb.Error != "" {
			err = errstr.New(rb.Error)
		} else {
			err = fmt.Errorf("`%v` response status code: %d", api, res.StatusCode)
		}
		c.verbose("request failed with error " + err.Error())
		return
	}

	return json.Unmarshal(resBody, out)
}

