package models

const HabiticaHabitType = "habit"
const HabiticaDailyType = "daily"
const HabiticaTodoType = "todo"

type HabiticaWebhook struct {
	Type      string              `json:"type"`
	Direction string              `json:"direction"`
	Task      HabiticaWebhookTask `json:"task"`
}

type HabiticaWebhookTask struct {
	Id   string `json:"id"`
	Up   int    `json:"counterUp"`
	Down int    `json:"counterDown"`

	Type  string `json:"type"`
	Text  string `json:"text"`
	Notes string `json:"notes"`
}

type HabiticaHabitRule struct {
	HabitId  string
	DailyId  string
	MinScore int
}
