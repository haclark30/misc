package todoist

type Task struct {
	ID           string    `json:"id"`
	ProjectId    string    `json:"project_id"`
	SectionId    string    `json:"section_id"`
	Content      string    `json:"content"`
	Description  string    `json:"description"`
	IsCompleted  bool      `json:"is_completed"`
	Labels       []string  `json:"labels"`
	ParentId     string    `json:"parent_id"`
	Order        int       `json:"order"`
	Priority     uint      `json:"priority"`
	Due          Due       `json:"Due"`
	Url          string    `json:"url"`
	CommentCount uint      `json:"comment_count"`
	CreatedAt    string    `json:"created_at"`
	CreatorId    string    `json:"creator_id"`
	AssigneeId   string    `json:"assignee_id"`
	Duration     *Duration `json:"duration,omitempty"`
}

type Due struct {
	String      string  `json:"string"`
	Date        string  `json:"date"`
	IsRecurring bool    `json:"is_recurring"`
	Datetime    *string `json:"datetime,omitempty"`
	Timezome    *string `json:"timezone,omitempty"`
}

type Duration struct {
	Amount uint   `json:"amount"`
	Unit   string `json:"unit"`
}

type TaskFilterOptions struct {
	ProjectId string   `url:"project_id,omitempty"`
	SectionId string   `url:"section_id,omitempty"`
	Label     string   `url:"label,omitempty"`
	Filter    string   `url:"filter,omitempty"`
	Lang      string   `url:"lang,omitempty"`
	IDs       []string `url:"ids,comma,omitempty"`
}

type Stats struct {
	KarmaLastUpdate    float64             `json:"karma_last_update"`
	KarmaTrend         string              `json:"karma_trend"`
	DaysItems          []DayItem           `json:"days_items"`
	CompletedCount     int                 `json:"completed_count"`
	KarmaUpdateReasons []KarmaUpdateReason `json:"karma_update_reasons"`
	Karma              float64             `json:"karma"`
	WeekItems          []WeekItem          `json:"week_items"`
	ProjectColors      ProjectColors       `json:"project_colors"`
	Goals              Goals               `json:"goals"`
}

type DayItem struct {
	TotalCompleted int `json:"total_completed"`
}

type KarmaUpdateReason struct {
}

type WeekItem struct {
	TotalCompleted int `json:"total_completed"`
}

type ProjectColors struct {
}

type Goals struct {
	DailyGoal int `json:"daily_goal"`
}
