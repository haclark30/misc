package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"misc/internal/models"
)

type HabitRuleStore interface {
	GetHabitRule(string) (*models.HabiticaHabitRule, error)
}

type DailyUpdater interface {
	ScoreDaily(string) error
}

type HabiticaMinHabitService struct {
	db      HabitRuleStore
	updater DailyUpdater
}

func NewHabitcaMinHabitService(db HabitRuleStore, updater DailyUpdater) HabiticaMinHabitService {
	return HabiticaMinHabitService{db: db, updater: updater}
}

func (h *HabiticaMinHabitService) CheckMinHabit(habitId string, currScore int) error {
	rule, err := h.db.GetHabitRule(habitId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Info("no habit rules found", "habitId", habitId)
			return nil
		}
		return fmt.Errorf("error getting habit rule: %w", err)
	}

	if currScore == rule.MinScore {
		if err := h.updater.ScoreDaily(rule.DailyId); err != nil {
			return fmt.Errorf("error scoring daily: %w", err)
		}
	}
	return nil
}
