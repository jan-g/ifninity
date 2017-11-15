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
}

type f struct {
	fid string
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
	url := flowService + "/graph/" + flowId.fid + "/terminationHook"

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, strings.NewReader(fnName))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/mock-blob")
	req.Header.Set("FnProject-DatumType", "blob")

	resp, err := client.Do(req)
	defer resp.Body.Close()

	return nil
}

func (flowId f) Commit() error {
	resp, err := http.Post(flowService + "/graph/" + flowId.fid + "/commit", "", nil)
	defer resp.Body.Close()

	return err
}