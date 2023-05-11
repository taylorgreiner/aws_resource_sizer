package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
)

func main() {
	fmt.Println("Welcome! This tool is to identify and size all resources within an AWS environment. All you need are AWS credentials with proper policies.")

	var aws_key string
	var aws_secret string
	var region string

	fmt.Print("AWS Key: ")
	_, err := fmt.Scanln(&aws_key)

	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	fmt.Print("AWS Secret: ")
	_, err = fmt.Scanln(&aws_secret)

	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	fmt.Print("AWS Region: ")
	_, err = fmt.Scanln(&region)

	if err != nil {
		fmt.Println("Error reading input:", err)
		return
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(aws_key, aws_secret, ""),
	})

	if err != nil {
		fmt.Println("Error creating session:", err)
		return
	}

	ec2Svc := ec2.New(sess)
	rdsSvc := rds.New(sess)

	file, err := os.Create("aws_resources.csv")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)
	defer csvWriter.Flush()

	csvWriter.Write([]string{"Resource Type", "ID", "Type/Class", "Size/State", "OS/DB Engine"})

	getEC2Sizing(ec2Svc, csvWriter)
	getEBSSizing(ec2Svc, csvWriter)
	getRDSSizing(rdsSvc, csvWriter)

	fmt.Println("The AWS resources have been written to aws_resources.csv")

	csvWriter.Flush()

	var exitval string
	fmt.Print("Press Enter key to exit...")
	_, err = fmt.Scanln(&exitval)
}

func getEC2Sizing(svc *ec2.EC2, csvWriter *csv.Writer) {
	input := &ec2.DescribeInstancesInput{}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("Error describing instances:", err)
		return
	}

	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			os := "Linux/UNIX"
			if instance.Platform != nil {
				os = *instance.Platform
			}
			csvWriter.Write([]string{"EC2 Instance", *instance.InstanceId, *instance.InstanceType, *instance.State.Name, os})
		}
	}
}

func getEBSSizing(svc *ec2.EC2, csvWriter *csv.Writer) {
	input := &ec2.DescribeVolumesInput{}
	result, err := svc.DescribeVolumes(input)
	if err != nil {
		fmt.Println("Error describing volumes:", err)
		return
	}

	for _, volume := range result.Volumes {
		csvWriter.Write([]string{"EBS Volume", *volume.VolumeId, fmt.Sprintf("%d GiB", *volume.Size), *volume.State})
	}
}

func getRDSSizing(svc *rds.RDS, csvWriter *csv.Writer) {
	input := &rds.DescribeDBInstancesInput{}
	result, err := svc.DescribeDBInstances(input)
	if err != nil {
		fmt.Println("Error describing RDS instances:", err)
		return
	}

	for _, instance := range result.DBInstances {
		csvWriter.Write([]string{"RDS Instance", *instance.DBInstanceIdentifier, *instance.DBInstanceClass, fmt.Sprintf("%d GiB", *instance.AllocatedStorage), *instance.Engine})
	}
}
