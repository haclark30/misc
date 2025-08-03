package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"misc/clients/habitica"
	"misc/clients/todoist"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/google/go-querystring/query"
	_ "github.com/joho/godotenv/autoload"
)

const (
	KINDLEWIDTH   = 61
	KINDLEHEIGHT  = 26
	Strikethrough = "\033[9m"
	Reset         = "\033[0m"
	MAXTODOS      = 5
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		f, _ := tea.LogToFile("test.log", "")
		defer f.Close()
		m := newModel(10, 10, lipgloss.DefaultRenderer())
		m, err := m.updateState()
		if err != nil {
			slog.Error("error updating state", "err", err)
		}
		prog := tea.NewProgram(m, tea.WithAltScreen())
		prog.Run()
		os.Exit(0)
	}
	s, err := wish.NewServer(
		wish.WithAddress(":23234"),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server")
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

// You can wire any Bubble Tea model up to the middleware with a function that
// handles the incoming ssh.Session. Here we just grab the terminal info and
// pass it to the new model. You can also return tea.ProgramOptions (such as
// tea.WithAltScreen) on a session by session basis.
func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	pty, _, _ := s.Pty()

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	renderer := bubbletea.MakeRenderer(s)
	m := newModel(pty.Window.Width, pty.Window.Height, renderer)
	m, err := m.updateState()
	if err != nil {
		slog.Error("error updating state", "err", err)
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type model struct {
	fitbitClient  *http.Client
	habClient     habitica.HabiticaClient
	todoistClient *todoist.TodoistRestClient
	width         int
	height        int
	txtStyle      lipgloss.Style
	quitStyle     lipgloss.Style
	dailys        []habitica.Daily
	habs          []habitica.Habit
	chores        []todoist.Task
	hygiene       []todoist.Task
	activity      ActivityResponse
	err           error
	content       string
	viewport      viewport.Model
	ready         bool
}

func newModel(width, height int, renderer *lipgloss.Renderer) model {
	fitbitClient := createFitbitClient()
	habClient := habitica.NewHabiticaClient(
		os.Getenv("HABITICA_API_USER"),
		os.Getenv("HABITICA_API_KEY"),
	)
	todoistClient := todoist.NewClient(os.Getenv("TODOIST_API_KEY"))
	txtStyle := renderer.NewStyle().Foreground(lipgloss.Color("31"))
	quitStyle := renderer.NewStyle().Foreground(lipgloss.Color("8"))

	m := model{
		fitbitClient:  fitbitClient,
		habClient:     habClient,
		todoistClient: todoistClient,
		width:         width,
		height:        height,
		txtStyle:      txtStyle,
		quitStyle:     quitStyle,
	}
	return m
}

type tickMsg struct{}

func (m model) updateState() (model, error) {
	m.fitbitClient = createFitbitClient()
	dailys, err := m.habClient.GetDailys()
	if err != nil {
		log.Error("error getting dailys", "err", err)
		return m, fmt.Errorf("error updating dailys: %w", err)
	}
	m.dailys = dailys

	habs, err := m.habClient.GetHabits()
	if err != nil {
		log.Error("error getting habits", "err", err)
		return m, fmt.Errorf("error updating habits: %w", err)
	}
	m.habs = habs

	chores, err := m.updateChores()
	if err != nil {
		slog.Error("error updating chores", "err", err)
		return m, fmt.Errorf("error updating chores: %w", err)
	}
	m.chores = chores

	hygiene, err := m.updateHygiene()
	if err != nil {
		slog.Error("error updating hygiene", "err", err)
		return m, fmt.Errorf("error updating hygiene: %w", err)
	}
	m.hygiene = hygiene

	activity, err := getFitbitActivity(m.fitbitClient)
	if err != nil {
		slog.Error("error getting fitbit", "err", err)
		return m, fmt.Errorf("error getting fitbit: %w", err)
	}
	m.activity = activity
	m.err = nil
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return tickMsg{} })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.ready = true
		} else {
			m.viewport.Height = msg.Height
			m.viewport.Width = msg.Width
		}
		slog.Info("window update", "height", m.height, "width", m.width)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			var err error
			m, err = m.updateState()
			if err != nil {
				slog.Error("error updating state", "err", err)
				m.err = fmt.Errorf("error updating state: %w", err)
				return m, nil
			}
			m.err = nil
		}
	case tickMsg:
		var err error
		m, err = m.updateState()
		if err != nil {
			slog.Error("error updating state", "err", err)
			m.err = fmt.Errorf("error updating state: %w", err)
			return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return tickMsg{} })
		}
		m.err = nil
		return m, tea.Tick(30*time.Second, func(t time.Time) tea.Msg { return tickMsg{} })
	}
	m.content = m.updateContent()
	m.viewport.SetContent(m.content)
	m.viewport, cmd = m.viewport.Update(msg)
	return m, tea.Batch(cmd)
}

func (m model) updateHabitica() ([]habitica.Habit, []habitica.Daily) {
	dailys, err := m.habClient.GetDailys()
	if err != nil {
		log.Error("error getting dailys", "err", err)
	}
	habs, err := m.habClient.GetHabits()
	if err != nil {
		log.Error("error getting habits", "err", err)
	}
	return habs, dailys
}

func (m model) updateChores() ([]todoist.Task, error) {
	filter := todoist.TaskFilterOptions{
		Filter: "##shared chores & (today | od)",
	}
	req, err := m.todoistClient.NewTodoistRequest(http.MethodGet, "tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create todoist tasks request: %w", err)
	}
	v, err := query.Values(filter)
	if err != nil {
		return nil, fmt.Errorf("unable to encode filter: %w", err)
	}
	req.URL.RawQuery = v.Encode()

	resp, err := m.todoistClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling todoist: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("error calling todoist", "code", resp.StatusCode)
	}

	var todoResp []todoist.Task
	err = json.NewDecoder(resp.Body).Decode(&todoResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding todoist task resp: %w", err)
	}
	sortTodoistTasks(todoResp)
	return todoResp, nil
}

func sortTodoistTasks(todoResp []todoist.Task) {
	sort.Slice(todoResp, func(i, j int) bool {
		var iDate, jDate time.Time
		if todoResp[i].Due.Datetime != nil {
			iDate, _ = time.Parse("2006-01-02T15:04:05", *todoResp[i].Due.Datetime)
		} else {
			iDate, _ = time.Parse("2006-01-02", todoResp[i].Due.Date)
		}
		if todoResp[j].Due.Datetime != nil {
			jDate, _ = time.Parse("2006-01-02T15:04:05", *todoResp[j].Due.Datetime)
		} else {
			jDate, _ = time.Parse("2006-01-02", todoResp[j].Due.Date)
		}
		return iDate.Before(jDate)
	})
}

func (m model) updateHygiene() ([]todoist.Task, error) {
	filter := &todoist.TaskFilterOptions{
		Filter: "##health and hygiene & (today | od)",
	}
	req, err := m.todoistClient.NewTodoistRequest(http.MethodGet, "tasks", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create todoist tasks request: %w", err)
	}
	v, err := query.Values(filter)
	if err != nil {
		return nil, fmt.Errorf("unable to encode filter: %w", err)
	}
	req.URL.RawQuery = v.Encode()

	resp, err := m.todoistClient.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling todoist: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("error calling todoist", "code", resp.StatusCode)
	}

	var todoResp []todoist.Task
	err = json.NewDecoder(resp.Body).Decode(&todoResp)
	if err != nil {
		return nil, fmt.Errorf("error decoding todoist task resp: %w", err)
	}
	sortTodoistTasks(todoResp)
	return todoResp, nil
}

func (m model) updateContent() string {
	if m.err != nil {
		return "error updatating state"
	}
	title := "Kindle Dash"
	title = m.txtStyle.Align(lipgloss.Center, lipgloss.Center).Render(title)

	stepsStr := fmt.Sprintf(
		"%d / %d steps",
		m.activity.Summary.Steps,
		m.activity.Goals.Steps,
	)

	minStr := fmt.Sprintf(
		"%d / %d active minutes",
		m.activity.Summary.VeryActiveMinutes,
		m.activity.Goals.ActiveMinutes,
	)

	fitbitStr := lipgloss.JoinVertical(
		lipgloss.Center,
		stepsStr,
		minStr,
	)

	dailyStr := ""
	dailyRows := make([][]string, 0)
	for _, d := range m.dailys {
		if d.IsDue && d.Completed {
			dailyStr = Strikethrough + d.Text + Reset
			dailyRows = append(dailyRows, []string{dailyStr, fmt.Sprintf("%d", d.Streak)})
		} else if d.IsDue {
			dailyStr = fmt.Sprintf("%s", d.Text)
			dailyRows = append(dailyRows, []string{dailyStr, fmt.Sprintf("%d", d.Streak)})
		}
	}

	dailyTable := table.New().Border(lipgloss.HiddenBorder()).Rows(dailyRows...).Render()
	dailyTable = lipgloss.NewStyle().MarginLeft(5).Render(dailyTable)

	habRows := make([][]string, 0)
	for _, h := range m.habs {
		habRows = append(habRows, []string{h.Text, fmt.Sprintf("%d", h.CounterUp)})
	}
	habitTable := table.New().Border(lipgloss.HiddenBorder()).Rows(habRows...).Render()

	choreRows := make([][]string, 0)
	for _, t := range m.chores {
		choreRows = append(choreRows, []string{t.Content})
	}
	choresTable := table.New().Border(lipgloss.HiddenBorder()).Rows(choreRows...).Render()

	hygieneRows := make([][]string, 0)
	for _, t := range m.hygiene {
		hygieneRows = append(hygieneRows, []string{t.Content})
	}
	hygieneTable := table.New().Border(lipgloss.HiddenBorder()).Rows(hygieneRows...).Render()
	hygieneTable = lipgloss.NewStyle().MarginLeft(5).Render(hygieneTable)

	pad := m.width / 2
	slog.Info("set margin", "pad", pad)
	style := lipgloss.NewStyle().Align(lipgloss.Center).Border(lipgloss.NormalBorder())
	style = style.SetString(
		lipgloss.JoinVertical(
			lipgloss.Center,
			title,
			fitbitStr,
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					habitTable,
					dailyTable,
				),
				lipgloss.JoinHorizontal(
					lipgloss.Top,
					choresTable,
					hygieneTable,
				),
			),
		),
	)

	ret := style.String()
	lines := strings.Split(ret, "\n")
	borderLen := utf8.RuneCountInString(lines[0])
	slog.Info("got borderLen", "borderLen", borderLen, "borderStr", strconv.Quote(lines[0]))
	return ret
}

func (m model) View() string {
	return m.viewport.View()
}
