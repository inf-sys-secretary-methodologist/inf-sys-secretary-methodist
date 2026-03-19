package usecases

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/entities"
	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/modules/tasks/domain/repositories"
)

// ========== MockTaskRepository ==========

type MockTaskRepository struct {
	mu             sync.RWMutex
	tasks          map[int64]*entities.Task
	watchers       map[int64][]*entities.TaskWatcher
	attachments    map[int64][]*entities.TaskAttachment
	comments       map[int64]*entities.TaskComment
	checklists     map[int64]*entities.TaskChecklist
	checklistItems map[int64]*entities.TaskChecklistItem
	history        map[int64][]*entities.TaskHistory
	nextID         atomic.Int64
}

func NewMockTaskRepository() *MockTaskRepository {
	m := &MockTaskRepository{
		tasks:          make(map[int64]*entities.Task),
		watchers:       make(map[int64][]*entities.TaskWatcher),
		attachments:    make(map[int64][]*entities.TaskAttachment),
		comments:       make(map[int64]*entities.TaskComment),
		checklists:     make(map[int64]*entities.TaskChecklist),
		checklistItems: make(map[int64]*entities.TaskChecklistItem),
		history:        make(map[int64][]*entities.TaskHistory),
	}
	return m
}

func (m *MockTaskRepository) Create(_ context.Context, task *entities.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	task.ID = m.nextID.Add(1)
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) Save(_ context.Context, task *entities.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tasks[task.ID] = task
	return nil
}

func (m *MockTaskRepository) GetByID(_ context.Context, id int64) (*entities.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if t, ok := m.tasks[id]; ok {
		return t, nil
	}
	return nil, ErrTaskNotFound
}

func (m *MockTaskRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tasks, id)
	return nil
}

func (m *MockTaskRepository) List(_ context.Context, _ repositories.TaskFilter, limit, offset int) ([]*entities.Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.Task
	for _, t := range m.tasks {
		all = append(all, t)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockTaskRepository) Count(_ context.Context, _ repositories.TaskFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.tasks)), nil
}

func (m *MockTaskRepository) GetByProject(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByAuthor(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByAssignee(_ context.Context, _ int64, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetByStatus(_ context.Context, _ domain.TaskStatus, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

func (m *MockTaskRepository) GetOverdueTasks(_ context.Context, _, _ int) ([]*entities.Task, error) {
	return nil, nil
}

// Watchers

func (m *MockTaskRepository) AddWatcher(_ context.Context, watcher *entities.TaskWatcher) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watchers[watcher.TaskID] = append(m.watchers[watcher.TaskID], watcher)
	return nil
}

func (m *MockTaskRepository) RemoveWatcher(_ context.Context, taskID, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	watchers := m.watchers[taskID]
	for i, w := range watchers {
		if w.UserID == userID {
			m.watchers[taskID] = append(watchers[:i], watchers[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *MockTaskRepository) GetWatchers(_ context.Context, taskID int64) ([]*entities.TaskWatcher, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.watchers[taskID], nil
}

func (m *MockTaskRepository) IsWatching(_ context.Context, taskID, userID int64) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, w := range m.watchers[taskID] {
		if w.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

// Attachments

func (m *MockTaskRepository) AddAttachment(_ context.Context, attachment *entities.TaskAttachment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	attachment.ID = m.nextID.Add(1)
	m.attachments[attachment.TaskID] = append(m.attachments[attachment.TaskID], attachment)
	return nil
}

func (m *MockTaskRepository) RemoveAttachment(_ context.Context, attachmentID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for taskID, atts := range m.attachments {
		for i, a := range atts {
			if a.ID == attachmentID {
				m.attachments[taskID] = append(atts[:i], atts[i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (m *MockTaskRepository) GetAttachments(_ context.Context, taskID int64) ([]*entities.TaskAttachment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.attachments[taskID], nil
}

func (m *MockTaskRepository) GetAttachmentByID(_ context.Context, attachmentID int64) (*entities.TaskAttachment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, atts := range m.attachments {
		for _, a := range atts {
			if a.ID == attachmentID {
				return a, nil
			}
		}
	}
	return nil, ErrTaskNotFound
}

// Comments

func (m *MockTaskRepository) AddComment(_ context.Context, comment *entities.TaskComment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	comment.ID = m.nextID.Add(1)
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockTaskRepository) UpdateComment(_ context.Context, comment *entities.TaskComment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.comments[comment.ID] = comment
	return nil
}

func (m *MockTaskRepository) DeleteComment(_ context.Context, commentID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.comments, commentID)
	return nil
}

func (m *MockTaskRepository) GetComments(_ context.Context, taskID int64) ([]*entities.TaskComment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entities.TaskComment
	for _, c := range m.comments {
		if c.TaskID == taskID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *MockTaskRepository) GetCommentByID(_ context.Context, commentID int64) (*entities.TaskComment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if c, ok := m.comments[commentID]; ok {
		return c, nil
	}
	return nil, ErrCommentNotFound
}

// Checklists

func (m *MockTaskRepository) AddChecklist(_ context.Context, checklist *entities.TaskChecklist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	checklist.ID = m.nextID.Add(1)
	m.checklists[checklist.ID] = checklist
	return nil
}

func (m *MockTaskRepository) UpdateChecklist(_ context.Context, checklist *entities.TaskChecklist) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checklists[checklist.ID] = checklist
	return nil
}

func (m *MockTaskRepository) DeleteChecklist(_ context.Context, checklistID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checklists, checklistID)
	return nil
}

func (m *MockTaskRepository) GetChecklists(_ context.Context, taskID int64) ([]*entities.TaskChecklist, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entities.TaskChecklist
	for _, cl := range m.checklists {
		if cl.TaskID == taskID {
			result = append(result, cl)
		}
	}
	return result, nil
}

// Checklist Items

func (m *MockTaskRepository) AddChecklistItem(_ context.Context, item *entities.TaskChecklistItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	item.ID = m.nextID.Add(1)
	m.checklistItems[item.ID] = item
	return nil
}

func (m *MockTaskRepository) UpdateChecklistItem(_ context.Context, item *entities.TaskChecklistItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checklistItems[item.ID] = item
	return nil
}

func (m *MockTaskRepository) DeleteChecklistItem(_ context.Context, itemID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.checklistItems, itemID)
	return nil
}

func (m *MockTaskRepository) GetChecklistItems(_ context.Context, checklistID int64) ([]*entities.TaskChecklistItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entities.TaskChecklistItem
	for _, item := range m.checklistItems {
		if item.ChecklistID == checklistID {
			result = append(result, item)
		}
	}
	return result, nil
}

// History

func (m *MockTaskRepository) AddHistory(_ context.Context, history *entities.TaskHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	history.ID = m.nextID.Add(1)
	m.history[history.TaskID] = append(m.history[history.TaskID], history)
	return nil
}

func (m *MockTaskRepository) GetHistory(_ context.Context, taskID int64, limit, offset int) ([]*entities.TaskHistory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.history[taskID]
	if offset >= len(h) {
		return nil, nil
	}
	end := offset + limit
	if end > len(h) {
		end = len(h)
	}
	return h[offset:end], nil
}

// ========== MockProjectRepository ==========

type MockProjectRepository struct {
	mu       sync.RWMutex
	projects map[int64]*entities.Project
	nextID   atomic.Int64
}

func NewMockProjectRepository() *MockProjectRepository {
	return &MockProjectRepository{
		projects: make(map[int64]*entities.Project),
	}
}

func (m *MockProjectRepository) Create(_ context.Context, project *entities.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	project.ID = m.nextID.Add(1)
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) Save(_ context.Context, project *entities.Project) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[project.ID] = project
	return nil
}

func (m *MockProjectRepository) GetByID(_ context.Context, id int64) (*entities.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if p, ok := m.projects[id]; ok {
		return p, nil
	}
	return nil, ErrProjectNotFound
}

func (m *MockProjectRepository) Delete(_ context.Context, id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, id)
	return nil
}

func (m *MockProjectRepository) List(_ context.Context, _ repositories.ProjectFilter, limit, offset int) ([]*entities.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var all []*entities.Project
	for _, p := range m.projects {
		all = append(all, p)
	}
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (m *MockProjectRepository) Count(_ context.Context, _ repositories.ProjectFilter) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.projects)), nil
}

func (m *MockProjectRepository) GetByOwner(_ context.Context, ownerID int64, limit, offset int) ([]*entities.Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*entities.Project
	for _, p := range m.projects {
		if p.OwnerID == ownerID {
			result = append(result, p)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *MockProjectRepository) GetByStatus(_ context.Context, _ domain.ProjectStatus, _, _ int) ([]*entities.Project, error) {
	return nil, nil
}
