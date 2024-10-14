package events

import (
	"context"

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

func (s scheduler) CreateSchedule(name, scheduleExpression string) error {

	_, err := s.client.CreateSchedule(context.TODO(), &eb_scheduler.CreateScheduleInput{
		Name:               &name,
		ScheduleExpression: &scheduleExpression,
		FlexibleTimeWindow: &types.FlexibleTimeWindow{
			Mode: types.FlexibleTimeWindowModeOff,
		},
		Target: &types.Target{
			Arn:     &config.SCHEDULER_ARN,
			RoleArn: &config.SCHEDULER_ROLE_ARN,
			Input:   &name,
		},
	})

	/*

			client.CreateSchedule(context.TODO(), &scheduler.CreateScheduleInput{
			Name:                       &note.Id,
			ScheduleExpression:         &note.Remainder,
			ScheduleExpressionTimezone: &config.TIMEZONE,
			FlexibleTimeWindow: &scheduler.FlexibleTimeWindow{
				Mode: "OFF",
			},
			Target: &types.Target{
				Arn:     &config.SCHEDULER_ARN,
				RoleArn: &config.SCHEDULER_ROLE_ARN,
				Input:   &note.Remainder,
			},
		})
	*/

	if err != nil {
		return err
	}

	return nil
}

func (s scheduler) UpdateSchedule(name, scheduleExpression string) error {
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
