package function

import (
	"github.com/gin-gonic/gin"

	"github.com/jan-g/ifninity/flow"
	"os"
	"mime"
	"strings"
	"mime/multipart"
	"io/ioutil"
	"fmt"
	"io"
)

/*
package com.fnproject.fn.examples;

import com.fnproject.fn.api.flow.Flow;
import com.fnproject.fn.api.flow.FlowFuture;
import com.fnproject.fn.api.flow.Flows;
import org.apache.http.client.HttpClient;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.impl.client.DefaultHttpClient;

import java.io.IOException;
import java.io.Serializable;
import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

public class LoadTest {

    public static class RunSpec implements Serializable {
        public int numImages;
        public int numBytesInData;
        public String slowFunction;
        public String fastFunction;
        public boolean notifyFinished;
        public String notifyURL;
    }

    public String handleRequest(RunSpec input) {
        Flow fl = Flows.currentFlow();
        if (input.notifyFinished) {
            fl.addTerminationHook((state) -> {
                HttpPost httppost = new HttpPost(input.notifyURL);
                httppost.setEntity(null);

                HttpClient httpclient = new DefaultHttpClient();
                try {
                    httpclient.execute(httppost);
                } catch (IOException e) {
                    // We can't do much at this stage really...
                    e.printStackTrace();
                }
            });
        }

        // Start by pretending we scrape numImages images
        ArrayList<String> fakeScrapes = new ArrayList<>();
        for (int i = 0; i < input.numImages; i++) {
            fakeScrapes.add("Image_" + i);
        }
        FlowFuture<ArrayList<String>> future = fl.completedValue(fakeScrapes);

        // Fan out to the processing nodes and replicate Vista's function calls
        future.thenCompose(resp -> {
            try {
                List<FlowFuture<?>> pendingTasks = resp
                        .stream()
                        .map(scrapeResult -> {
                            try {
                                return Flows.currentFlow()
                                        .invokeFunction(input.slowFunction, new byte[input.numBytesInData], byte[].class)
                                        .thenCompose((plateResp) -> {
                                            try {
                                                return Flows.currentFlow()
                                                        .invokeFunction(input.slowFunction, new byte[input.numBytesInData], byte[].class)
                                                        .thenCompose((drawResp) -> {
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

        return "Run started.\n";
    }

}
 */

func Vista(c *gin.Context) {
	// Are we a main invocation?
	flowId := c.GetHeader("fnproject-flowid")
	if flowId == "" {
		handleRequest(c)
	} else {
		// Extract the other salient pieces
		stageId := c.GetHeader("fnproject-stageid")

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

		if stageId != "" {
			stage[strings.Split(items[0], "|")[0]](c, flowId, stageId, items)
		}
	}
}


type RunSpec struct {
	NumImages int64
	NumBytesInData int64
	SlowFunction string
	FastFunction string
	NotifyFinished bool
	NotifyURL string
}

func handleRequest(c *gin.Context) {
	// TODO: generate stages
	var rs RunSpec
	err := c.BindJSON(&rs)
	if err != nil {
		panic(err)
	}

	flow, err := flow.NewFlow("t/flow-load-test-vista")
	if err != nil {
		panic(err)
	}
	flow.AddTerminationHook("terminate|" + rs.NotifyURL)
	fmt.Printf("Input RunSpec = %+v\n", rs)
	flow.Commit()
}

// Stages

var stage = map[string]func(*gin.Context, string, string, []string){
	"terminate": terminationHook,
}

func terminationHook(c *gin.Context, flowId string, stageId string, items []string) {
	// TODO: callback with termination notification
	/*
	HttpPost httppost = new HttpPost(input.notifyURL);
	httppost.setEntity(null);

	HttpClient httpclient = new DefaultHttpClient();
	try {
		httpclient.execute(httppost);
	} catch (IOException e) {
	// We can't do much at this stage really...
	e.printStackTrace();
	}
	*/
	fmt.Printf("Terminating flow %s with stage %s items = %v\n", flowId, stageId, items)
	returnDatum(c, "empty", true, "", nil)
}

var notifyURL string

func init() {
	notifyURL = os.Getenv("NOTIFY_URL")
	if notifyURL == "" {
		panic("NOTIFY_URL must be set")
	}
}


func returnDatum(c *gin.Context, datumType string, success bool, contentType string, body []byte) {
	c.Writer.WriteHeaderNow()
	s := "success"
	if !success {
		s = "failure"
	}
	ct := ""
	if contentType != "" {
		ct = "Content-Type: " + contentType + "\r\n"
	}
	c.Writer.Write([]byte(
		"HTTP/1.1 200\r\n" +
		"FnProject-DatumType: " + datumType + "\r\n" +
		"FnProject-ResultStatus: " + s + "\r\n" +
		ct +
		"\r\n"))
	c.Writer.Write(body)
}
