package events

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	eb_scheduler "github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/manishMandal02/tabsflow-backend/config"
)

type scheduler struct {
	client *eb_scheduler.Client
}

func NewScheduler() *scheduler {
	return &scheduler{
		client: eb_scheduler.NewFromConfig(config.AWS_CONFIG),
	}
}

// creates a schedule
//
// name - name of the schedule
//
// dateTime - date & time to trigger the target. ex: at(yyyy-mm-ddThh:mm:ss)
func (s scheduler) CreateSchedule(id, dateTime string, event *string) error {

	scheduleExpression := fmt.Sprintf("at(%s)", dateTime)

	_, err := s.client.CreateSchedule(context.TODO(), &eb_scheduler.CreateScheduleInput{
		Name:               &id,
		ScheduleExpression: &scheduleExpression,
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     &config.NOTIFICATIONS_QUEUE_ARN,
			RoleArn: &config.SCHEDULER_ROLE_ARN,
			Input:   event,
			RetryPolicy: &types.RetryPolicy{
				MaximumRetryAttempts:     aws.Int32(5),
				MaximumEventAgeInSeconds: aws.Int32(720),
			},
		},
		ActionAfterCompletion: types.ActionAfterCompletionDelete,
	})

	if err != nil {
		return err
	}

	return nil
}

// creates a schedule
//
// name - name of the schedule
//
// dateTime - date & time to trigger the target. ex: at(yyyy-mm-ddThh:mm:ss)
func (s scheduler) UpdateSchedule(name, dateTime string) error {

	scheduleExpression := fmt.Sprintf("at(%s)", dateTime)

	_, err := s.client.UpdateSchedule(context.TODO(), &eb_scheduler.UpdateScheduleInput{
		Name:               &name,
		ScheduleExpression: &scheduleExpression,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s scheduler) DeleteSchedule(name string) error {
	_, err := s.client.DeleteSchedule(context.TODO(), &eb_scheduler.DeleteScheduleInput{
		Name: &name,
	})

	if err != nil {
		return err
	}
	return nil
}
