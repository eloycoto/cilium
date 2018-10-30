// Copyright 2018 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package togroups

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/cilium/cilium/pkg/logging"
	"github.com/cilium/cilium/pkg/logging/logfields"
)

// AND setup
// - aws:
//     securityGroups: [sg-baz1, sg-baz2]
//     region: us-west1
//     instanceLabels: [foo, bar]
// OR setup
// - group:
//   - aws:
//      securityGroupID: sg-baz1
//      region: us-east1
// - group:
//   - aws:
//      securityGroup: sg-baz2
//      region: us-west1

var (
	EC2_FILTER_MAPPING = map[string]string{
		POLICY_SECURITY_GROUP_ID_KEY:   "instances.group-id",
		POLICY_SECURITY_GROUP_NAME_KEY: "instances.group-name",
		POLICY_EC2_LABELS_KEY:          "instance.labels",
	}

	log = logging.DefaultLogger.WithField(logfields.LogSubsys, "ToGroupsAws")
)

const (
	awsLogLevel            = aws.LogDebugWithSigning
	AWS_DEFAULT_REGION_KEY = "AWS_DEFAULT_REGION"
	AWS_DEFAULT_REGION     = "eu-west-1"

	POLICY_REGION_KEY              = "region"
	POLICY_SECURITY_GROUP_ID_KEY   = "securityGroupID"
	POLICY_SECURITY_GROUP_NAME_KEY = "securityGroupName"
	POLICY_EC2_LABELS_KEY          = "labels"
)

// InitializeAWSAccount retrieve the env variables from the runtime and it
// iniliazes the account in the specified region.
func InitializeAWSAccount(region string) (aws.Config, error) {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, fmt.Errorf("Cannot initialize aws connector: %s", err)
	}
	cfg.LogLevel = awsLogLevel
	return cfg, nil
}

//GetInstancesFromFilter returns the instances IPs in aws EC2 filter by the
//given filter
func GetInstancesIpsFromFilter(filter map[string][]string) []string {
	region := AWS_DEFAULT_REGION
	regionValues, ok := filter["region"]
	if ok {
		region = regionValues[0]
	}

	input := &ec2.DescribeInstancesInput{}

	if len(filter) > 0 {
		input.Filters = []ec2.Filter{}
	}

	for key, val := range filter {
		newFilterName, ok := EC2_FILTER_MAPPING[key]
		if !ok {
			log.Warning("AWS policy key not recognized %s", key)
		}

		newFilter := ec2.Filter{
			Name:   aws.String(newFilterName),
			Values: val,
		}
		input.Filters = append(input.Filters, newFilter)

	}

	cfg, err := InitializeAWSAccount(region)
	if err != nil {
		log.Errorf("New error here!")
	}
	svc := ec2.New(cfg)
	req := svc.DescribeInstancesRequest(input)
	result, err := req.Send()

	if err != nil {
		log.Errorf("Can't get the data")
	}
	return awsDumpIpsFromRequest(result)
}

func GetDefaultRegion() string {
	val := os.Getenv(AWS_DEFAULT_REGION_KEY)
	if val != "" {
		return val
	}
	return AWS_DEFAULT_REGION
}

func awsDumpIpsFromRequest(req *ec2.DescribeInstancesOutput) []string {
	result := []string{}

	if len(req.Reservations) == 0 {
		return result
	}

	for _, reservation := range req.Reservations {
		for _, instance := range reservation.Instances {
			for _, iface := range instance.NetworkInterfaces {
				for _, ifaceIP := range iface.PrivateIpAddresses {
					result = append(result, string(*ifaceIP.PrivateIpAddress))
					if ifaceIP.Association != nil {
						result = append(result, string(*ifaceIP.Association.PublicIp))
					}
				}
			}
		}
	}
	return result
}
