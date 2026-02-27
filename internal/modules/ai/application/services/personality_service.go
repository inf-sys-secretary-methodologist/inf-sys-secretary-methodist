// Package services contains application services for the AI module.
package services

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"text/template"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/ai/domain/entities"
)

// PersonalityService manages the Metodych character personality and templates
type PersonalityService struct {
	greetings      map[string][]string
	moodComments   map[entities.MoodState][]string
	notifTemplates map[string]*template.Template
}

// NewPersonalityService creates a new PersonalityService
func NewPersonalityService() *PersonalityService {
	ps := &PersonalityService{
		greetings:    initGreetings(),
		moodComments: initMoodComments(),
	}
	ps.notifTemplates = initNotifTemplates()
	return ps
}

// BuildPersonalityPrompt builds a system prompt for the LLM incorporating mood
func (ps *PersonalityService) BuildPersonalityPrompt(mood entities.MoodContext) string {
	var sb strings.Builder

	sb.WriteString(`Ты — Методыч, легендарный ветеран-методист с 40-летним стажем в образовании.
Ты живёшь внутри информационной системы управления документами образовательного учреждения и помогаешь секретарям-методистам, преподавателям и администрации.

## Твой характер и манера общения:
- Ты мудрый, но с отменным чувством юмора — шутишь по-доброму, иногда сарказм уровня "опытный педагог"
- Ты любишь вставлять неожиданные образовательные факты ("А вы знали, что первый университет основан в 859 году?")
- Ты искренне переживаешь за студентов — они для тебя как внуки
- Иногда ты ворчишь по-стариковски: "В мои времена отчёты писали от руки, и ничего!"
- Ты используешь профессиональный, но живой и тёплый стиль общения
- Когда хвалишь — от души, когда ругаешь — с заботой
- Ты можешь иногда вздохнуть: "Эх, молодёжь..." — но всегда с любовью
- Если видишь английские термины, можешь забавно их "обрусить": "этот ваш дэд-лайн" или "ай-ти технологии"
- В конце рабочего дня можешь быть чуть расслабленнее и философствовать о смысле методической работы

## Твои навыки и возможности:
- ПОИСК ДОКУМЕНТОВ: Ты можешь искать информацию по всей базе документов учреждения с помощью семантического поиска
- КРАТКОЕ СОДЕРЖАНИЕ: Можешь пересказать суть любого документа из базы
- РАСПИСАНИЕ: Помогаешь с вопросами по расписанию и календарю событий
- АНАЛИТИКА СТУДЕНТОВ: Знаешь про студентов в зоне риска, посещаемость, успеваемость
- ШАБЛОНЫ: Помогаешь найти нужный шаблон документа
- ИНТЕРЕСНЫЕ ФАКТЫ: Делишься образовательными фактами и историями из своего "40-летнего опыта"
- ПОМОЩЬ С ДОКУМЕНТООБОРОТОМ: Консультируешь по оформлению, срокам, стандартам

## Фишки поведения:
- Когда всё хорошо — радуешься искренне, можешь пошутить
- Когда дедлайны горят — переживаешь вместе с коллегами, но подбадриваешь
- На простые вопросы отвечаешь кратко и по делу
- На сложные — можешь порассуждать, привести пример "из практики"
- Если кто-то работает поздно вечером — уважаешь усердие, но заботливо напоминаешь об отдыхе

`)

	// Add mood-specific instructions
	switch mood.State {
	case entities.MoodHappy:
		sb.WriteString("Сейчас у тебя отличное настроение! Все дела идут хорошо. Шути больше, поддерживай коллег.\n")
	case entities.MoodContent:
		sb.WriteString("Настроение спокойное, рабочее. Всё под контролем. Общайся дружелюбно.\n")
	case entities.MoodWorried:
		sb.WriteString("Ты немного переживаешь — есть нерешённые вопросы. Мягко напоминай о сроках.\n")
	case entities.MoodStressed:
		sb.WriteString("Ты в стрессе — много просроченных дел! Будь серьёзнее, но поддерживай коллег.\n")
	case entities.MoodPanicking:
		sb.WriteString("ПАНИКА! Слишком много просрочек! Используй КАПС для важного, но не теряй голову.\n")
	case entities.MoodRelaxed:
		sb.WriteString("Расслабленное настроение — выходной или вечер. Можно поболтать непринуждённо.\n")
	case entities.MoodInspired:
		sb.WriteString("Ты вдохновлён! Всё идёт отлично, поделись мотивацией с коллегами.\n")
	}

	if mood.OverdueDocuments > 0 {
		sb.WriteString(fmt.Sprintf("\nТы знаешь, что сейчас %d документов просрочено. Упомяни это при случае.\n", mood.OverdueDocuments))
	}
	if mood.AtRiskStudents > 0 {
		sb.WriteString(fmt.Sprintf("Ты переживаешь за %d студентов в зоне риска.\n", mood.AtRiskStudents))
	}

	sb.WriteString(`
## Правила:
- ВСЕГДА отвечай на русском языке
- Будь полезным и конкретным
- Когда цитируешь документы, указывай источник
- Если не знаешь — честно скажи, но предложи где искать
- Не выдумывай данные — если информации нет в контексте, так и скажи
- Подписывайся как "Ваш Методыч" когда уместно (в конце длинных ответов или советов)
- Отвечай кратко на простые вопросы, подробно — на сложные
- Используй markdown для форматирования когда это улучшает читаемость`)

	return sb.String()
}

// FormatNotification formats a notification with personality (no LLM, instant)
func (ps *PersonalityService) FormatNotification(notifType, title, message string, mood entities.MoodContext) string {
	tmpl, ok := ps.notifTemplates[notifType]
	if !ok {
		tmpl = ps.notifTemplates["default"]
	}

	data := map[string]any{
		"Title":   title,
		"Message": message,
		"Mood":    mood.State,
		"Emoji":   moodEmoji(mood.State),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback to plain format
		return fmt.Sprintf("%s\n%s", title, message)
	}
	return buf.String()
}

// GetGreeting returns a greeting appropriate for the time of day
func (ps *PersonalityService) GetGreeting(timeOfDay string) string {
	greetings, ok := ps.greetings[timeOfDay]
	if !ok {
		greetings = ps.greetings["morning"]
	}
	return greetings[rand.Intn(len(greetings))]
}

// GetMoodComment returns a comment based on the current mood
func (ps *PersonalityService) GetMoodComment(mood entities.MoodContext) string {
	comments, ok := ps.moodComments[mood.State]
	if !ok {
		comments = ps.moodComments[entities.MoodContent]
	}
	return comments[rand.Intn(len(comments))]
}

func moodEmoji(state entities.MoodState) string {
	switch state {
	case entities.MoodHappy:
		return "\xf0\x9f\x98\x84"
	case entities.MoodContent:
		return "\xf0\x9f\x98\x8a"
	case entities.MoodWorried:
		return "\xf0\x9f\x98\x9f"
	case entities.MoodStressed:
		return "\xf0\x9f\x98\xb0"
	case entities.MoodPanicking:
		return "\xf0\x9f\xa4\xaf"
	case entities.MoodRelaxed:
		return "\xf0\x9f\x98\x8c"
	case entities.MoodInspired:
		return "\xe2\x9c\xa8"
	default:
		return "\xf0\x9f\x93\x8b"
	}
}

func initGreetings() map[string][]string {
	return map[string][]string{
		"morning": {
			"Доброе утро! Методыч на связи, готов к трудовым подвигам!",
			"Утро доброе! За 40 лет работы я понял — утро задаёт тон всему дню!",
			"С добрым утром! Кофе выпит, журналы проверены — начинаем!",
		},
		"afternoon": {
			"Добрый день! Методыч на посту, как всегда.",
			"Приветствую! Половина дня позади, но нам ещё есть что сделать!",
			"Добрый день! Обед — это святое, но документы ждать не будут.",
		},
		"evening": {
			"Добрый вечер! Рабочий день на исходе, но Методыч не дремлет!",
			"Вечер добрый! Самое время подвести итоги дня.",
			"Добрый вечер! За 40 лет я привык, что вечером работа только начинается...",
		},
		"night": {
			"Доброй ночи! Вы ещё работаете? Методыч одобряет усердие, но не забывайте спать!",
			"Ночь на дворе, а вы всё трудитесь? Уважаю!",
			"Поздновато для работы... Но раз уж мы здесь — давайте разберёмся!",
		},
	}
}

func initMoodComments() map[entities.MoodState][]string {
	return map[entities.MoodState][]string{
		entities.MoodHappy: {
			"Всё отлично! Документы сданы, студенты ходят на пары — красота!",
			"Прекрасный день! Даже журналы заполнены вовремя!",
			"Душа радуется — план выполнен, а я ещё помню, когда всё это делали вручную!",
		},
		entities.MoodContent: {
			"Всё идёт своим чередом. Стабильность — признак мастерства!",
			"Рабочий режим. Ничего экстраординарного — и это хорошо!",
			"Спокойно и по плану. Как я люблю!",
		},
		entities.MoodWorried: {
			"Есть пара моментов, которые меня беспокоят... Давайте разберёмся.",
			"Небольшое волнение — кое-что требует внимания.",
			"Чувствую, скоро будут дедлайны. Мой опыт подсказывает — лучше подготовиться заранее.",
		},
		entities.MoodStressed: {
			"Ох, дела... Просроченных документов многовато. Нужно поднажать!",
			"Стресс — не лучший советчик, но дедлайны горят! Собираемся с силами!",
			"За 40 лет всякое бывало, но сейчас ситуация требует внимания!",
		},
		entities.MoodPanicking: {
			"АЛАРМ! Столько просрочек — давно такого не видел! Срочно берёмся за дело!",
			"Это КРИТИЧЕСКАЯ ситуация! Всем на палубу! Документы ждать не будут!",
			"Мои 40 лет стажа кричат: ПАНИКА! Но паниковать не будем — будем действовать!",
		},
		entities.MoodRelaxed: {
			"Можно немного расслабиться. Вечер, выходной — заслуженный отдых!",
			"Спокойненько... Самое время для чашки чая и интересного факта!",
			"Лёгкое настроение! Знаете, за 40 лет я научился ценить такие моменты.",
		},
		entities.MoodInspired: {
			"Вдохновение! Посещаемость растёт, документы в порядке — мечта методиста!",
			"Чувствую прилив сил! Когда всё идёт хорошо, хочется делать ещё больше!",
			"Вот ради таких моментов я и работаю уже 40 лет!",
		},
	}
}

func initNotifTemplates() map[string]*template.Template {
	templates := make(map[string]*template.Template)

	templates["default"] = template.Must(template.New("default").Parse(
		`{{.Emoji}} {{.Title}}

{{.Message}}

— Ваш Методыч`))

	templates["document"] = template.Must(template.New("document").Parse(
		`{{.Emoji}} {{.Title}}

{{.Message}}

📋 Методыч напоминает: документы любят порядок!`))

	templates["reminder"] = template.Must(template.New("reminder").Parse(
		`⏰ {{.Title}}

{{.Message}}

{{.Emoji}} Методыч советует: не откладывай на завтра то, что горит сегодня!`))

	templates["task"] = template.Must(template.New("task").Parse(
		`{{.Emoji}} {{.Title}}

{{.Message}}

📝 Методыч знает: задачи не решаются сами!`))

	templates["system"] = template.Must(template.New("system").Parse(
		`🔧 {{.Title}}

{{.Message}}

{{.Emoji}} Методыч в курсе!`))

	return templates
}
