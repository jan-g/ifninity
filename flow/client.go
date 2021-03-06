package flow

import (
	"os"
	"net/http"
	"errors"
	"strings"
	"fmt"
)

var flowService string

func init() {
	flowService = os.Getenv("COMPLETER_BASE_URL")
	if flowService == "" {
		panic("COMPLETER_BASE_URL must be defined")
	}
}

type Flow interface {
	Id() string
	AddTerminationHook(fnName string) error
	Commit() error

	CompletedValue(value string) (Stage, error)
	InvokeFunction(fnPath string, contentType string, payload string) (Stage, error)
	AllOf(stages ...Stage) (Stage, error)
}

type Stage interface {
	Id() string
	ThenCompose(fnName string) (Stage, error)
	ThenApply(fnName string) (Stage, error)
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

func ThisFlow(id string) Flow {
	return f{id}
}

func (flow f) Id() string {
	return flow.fid
}

func postBlob(path string, arg string, contentType string) (string, error) {
	newMap := map[string]string{
		"Content-Type": contentType,
		"FnProject-DatumType": "blob",
		"FnProject-ResultStatus": "success",
	}
	return postRequest(path, arg, newMap)
}

func postHttpReq(path string, contentType string, payload string) (string, error) {
	newMap := map[string]string{
		"Content-Type": contentType,
		"FnProject-DatumType": "httpreq",
		"FnProject-Method": "POST",
	}
	return postRequest(path, payload, newMap)
}

func postRequest(path string, arg string, headers map[string]string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", flowService + path, strings.NewReader(arg))
	if err != nil {
		return "", err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	stageId := resp.Header.Get("FnProject-StageId")
	if stageId != "" {
		return stageId, nil
	}
	return "", fmt.Errorf("FnProject-StageId heading missing in postRequest response")
}


func (flowId f) AddTerminationHook(fnName string) error {
	url := "/graph/" + flowId.fid + "/terminationHook"

	_, err := postBlob(url, fnName, "application/golang")
	return err
}


func (flowId f) Commit() error {
	resp, err := http.Post(flowService + "/graph/" + flowId.fid + "/commit", "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return err
}

func (flowId f) CompletedValue(value string) (Stage, error) {
	url := "/graph/" + flowId.fid + "/completedValue"

	sid, err := postBlob(url, value, "text/plain")
	if err != nil {
		return s{}, nil
	}

	return s{fid: flowId.fid, sid: sid}, nil
}

func (flowId f) InvokeFunction(fnName string, contentType string, body string) (Stage, error) {
	url := "/graph/" + flowId.fid + "/invokeFunction?functionId=" + fnName

	sid, err := postHttpReq(url, contentType, body)
	if err != nil {
		return s{}, nil
	}

	return s{fid: flowId.fid, sid: sid}, nil
}

func (flowId f) AllOf(stages ...Stage) (Stage, error) {
	var cids []string
	for _, stage := range stages {
		cids = append(cids, stage.Id())
	}
	url := "/graph/" + flowId.fid + "/allOf?cids=" + strings.Join(cids, ",")

	sid, err := postHttpReq(url, "", "")
	if err != nil {
		return s{}, nil
	}

	return s{fid: flowId.fid, sid: sid}, nil
}


func ThisStage(fid string, sid string) Stage {
	return s{fid: fid, sid: sid}
}

func (stage s) Id() string {
	return stage.sid
}



func (stage s) thenOperate(fnName string, op string) (Stage, error) {
	url := "/graph/" + stage.fid + "/stage/" + stage.sid + "/" + op

	sid, err := postRequest(url, fnName, nil)
	if err != nil {
		return s{}, err
	}

	return s{fid: stage.fid, sid: sid}, nil
}


func (stage s) ThenCompose(fnName string) (Stage, error) {
	return stage.thenOperate(fnName, "thenCompose")
}

func (stage s) ThenApply(fnName string) (Stage, error) {
	return stage.thenOperate(fnName, "thenApply")
}
