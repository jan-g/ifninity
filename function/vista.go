package function

import (
	"github.com/gin-gonic/gin"

	"github.com/jan-g/ifninity/flow"
	"mime"
	"strings"
	"mime/multipart"
	"io/ioutil"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"bytes"
)


func Vista(c *gin.Context) {
	// Are we a main invocation?
	flowId := c.GetHeader("fnproject-flowid")
	if flowId == "" {
		handleRequest(c)
	} else {
		fl := flow.ThisFlow(flowId)
		// Extract the other salient pieces
		stageId := c.GetHeader("fnproject-stageid")
		var st flow.Stage
		if stageId != "" {
			st = flow.ThisStage(flowId, stageId)
		}

		// Extract the bits from the multipart
		mediaType, params, err := mime.ParseMediaType(c.Request.Header.Get("Content-Type"))
		if err != nil {
			panic(err)
		}
		var items []string
		if strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(c.Request.Body, params["boundary"])
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					panic(err)
				}
				slurp, err := ioutil.ReadAll(p)
				if err != nil {
					panic(err)
				}
				items = append(items, string(slurp))
			}
		}

		fmt.Printf("%s %s %s\n", fl.Id(), st.Id(), items[0])
		stage[strings.Split(items[0], "|")[0]](c, fl, st, items)
	}
}


type RunSpec struct {
	NumImages int
	NumBytesInData int
	SlowFunction string
	FastFunction string
	NotifyFinished bool
	NotifyURL string
}

func handleRequest(c *gin.Context) {
	var rs RunSpec
	err := c.BindJSON(&rs)
	if err != nil {
		c.JSON(400, map[string]string{"error": "please specify load setup"})
		return
	}

	fl, err := flow.NewFlow("t/flow-load-test-vista")
	if err != nil {
		panic(err)
	}
	notifyURL := ""
	if rs.NotifyFinished {
		notifyURL = rs.NotifyURL
	}
	fl.AddTerminationHook("terminate|" + notifyURL)
	fmt.Printf("Input RunSpec = %+v\n", rs)

	var fakeScrapes []string
	for i := 0; i < rs.NumImages; i++ {
		fakeScrapes = append(fakeScrapes, fmt.Sprintf("Image_%d", i))
	}

	st, err := fl.CompletedValue(strings.Join(fakeScrapes, "|"))
	if err != nil {
		panic(err)
	}

	st, err = st.ThenCompose("start-flows|" + rs.FastFunction + "|" + rs.SlowFunction + "|" + strconv.Itoa(rs.NumBytesInData))
	if err != nil {
		panic(err)
	}

	fl.Commit()
}

// Stages

var stage = map[string]func(*gin.Context, flow.Flow, flow.Stage, []string){
	"terminate": terminationHook,
	"start-flows": startFlows,
	"httpresp-to-string": httpRespToString,
	"second-slow": secondSlow,
	"fast-handoff": fastHandoff,
}

func terminationHook(c *gin.Context, fl flow.Flow, st flow.Stage, items []string) {
	notifyURL := strings.Split(items[0], "|")[1]
	if notifyURL != "" {
		resp, err := http.Post(notifyURL, "text/plain", bytes.NewReader([]byte{}))
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Termination callback result: %+v %v\n", resp, b)
	}

	fmt.Printf("Terminating flow %v with stage %v items = %v\n", fl, st, items)
	returnEmpty(c)
}


func startFlows(c *gin.Context, fl flow.Flow, st flow.Stage, items []string) {
	/*
        (resp -> {
            try {
                List<FlowFuture<?>> pendingTasks = resp
                        .stream()
                        .map(scrapeResult -> {
                            try {
                                return Flows.currentFlow()
                                        .invokeFunction(input.slowFunction, new byte[input.numBytesInData], byte[].class)
										[ thenApply ]
                                        .thenCompose((plateResp) -> {   // second-slow

                            } catch (Exception e) {
                                e.printStackTrace();
                                throw e;
                            }
                        }).collect(Collectors.toList());

                return Flows.currentFlow()
                        .allOf(pendingTasks.toArray(new FlowFuture[pendingTasks.size()]));
            } catch (Exception e) {
                e.printStackTrace();
                throw e;
            }

        }).whenComplete((v, throwable) -> {
            if (throwable != null) {
                System.err.println("Failed!");
            } else {
                System.err.println("Succeeded.");
            }
        });
	*/
	closure := strings.Split(items[0], "|")
	fastFunction := closure[1]
	slowFunction := closure[2]
	numBytesInData, _ := strconv.Atoi(closure[3])
	stages := strings.Split(items[1], "|")
	var futures []flow.Stage

	for _ = range stages {
		stage, err := fl.InvokeFunction(slowFunction, "application/octet-stream", string(make([]byte, numBytesInData)))
		if err != nil {
			panic(err)
		}
		stage, err = stage.ThenApply("httpresp-to-string")
		if err != nil {
			panic(err)
		}
		stage, err = stage.ThenCompose("second-slow|" + fastFunction + "|" + slowFunction)
		if err != nil {
			panic(err)
		}
		futures = append(futures, stage)
	}

	stage, err := fl.AllOf(futures...)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Starting parallel flows with stages %+v\n", stages)
	returnStage(c, stage)
}

func httpRespToString(c *gin.Context, fl flow.Flow, st flow.Stage, items []string) {
	returnBlob(c, items[1])
}

func secondSlow(c *gin.Context, fl flow.Flow, st flow.Stage, items []string) {
	/*
		.thenCompose((plateResp) -> {   // second-slow
			try {
				return Flows.currentFlow()
				.invokeFunction(input.slowFunction, new byte[input.numBytesInData], byte[].class)
				[ thenApply ]
				.thenCompose((drawResp) -> {   // fastHandoff
				try {
				return Flows.currentFlow().allOf(
				Flows.currentFlow().invokeFunction(input.fastFunction, new byte[input.numBytesInData], byte[].class),
				Flows.currentFlow().invokeFunction(input.fastFunction, new byte[input.numBytesInData], byte[].class)
			);
			} catch (Exception e) {
				e.printStackTrace();
				throw e;
			}
			});
			} catch (Exception e) {
			e.printStackTrace();
			throw e;
			}
		});
	*/
	closure := strings.Split(items[0], "|")
	fastFunction := closure[1]
	slowFunction := closure[2]
	input := items[1]

	stage, err := fl.InvokeFunction(slowFunction, "application/octet-stream", input)
	if err != nil {
		panic(err)
	}
	stage, err = stage.ThenApply("httpresp-to-string")
	if err != nil {
		panic(err)
	}
	stage, err = stage.ThenCompose("fast-handoff|" + fastFunction)
	if err != nil {
		panic(err)
	}
	returnStage(c, stage)
}

func fastHandoff(c *gin.Context, fl flow.Flow, st flow.Stage, items []string) {
	/*
				.thenCompose((drawResp) -> {   // fastHandoff
				try {
				return Flows.currentFlow().allOf(
				Flows.currentFlow().invokeFunction(input.fastFunction, new byte[input.numBytesInData], byte[].class),
				Flows.currentFlow().invokeFunction(input.fastFunction, new byte[input.numBytesInData], byte[].class)
			);
	*/
	closure := strings.Split(items[0], "|")
	fastFunction := closure[1]
	input := items[1]

	stage1, err := fl.InvokeFunction(fastFunction, "application/octet-stream", input)
	if err != nil {
		panic(err)
	}
	stage1, err = stage1.ThenApply("httpresp-to-string")
	if err != nil {
		panic(err)
	}
	stage2, err := fl.InvokeFunction(fastFunction, "application/octet-stream", input)
	if err != nil {
		panic(err)
	}
	stage2, err = stage2.ThenApply("httpresp-to-string")
	if err != nil {
		panic(err)
	}
	stage, err := fl.AllOf(stage1, stage2)
	if err != nil {
		panic(err)
	}
	returnStage(c, stage)
}