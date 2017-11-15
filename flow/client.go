package flow

import (
	"os"
	"net/http"
	"errors"
	"strings"
)

var flowService string

func init() {
	flowService = os.Getenv("COMPLETER_BASE_URL")
	if flowService == "" {
		panic("COMPLETER_BASE_URL must be defined")
	}
}

type Flow interface {
	AddTerminationHook(fnName string) error
	Commit() error

	CompletedValue(value string) (Stage, error)
}

type Stage interface {
	ThenCompose(fnName string) (Stage, error)
}

type f struct {
	fid string
}

type s struct {
	fid string
	sid string
}

func NewFlow(function string) (Flow, error) {
	resp, err := http.Post(flowService + "/graph?functionId=" + function, "", nil)
	if err != nil {
		return f{}, err
	}
	defer resp.Body.Close()
	flowId := resp.Header.Get("FnProject-FlowID")
	if flowId == "" {
		return f{}, errors.New("FnProject-FlowID header missing in NewFlow response")
	}
	return f{flowId}, nil
}

func (flowId f) AddTerminationHook(fnName string) error {
	url := "/graph/" + flowId.fid + "/terminationHook"

	_, err := postRequest(url, fnName, nil)
	return err
}

func postRequest(path string, arg string, headers map[string]string) (http.Header, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", flowService + path, strings.NewReader(arg))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/mock-blob")
	req.Header.Set("FnProject-DatumType", "blob")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	defer resp.Body.Close()

	return resp.Header, err
}

func (flowId f) Commit() error {
	resp, err := http.Post(flowService + "/graph/" + flowId.fid + "/commit", "", nil)
	defer resp.Body.Close()

	return err
}

func (flowId f) CompletedValue(value string) (Stage, error) {
	url := "/graph/" + flowId.fid + "/completedValue"

	h, err := postRequest(url, value, map[string]string{"FnProject-ResultStatus": "success"})
	if err != nil {
		return s{}, nil
	}

	return s{fid: flowId.fid, sid: h.Get("FnProject-StageId")}, nil
}

func (stage s) ThenCompose(fnName string) (Stage, error) {
	url := "/graph/" + stage.fid + "/stage/" + stage.sid + "/thenCompose"

	h, err := postRequest(url, fnName, nil)
	if err != nil {
		return s{}, nil
	}

	return s{fid: stage.fid, sid: h.Get("FnProject-StageId")}, nil
}