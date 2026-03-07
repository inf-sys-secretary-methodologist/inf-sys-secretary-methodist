package agent

import "fmt"

// namedAgents defines the core cast of characters.
var namedAgents = []Agent{
	{
		Name:        "Марина Петровна Соколова",
		Email:       "m.sokolova@uni.local",
		Password:    "AgentPass!2024ms",
		Role:        "academic_secretary",
		Personality: "Методичная, точная, всё по регламенту. Всегда вежлива, но строга к срокам.",
	},
	{
		Name:        "Алексей Николаевич Козлов",
		Email:       "a.kozlov@uni.local",
		Password:    "AgentPass!2024ak",
		Role:        "methodist",
		Personality: "Вдумчивый, внимательный к деталям. Любит структурировать информацию.",
	},
	{
		Name:        "Елена Сергеевна Иванова",
		Email:       "e.ivanova@uni.local",
		Password:    "AgentPass!2024ei",
		Role:        "system_admin",
		Personality: "Решительная, быстрая. Предпочитает действовать, а не обсуждать.",
	},
	{
		Name:        "Дмитрий Иванович Волков",
		Email:       "d.volkov@uni.local",
		Password:    "AgentPass!2024dv",
		Role:        "teacher",
		Personality: "Математик, любит структурность и логику. Чёткие формулировки.",
	},
	{
		Name:        "Ольга Витальевна Михайлова",
		Email:       "o.mikhailova@uni.local",
		Password:    "AgentPass!2024om",
		Role:        "teacher",
		Personality: "Литератор, творческая натура. Пишет красиво и развёрнуто.",
	},
	{
		Name:        "Сергей Павлович Новиков",
		Email:       "s.novikov@uni.local",
		Password:    "AgentPass!2024sn",
		Role:        "teacher",
		Personality: "Физик, аналитический склад ума. Любит факты и цифры.",
	},
	{
		Name:        "Иван Смирнов",
		Email:       "i.smirnov@uni.local",
		Password:    "AgentPass!2024is",
		Role:        "student",
		Personality: "3 курс, активный. Участвует в студенческом совете.",
	},
	{
		Name:        "Анна Кузнецова",
		Email:       "a.kuznetsova@uni.local",
		Password:    "AgentPass!2024akuz",
		Role:        "student",
		Personality: "1 курс, старательная. Всегда сдаёт всё вовремя.",
	},
}

var anonymousFirstNames = []string{
	"Александр", "Михаил", "Максим", "Артём", "Даниил",
	"Никита", "Кирилл", "Егор", "Роман", "Тимофей",
	"Мария", "София", "Дарья", "Виктория", "Полина",
	"Екатерина", "Алиса", "Варвара", "Ксения", "Валерия",
}

var anonymousLastNames = []string{
	"Попов", "Лебедев", "Морозов", "Зайцев", "Павлов",
	"Семёнов", "Голубев", "Виноградов", "Богданов", "Воробьёв",
	"Фёдоров", "Медведев", "Орлов", "Макаров", "Степанов",
	"Андреев", "Ковалёв", "Белов", "Тарасов", "Жуков",
}

// CreateNamedAgents returns all named agents.
func CreateNamedAgents() []*Agent {
	agents := make([]*Agent, len(namedAgents))
	for i := range namedAgents {
		a := namedAgents[i] // copy
		agents[i] = &a
	}
	return agents
}

// CreateAnonymousAgents generates anonymous agents: students and teachers.
func CreateAnonymousAgents(numStudents, numTeachers int) []*Agent {
	var agents []*Agent

	for i := 1; i <= numStudents; i++ {
		nameIdx := (i - 1) % len(anonymousFirstNames)
		lastIdx := (i - 1) % len(anonymousLastNames)
		agents = append(agents, &Agent{
			Name:        fmt.Sprintf("%s %s", anonymousFirstNames[nameIdx], anonymousLastNames[lastIdx]),
			Email:       fmt.Sprintf("student_%02d@uni.local", i),
			Password:    fmt.Sprintf("AgentPass!s%02d", i),
			Role:        "student",
			Personality: "Обычный студент, выполняет задания и следит за расписанием.",
		})
	}

	for i := 1; i <= numTeachers; i++ {
		nameIdx := (numStudents + i - 1) % len(anonymousFirstNames)
		lastIdx := (numStudents + i - 1) % len(anonymousLastNames)
		agents = append(agents, &Agent{
			Name:        fmt.Sprintf("%s %s", anonymousFirstNames[nameIdx], anonymousLastNames[lastIdx]),
			Email:       fmt.Sprintf("teacher_%02d@uni.local", i),
			Password:    fmt.Sprintf("AgentPass!t%02d", i),
			Role:        "teacher",
			Personality: "Преподаватель, ведёт занятия и проверяет работы студентов.",
		})
	}

	return agents
}
