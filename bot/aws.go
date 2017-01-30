package bot

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func (b *Bot) getInstanceIPs(instanceIDs []string) (map[string]string, error) {
	instanceIPs := map[string]string{}
	sess, err := session.NewSession()
	if err != nil {
		return instanceIPs, err
	}
	svc := ec2.New(sess, &aws.Config{})
	filterValues := []*string{}
	for _, id := range instanceIDs {
		filterValues = append(filterValues, aws.String(id))
	}
	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-id"),
				Values: filterValues,
			},
		},
	}
	resp, err := svc.DescribeInstances(input)
	if err != nil {
		return instanceIPs, err
	}
	for idx, _ := range resp.Reservations {
		for _, inst := range resp.Reservations[idx].Instances {
			if *inst.PrivateIpAddress != "" {
				instanceIPs[*inst.InstanceId] = *inst.PrivateIpAddress
			}
		}
	}
	return instanceIPs, nil
}
