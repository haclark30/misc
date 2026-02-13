package services

import (
	"log"
	"misc/clients/fitbit"
)

type FitbitService struct {
	fitClient fitbit.FitbitClient
}

func NewFitbitService(client fitbit.FitbitClient) FitbitService {
	return FitbitService{client}
}

func (f FitbitService) GetWorkouts() []fitbit.Activity {
	activity, err := f.fitClient.GetFitbitActivity()
	if err != nil {
		log.Fatal(err)
	}
	var retActivities []fitbit.Activity

	for _, act := range activity.Activities {
		if act.Name != "Walk" {
			retActivities = append(retActivities, act)
		}
	}
	return retActivities
}
