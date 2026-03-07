package agent

// Pool manages a collection of agents.
type Pool struct {
	agents  []*Agent
	byEmail map[string]*Agent
	byRole  map[string][]*Agent
	byName  map[string]*Agent
}

// NewPool creates an empty pool.
func NewPool() *Pool {
	return &Pool{
		byEmail: make(map[string]*Agent),
		byRole:  make(map[string][]*Agent),
		byName:  make(map[string]*Agent),
	}
}

// Add registers an agent in the pool.
func (p *Pool) Add(a *Agent) {
	p.agents = append(p.agents, a)
	p.byEmail[a.Email] = a
	p.byRole[a.Role] = append(p.byRole[a.Role], a)
	p.byName[a.Name] = a
}

// All returns all agents.
func (p *Pool) All() []*Agent {
	return p.agents
}

// ByEmail finds an agent by email.
func (p *Pool) ByEmail(email string) *Agent {
	return p.byEmail[email]
}

// ByName finds an agent by name.
func (p *Pool) ByName(name string) *Agent {
	return p.byName[name]
}

// ByRole returns all agents with the given role.
func (p *Pool) ByRole(role string) []*Agent {
	return p.byRole[role]
}

// FirstByRole returns the first agent with the given role, or nil.
func (p *Pool) FirstByRole(role string) *Agent {
	agents := p.byRole[role]
	if len(agents) == 0 {
		return nil
	}
	return agents[0]
}

// CreateDefault creates a pool with all named agents and the specified
// number of anonymous students/teachers.
func CreateDefault(numStudents, numTeachers int) *Pool {
	pool := NewPool()
	for _, a := range CreateNamedAgents() {
		pool.Add(a)
	}
	for _, a := range CreateAnonymousAgents(numStudents, numTeachers) {
		pool.Add(a)
	}
	return pool
}
