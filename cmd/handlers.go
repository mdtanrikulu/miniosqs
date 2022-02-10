package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// reference: https://docs.aws.amazon.com/AmazonS3/latest/dev/notification-content-structure.html
type Event struct {
	EventVersion string `json:"eventVersion"`
	EventSource  string `json:"eventSource"`
	AwsRegion    string `json:"awsRegion"`
	EventTime    string `json:"eventTime"`
	EventName    string `json:"eventName"`
	UserIdentity struct {
		PrincipalId string `json:"principalId"`
	}
	RequestParameters struct {
		SourceIPAddress string `json:"sourceIPAddress"`
	}
	ResponseElements struct {
		XAmzRequestId string `json:"x-amz-request-id"`
		XAmzId2       string `json:"x-amz-id-2"`
	}
	S3 struct {
		S3SchemaVersion string `json:"s3SchemaVersion"`
		ConfigurationId string `json:"configurationId"`
		Bucket          struct {
			Name          string `json:"name"`
			OwnerIdentity struct {
				PrincipalId string `json:"principalId"`
			}
			Arn string `json:"arn"`
		}
		Object struct {
			Key       string `json:"key"`
			Size      int    `json:"size"`
			ETag      string `json:"eTag"`
			VersionId string `json:"versionId"`
			Sequencer string `json:"sequencer"`
		}
	}
}

type Records struct {
	Records []Event
}

type AuditLog struct {
	Version      string `json:"version"`
	DeploymentID string `json:"deploymentid"`
	Time         string `json:"time"`
	Trigger      string `json:"trigger"`
	Api          struct {
		Name           string `json:"name"`
		Bucket         string `json:"bucket"`
		Object         string `json:"object"`
		Status         string `json:"status"`
		StatusCode     int    `json:"statusCode"`
		RX             int    `json:"rx"`
		TX             int    `json:"tx"`
		TimeToResponse string `json:"timeToResponse"`
	}
	RemoteHost   string `json:"remotehost"`
	RequestID    string `json:"requestID"`
	UserAgent    string `json:"userAgent"`
	RequestQuery struct {
		XID string `json:"x-id"`
	}
	RequestHeader struct {
		AmzSdkInvocationID string `json:"Amz-Sdk-Invocation-Id"`
		AmzSdkRequest      string `json:"Amz-Sdk-Request"`
		Authorization      string `json:"Authorization"`
		Connection         string `json:"Connection"`
		ContentLength      string `json:"Content-Length"`
		ContentType        string `json:"Content-Type"`
		UserAgent          string `json:"User-Agent"`
		XAmzContentSha256  string `json:"X-Amz-Content-Sha256"`
		XAmzDate           string `json:"X-Amz-Date"`
		XAmzUserAgent      string `json:"X-Amz-User-Agent"`
	}
	ResponseHeader struct {
		AcceptRanges            string `json:"Accept-Ranges"`
		ContentLength           string `json:"Content-Length"`
		ContentSecurityPolicy   string `json:"Content-Security-Policy"`
		ETag                    string `json:"ETag"`
		Server                  string `json:"Server"`
		StrictTransportSecurity string `json:"Strict-Transport-Security"`
		Vary                    string `json:"Vary"`
		XAmzRequestID           string `json:"X-Amz-Request-Id"`
		XContentTypeOptions     string `json:"X-Content-Type-Options"`
		XXssProtection          string `json:"X-Xss-Protection"`
	}
}

func (app *application) notifySqs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", 405)
		return
	}

	if r.ContentLength == 0 {
		// minio appears to first make an empty post with each event
		app.infoLog.Printf("Empty request body")
		fmt.Fprintf(w, "Empty body")
		return
	}

	dec := json.NewDecoder(r.Body)

	var newE = func() interface{} {
		if app.messageType == "AUDIT" {
			c := new(AuditLog)
			return c
		} else if app.messageType == "EVENT" {
			c := new(Records)
			return c
		} else {
			err := fmt.Errorf("Unknown log type: %s", app.messageType)
			app.errorLog.Printf("Unknown log type: %s", err)
			app.serverError(w, err)
			return nil
		}
	}

	e := newE()

	err := dec.Decode(&e)
	if err != nil {
		app.errorLog.Printf("Failed to decode JSON: %s", err)
		app.serverError(w, err)
		return
	}
	app.infoLog.Printf("Event: %+v", e)

	j, err := json.Marshal(e)
	if err != nil {
		app.errorLog.Printf("JSON marshal error: %s", err)
		app.serverError(w, err)
		return
	}

	result, err := app.sqsClient.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(string(j)),
		QueueUrl:    &app.queueUrl,
	})
	if err != nil {
		app.errorLog.Printf("Failed to send message: %s", err)
		app.serverError(w, err)
		return
	}
	fmt.Fprintf(w, "Message ID: %s", *result.MessageId)
	app.infoLog.Printf("SQS message created with ID: %s", *result.MessageId)
}
