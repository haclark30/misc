package todoist

type TaskResp struct {
	Tasks []Task `json:"results"`
}

type Task struct {
	ID          string  `json:"id"`
	ProjectId   string  `json:"project_id"`
	SectionId   *string `json:"section_id"`
	Content     string  `json:"content"`
	Description string  `json:"description"`
	// IsCompleted  bool      `json:"is_completed"`
	// Labels   []string `json:"labels"`
	// ParentId *string  `json:"parent_id"`
	// Order        int       `json:"order"`
	// Priority uint `json:"priority"`
	Due Due `json:"due"`
	// Url          string    `json:"url"`
	// CommentCount uint      `json:"comment_count"`
	// CreatedAt string `json:"added_ad"`
	// CreatorId string `json:"added_by_uid"`
	// AssigneeId   string    `json:"assignee_id"`
	// Duration *Duration `json:"duration,omitempty"`
}

type Due struct {
	String      string  `json:"string"`
	Date        *string `json:"date"`
	IsRecurring bool    `json:"is_recurring"`
}

type Duration struct {
	Amount uint   `json:"amount"`
	Unit   string `json:"unit"`
}

type TaskFilterOptions struct {
	Query string `url:"query,omitempty"`
	Lang  string `url:"lang,omitempty"`
	Limit int    `url:"limit,omitempty"`
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
