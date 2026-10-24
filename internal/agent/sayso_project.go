package agent

import (
	"errors"
	"github.com/string1225/remote-agent-cli/internal/protocol"
	"github.com/string1225/remote-agent-cli/internal/store"
)

// Binding an existing SaySo project retains all discussion IDs and recordings.
// New client projects get one pinned discussion, rather than a second project tree.
func (s *saysoStore) bindProject(serviceID string) (*soProject, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var service protocol.Service
	for _, v := range s.a.services() {
		if v.ID == serviceID {
			service = v
			break
		}
	}
	if service.ID == "" {
		return nil, errors.New("项目不存在")
	}
	var project *soProject
	for _, v := range s.state.Projects {
		if v.ServiceID == serviceID {
			project = v
			break
		}
	}
	if project == nil {
		for _, v := range s.state.Projects {
			if v.ServiceID == "" && v.RootPath == service.Workspace {
				project = v
				break
			}
		}
	}
	if project == nil {
		project = &soProject{ID: store.Token(), Color: "#557461", CreatedAt: soNow()}
		s.state.Projects = append(s.state.Projects, project)
	}
	project.ServiceID = service.ID
	project.Name = service.Name
	project.RootPath = service.Workspace
	if s.session(project.PinnedSessionID) == nil {
		for _, v := range s.state.Sessions {
			if v.ProjectID == project.ID {
				project.PinnedSessionID = v.ID
				break
			}
		}
		if s.session(project.PinnedSessionID) == nil {
			v := &soSession{ID: store.Token(), ProjectID: project.ID, Title: "SaySo · " + service.Name, Status: "paused", StartedAt: soNow(), Recordings: []soRecording{}, Transcript: []soUtterance{}, AnalysisStatus: "idle", AnalysisTraces: []*soTrace{}}
			s.state.Sessions = append([]*soSession{v}, s.state.Sessions...)
			project.PinnedSessionID = v.ID
		}
	}
	if err := s.save(); err != nil {
		return nil, err
	}
	v := *project
	return &v, nil
}
