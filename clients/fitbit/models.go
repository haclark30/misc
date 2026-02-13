package fitbit

type ActivityResponse struct {
	Activities []Activity    `json:"activities"`
	Summary    FitbitSummary `json:"summary"`
	Goals      FitbitGoals   `json:"goals"`
}

type Activity struct {
	Name     string `json:"name"`
	Duration int    `json:"duration"`
}

type WeightResponse struct {
	WeightRecords []WeightRecord `json:"weight"`
}

type WeightRecord struct {
	Weight float64 `json:"weight"`
	Date   string  `json:"date"`
}

type FitbitSummary struct {
	Steps             int `json:"steps"`
	VeryActiveMinutes int `json:"veryActiveMinutes"`
}

type FitbitGoals struct {
	ActiveMinutes int `json:"activeMinutes"`
	Steps         int `json:"steps"`
}
