package services

import (
	"fmt"
	"log/slog"
	"misc/internal/models"
	"strings"
)

type TodoistHabiticaRuleStore interface {
	GetTodoistHabiticaTextRules() ([]models.TodoistHabiticaTextRule, error)
	GetTodoistHabiticaProjectRule(string) (models.TodoistHabiticaProjectRule, error)
}

type TodoistHabiticaService struct {
	db      TodoistHabiticaRuleStore
	updater DailyUpdater
}

func NewTodoistHabiticaService(db TodoistHabiticaRuleStore, updater DailyUpdater) TodoistHabiticaService {
	return TodoistHabiticaService{
		db:      db,
		updater: updater,
	}
}

func (s *TodoistHabiticaService) ScoreTask(taskStr, projectId string) error {
	// check text rules first, if we hit one score the task and return
	rules, err := s.db.GetTodoistHabiticaTextRules()

	if err != nil {
		return fmt.Errorf("error getting text rules: %w", err)
	}

	slog.Info("got text rules", "rules", rules, "taskStr", taskStr)
	for _, rule := range rules {
		if strings.HasPrefix(strings.ToLower(taskStr), rule.Rule) {
			if err := s.updater.ScoreDaily(rule.HabitId); err != nil {
				return fmt.Errorf("error scoring habit: %w", err)
			}
			return nil
		}
	}

	// check project rule
	rule, err := s.db.GetTodoistHabiticaProjectRule(projectId)
	slog.Info("got project rule", "rule", rule)
	if err != nil {
		return fmt.Errorf("error getting project rule: %w", err)
	}
	err = s.updater.ScoreDaily(rule.HabitId)
	if err != nil {
		return fmt.Errorf("error scoring habit: %w", err)
	}
	return nil
}
