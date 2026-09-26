package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Status string

const (
	InProgressStatus Status = "in_progress"
	WinStatus        Status = "win"
	LoseStatus       Status = "lose"
)

type State string

const (
	MenuState   State = "menu"
	InGameState State = "in_game"
	FinishState State = "finish"
)

type Difficulty string

const (
	EasyDifficulty   Difficulty = "easy"
	MediumDifficulty Difficulty = "medium"
	HardDifficulty   Difficulty = "hard"
)

type Command string

const (
	ExitCommand             Command = "q"
	StartCommand            Command = "s"
	DifficultySwitchCommand Command = "x"
)

type DistanceHint string

const (
	ColdDistanceHint DistanceHint = "cold"
	WarmDistanceHint DistanceHint = "warm"
	HotDistanceHint  DistanceHint = "hot"
)

type TargetDirection string

const (
	HigherTargetDirection TargetDirection = "higher"
	LowerTargetDirection  TargetDirection = "lower"
)

type Game struct {
	running      bool
	target       int
	attemptsLeft int
	state        State
	status       Status
	difficulty   Difficulty
	guesses      []int
	error        string
	startedAt    time.Time
}

type GameResult struct {
	Target     int        `json:"target"`
	Attempts   int        `json:"attempts"`
	Difficulty Difficulty `json:"difficulty"`
	Status     Status     `json:"status"`
	Guesses    []int      `json:"guesses"`
	StartedAt  time.Time  `json:"started_at"`
}

type DifficultyConfig struct {
	min      int
	max      int
	attempts int
	label    string
}

// Colors
const (
	CYELLOW    = "\033[33m"
	CCYAN      = "\033[36m"
	CORANGE    = "\033[38;5;208m"
	CBRIGHTRED = "\033[91m"
	CGREEN     = "\033[32m"
	CRED       = "\033[31m"
	CRESET     = "\033[0m"
)

// Constants

const WarmDistanceHintDistanceThreshold = 15
const HotDistanceHintDistanceThreshold = 5

const ResultsFilename = "results.json"

// Templates
const MenuStateTemplate = string(`
Игра 'Угадай число'
Угадайте число за отведенное количество попыток!

Текущая сложность: %s

Доступные команды:
[` + StartCommand + `] - Начать игру
[` + ExitCommand + `] - Выйти из игры
[` + DifficultySwitchCommand + `] - Смена сложности

%s
Введите команду:
`)

const InGameStateTemplate = `
Число загадано, попробуй угадать

Текущая сложность: %s
Осталось попыток: %d

Предыдущие попытки:
%s
%s
Введите число от %d до %d:
`

const FinishStateTemplate = string(`
Вы %s!

Текущая сложность: %s
Загаданное число было: %d
Осталось попыток: %d

Доступные команды:
[` + StartCommand + `] - Начать игру
[` + ExitCommand + `] - Выйти из игры
[` + DifficultySwitchCommand + `] - Смена сложности

%s
Введите команду:
`)

func generateNumber(min, max int) int {
	return rand.Intn((max+1)-min) + min
}

func getDirection(guess, target int) TargetDirection {
	dist := target - guess

	if dist > 0 {
		return HigherTargetDirection
	}

	return LowerTargetDirection
}

func (d TargetDirection) Format() string {
	if d == HigherTargetDirection {
		return "Секретное число выше ↑"
	}

	return "Секретное число ниже ↓"
}

func getDistance(guess, target int) DistanceHint {
	dist := target - guess

	if dist < 0 {
		dist = -dist
	}

	if dist <= HotDistanceHintDistanceThreshold {
		return HotDistanceHint
	}

	if dist <= WarmDistanceHintDistanceThreshold {
		return WarmDistanceHint
	}

	return ColdDistanceHint
}

func (d DistanceHint) Label() string {
	switch d {
	case ColdDistanceHint:
		return "Холодно"
	case WarmDistanceHint:
		return "Тепло"
	case HotDistanceHint:
		return "Горячо"
	default:
		return "Ошибка"
	}
}

func (d DistanceHint) Color() string {
	switch d {
	case ColdDistanceHint:
		return CCYAN
	case WarmDistanceHint:
		return CORANGE
	case HotDistanceHint:
		return CBRIGHTRED
	default:
		return CRESET
	}
}

func (d DistanceHint) Format() string {
	return fmt.Sprintf("%s%s%s", d.Color(), d.Label(), CRESET)
}

func (d Difficulty) Config() DifficultyConfig {
	switch d {
	case EasyDifficulty:
		return DifficultyConfig{
			label:    "Лёгкая",
			min:      1,
			max:      50,
			attempts: 15,
		}
	case MediumDifficulty:
		return DifficultyConfig{
			label:    "Средняя",
			min:      1,
			max:      100,
			attempts: 10,
		}
	case HardDifficulty:
		return DifficultyConfig{
			label:    "Сложная",
			min:      1,
			max:      200,
			attempts: 5,
		}
	default:
		panic("Неизвестная сложность")
	}
}

func (d Difficulty) Color() string {
	switch d {
	case EasyDifficulty:
		return CGREEN
	case MediumDifficulty:
		return CORANGE
	case HardDifficulty:
		return CRED
	default:
		return CRESET
	}
}

func (d Difficulty) Format() string {
	dConfig := d.Config()
	return fmt.Sprintf("%s%s%s", d.Color(), dConfig.label, CRESET)
}

func (game *Game) Start() {
	dConfig := game.difficulty.Config()

	game.target = generateNumber(dConfig.min, dConfig.max)
	game.attemptsLeft = dConfig.attempts
	game.guesses = make([]int, 0)
	game.status = InProgressStatus
	game.state = InGameState
	game.startedAt = time.Now()
}

func (game *Game) FormatGuesses() string {
	var result = ""

	for _, guess := range game.guesses {
		dist := getDistance(guess, game.target)
		dir := getDirection(guess, game.target)
		result += fmt.Sprintf(
			"%s%d%s | %s%s%s [%s]\n",
			CYELLOW,
			guess,
			CRESET,
			CYELLOW,
			dir.Format(),
			CRESET,
			dist.Format(),
		)
	}

	return result
}

func (game *Game) SwitchDifficulty() {
	var nextDifficulty Difficulty

	switch game.difficulty {
	case EasyDifficulty:
		nextDifficulty = MediumDifficulty
	case MediumDifficulty:
		nextDifficulty = HardDifficulty
	default:
		nextDifficulty = EasyDifficulty
	}

	game.difficulty = nextDifficulty
}

func (game *Game) SaveResult() error {
	dConfig := game.difficulty.Config()
	result := GameResult{
		Target:     game.target,
		Attempts:   dConfig.attempts - game.attemptsLeft,
		Difficulty: game.difficulty,
		Status:     game.status,
		Guesses:    game.guesses,
		StartedAt:  game.startedAt,
	}

	var results []GameResult
	data, err := os.ReadFile(ResultsFilename)

	if err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &results); err != nil {
			return err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	results = append(results, result)

	data, err = json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ResultsFilename, data, 0644)
}

func (game *Game) HandleCommand(cmd Command) {
	game.error = ""

	switch cmd {
	case ExitCommand:
		game.running = false
	case StartCommand:
		game.Start()
	case DifficultySwitchCommand:
		game.SwitchDifficulty()
	default:
		game.error = "Некорректная команда!"
		return
	}
}

func (game *Game) HandleGuess(guess int) {
	game.error = ""

	dConfig := game.difficulty.Config()
	if guess < dConfig.min || guess > dConfig.max {
		game.error = "Некорректное число!"

		return
	}

	game.guesses = append(game.guesses, guess)
	game.attemptsLeft -= 1

	if guess == game.target {
		game.status = WinStatus
		game.state = FinishState
	} else if game.attemptsLeft == 0 {
		game.status = LoseStatus
		game.state = FinishState
	}

	if game.state == FinishState {
		if err := game.SaveResult(); err != nil {
			game.error = "Не удалось сохранить результат!"
		}
	}
}

func (game *Game) Process() {
	switch game.state {
	case MenuState, FinishState:
		var cmd Command
		_, err := fmt.Scanln(&cmd)

		if err != nil {
			game.error = "Некорректная команда!"
			return
		}

		game.HandleCommand(cmd)
	case InGameState:
		var guess int
		_, err := fmt.Scanln(&guess)

		if err != nil {
			game.error = "Введите число!"
			return
		}

		game.HandleGuess(guess)
	}

}

func (game *Game) Clear() {
	fmt.Print("\033[H\033[2J")
}

func (game *Game) Render() {
	dConfig := game.difficulty.Config()
	switch game.state {
	case MenuState:
		fmt.Printf(MenuStateTemplate, game.difficulty.Format(), game.error)
	case InGameState:
		fmt.Printf(
			InGameStateTemplate,
			game.difficulty.Format(),
			game.attemptsLeft,
			game.FormatGuesses(),
			game.error,
			dConfig.min,
			dConfig.max,
		)
	case FinishState:
		var statusLabel string

		if game.status == WinStatus {
			statusLabel = fmt.Sprintf("%sвыиграли%s", CGREEN, CRESET)
		} else {
			statusLabel = fmt.Sprintf("%sпроиграли%s", CRED, CRESET)
		}

		fmt.Printf(FinishStateTemplate, statusLabel, game.difficulty.Format(), game.target, game.attemptsLeft, game.error)
	}
}

func NewGame() *Game {
	return &Game{
		running:    true,
		state:      MenuState,
		difficulty: MediumDifficulty,
	}
}

func main() {

	game := NewGame()

	for game.running {
		game.Clear()
		game.Render()
		game.Process()
	}
}
