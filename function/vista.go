package function

import "github.com/gin-gonic/gin"

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
		if stageId != "" {
			stage[stageId](c, flowId)
		}
	}
}

var stage = map[string]func(*gin.Context, string){
	"0": terminationHook,
}

func handleRequest(c *gin.Context) {
	// TODO: generate stages
}

// Stages

func terminationHook(c *gin.Context, flowId string) {
	// TODO: callback with termination notification
	returnDatum(c, "empty", true, "", nil)
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