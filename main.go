package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codepipeline"

	"github.com/appleboy/easyssh-proxy"
)

// Local application variables
var (
	awsSession *session.Session
)

type CodePipelineEvent struct {
	Job struct {
		ID   string `json:"id"`
		Data struct {
			ActionConfiguration struct {
				Configuration struct {
					UserParameters string `json:"UserParameters"`
				} `json:"configuration"`
			} `json:"actionConfiguration"`
		} `json:"data"`
	} `json:"CodePipeline.job"`
}

type TiDBClusterVar struct {
	TiDBPublicIp    string `json:"TiDBPublicIp"`
	TiDBInstanceID  string `json:"TiDBInstanceID"`
	TiKV1InstanceID string `json:"TiKV1InstanceID"`
	TiKV2InstanceID string `json:"TiKV2InstanceID"`
}

func StartChaos(instanceid string) (string, error) {
	ssh := &easyssh.MakeConfig{
		User:   "ec2-user",
		Server: "hostname",
		// Optional key or Password without either we try to contact your agent SOCKET
		//Password: "password",
		// Paste your source content of private key
		Key: `-----BEGIN RSA PRIVATE KEY-----
You key here, also can 
-----END RSA PRIVATE KEY-----`,
		// KeyPath: "/Users/username/.ssh/id_rsa",
		Port:    "22",
		Timeout: 60 * time.Second,

		// Parse PrivateKey With Passphrase
		Passphrase: "",
	}
	// Call Run method with command you want to run on remote server.
	// trigger Chaos Mesh to run chaos, we can later using client go to replace. This is ugly
	stdout, _, _, err := ssh.Run(fmt.Sprintf("sh run.sh %s", instanceid), 60*time.Second)
	// Handle errors
	if err != nil {
		return "", err
	} else {
		return fmt.Sprintf("stdout is %s", stdout), nil
	}

}

func HandleRequest(ctx context.Context, event *CodePipelineEvent) (string, error) {
	// Start a new CodePipeline service
	pipeline := codepipeline.New(awsSession)

	params := event.Job.Data.ActionConfiguration.Configuration.UserParameters
	v := &TiDBClusterVar{}
	err := json.Unmarshal([]byte(params), &v)
	if err != nil {
		return "", err
	}

	input := &codepipeline.PutJobSuccessResultInput{
		JobId: aws.String(event.Job.ID),
	}

	out, err := StartChaos(v.TiKV2InstanceID)
	if err != nil {
		return "", err
	}

	output, err := pipeline.PutJobSuccessResult(input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Hello !\n %s\n%s \n%s", event.Job.ID, output.String(), out), nil
}

func main() {

	// Create a new AWS session
	// We can also make region config
	if awsSession == nil {
		awsSession = session.Must(session.NewSession(&aws.Config{
			Region: aws.String("us-east-2"),
		}))
	}
	lambda.Start(HandleRequest)
}
