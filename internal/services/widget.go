package services

import (
	"log/slog"
	"misc/clients/habitica"
	"misc/clients/todoist"
	"misc/internal/models"
	"sync"
)

type widgetService struct {
	habTaskRepo HabiticaTaskRepository
	tdTaskRepo  TodoistTaskRepository
}

type HabiticaTaskRepository interface {
	GetHabits() ([]habitica.Habit, error)
	GetDailys() ([]habitica.Daily, error)
}

type TodoistTaskRepository interface {
	GetTasks(*todoist.TaskFilterOptions) ([]todoist.Task, error)
	GetStats() (todoist.Stats, error)
}

func NewWidgetService(habRepo HabiticaTaskRepository, tdRepo TodoistTaskRepository) *widgetService {
	return &widgetService{habRepo, tdRepo}
}

func (w *widgetService) GetWidgetResponse() models.WidgetResponse {
	widgetResp := models.WidgetResponse{}
	ch := make(chan error, 4)
	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		habits, err := w.habTaskRepo.GetHabits()
		if err != nil {
			slog.Error("error getting habits")
		}
		ch <- err

		for _, h := range habits {
			if h.Text == "Water" {
				widgetResp.HabiticaWaterValue = h.CounterUp - h.CounterDown
				widgetResp.HabiticaWaterGoal, _ = h.ParseGoal()
			}

			if h.Text == "Reading" {
				widgetResp.HabiticaReadValue = h.CounterUp
				widgetResp.HabiticaReadGoal, _ = h.ParseGoal()
			}
		}
		wg.Done()
	}()

	go func() {
		dailys, err := w.habTaskRepo.GetDailys()
		if err != nil {
			slog.Error("error getting dailys")
		}
		ch <- err

		for _, d := range dailys {
			if d.IsDue {
				widgetResp.HabiticaDailysDue += 1
			}
			if d.IsDue && d.Completed {
				widgetResp.HabiticaDailysDone += 1
			}
		}
		wg.Done()
	}()

	go func() {
		filter := &todoist.TaskFilterOptions{
			Query: "today | od",
			Limit: 200,
		}
		tasks, err := w.tdTaskRepo.GetTasks(filter)
		if err != nil {
			slog.Error("error getting todoist tasks", "err", err)
		}
		ch <- err

		widgetResp.TodoistTasksDue = len(tasks)
		wg.Done()
	}()

	go func() {
		stats, err := w.tdTaskRepo.GetStats()
		if err != nil {
			slog.Error("error getting todoist stats", "err", err)
		}
		ch <- err

		widgetResp.TodoistTasksDone = stats.DaysItems[0].TotalCompleted
		widgetResp.TodoistTasksGoal = stats.Goals.DailyGoal
		wg.Done()
	}()

	wg.Wait()
	close(ch)

	for e := range ch {
		if e != nil {
			widgetResp.Errors = append(widgetResp.Errors, e.Error())
		}
	}

	return widgetResp
}
