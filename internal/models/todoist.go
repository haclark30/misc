package models

type TodoistWebhook struct {
	EventName string       `json:"event_name"`
	EventData TodoistEvent `json:"event_data"`
}

type TodoistEvent struct {
	Content     string `json:"content"`
	Description string `json:"description"`
	ProjectId   string `json:"project_id"`
	Id          string `json:"id"`
}

type TodoistHabiticaTextRule struct {
	Name    string
	Rule    string
	HabitId string
}

type TodoistHabiticaProjectRule struct {
	Name      string
	ProjectId string
	HabitId   string
}
