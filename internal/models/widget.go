package models

type WidgetResponse struct {
	HabiticaWaterValue int      `json:"habitica_water_value"`
	HabiticaWaterGoal  int      `json:"habitica_water_goal"`
	HabiticaReadValue  int      `json:"habitica_read_value"`
	HabiticaReadGoal   int      `json:"habitica_read_goal"`
	HabiticaDailysDone int      `json:"habitica_dailys_done"`
	HabiticaDailysDue  int      `json:"habitica_dailys_due"`
	TodoistTasksDone   int      `json:"todoist_tasks_done"`
	TodoistTasksGoal   int      `json:"todoist_tasks_goal"`
	TodoistTasksDue    int      `json:"todoist_tasks_due"`
	Errors             []string `json:"errors,omitempty"`
}
