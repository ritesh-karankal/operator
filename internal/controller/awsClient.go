package controller

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func awsClient(region string) *ec2.Client {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		fmt.Println("Error loading AWS config:", err)
		os.Exit(1)
	}

	return ec2.NewFromConfig(cfg)
}
