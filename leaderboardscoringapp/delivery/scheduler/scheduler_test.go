package scheduler

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/project"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

type projectServiceMock struct {
	responses map[int32]*project.ListProjectsResponse
	requests  []*project.ListProjectsRequest
	err       error
}

func (m *projectServiceMock) ListProjects(ctx context.Context, req *project.ListProjectsRequest) (*project.ListProjectsResponse, error) {
	m.requests = append(m.requests, req)

	if m.err != nil {
		return nil, m.err
	}

	response, ok := m.responses[req.Offset]
	if !ok {
		return nil, fmt.Errorf("unexpected offset: %d", req.Offset)
	}

	return response, nil
}

type leaderboardServiceMock struct {
	calls [][]string
	err   error
}

func (m *leaderboardServiceMock) LeaderboardSnapshot(ctx context.Context, projectIDs []string) error {
	m.calls = append(m.calls, append([]string(nil), projectIDs...))
	return m.err
}

func TestMain(m *testing.M) {
	err := logger.Init(logger.Config{
		Level:            "error",
		FilePath:         "test.log",
		FileMaxSizeInMB:  1,
		FileMaxAgeInDays: 1,
	})
	if err != nil {
		panic(err)
	}

	code := m.Run()

	_ = logger.Close()

	os.Exit(code)
}

func TestScheduler_snapshotLeaderboardTask_ZeroProjects(t *testing.T) {
	projectSvc := &projectServiceMock{
		responses: map[int32]*project.ListProjectsResponse{
			0: {
				Projects:   nil,
				TotalCount: 0,
			},
		},
	}

	leaderboardSvc := &leaderboardServiceMock{}

	s := Scheduler{
		project:        projectSvc,
		leaderboardSvc: leaderboardSvc,
		cfg: Config{
			PageSizeListProject:       100,
			SnapshotJobContextTimeout: time.Minute,
		},
	}

	s.snapshotLeaderboardTask(context.Background())

	require.Len(t, leaderboardSvc.calls, 1)
	require.Empty(t, leaderboardSvc.calls[0])

	require.Len(t, projectSvc.requests, 1)
	require.Equal(t, int32(100), projectSvc.requests[0].PageSize)
	require.Equal(t, int32(0), projectSvc.requests[0].Offset)
	require.Equal(t, project.StatusActive, projectSvc.requests[0].Status)
}

func TestScheduler_snapshotLeaderboardTask_OneProject(t *testing.T) {
	projectSvc := &projectServiceMock{
		responses: map[int32]*project.ListProjectsResponse{
			0: {
				Projects: []project.ProjectItem{
					{ProjectID: "project-1"},
				},
				TotalCount: 1,
			},
		},
	}

	leaderboardSvc := &leaderboardServiceMock{}

	s := Scheduler{
		project:        projectSvc,
		leaderboardSvc: leaderboardSvc,
		cfg: Config{
			PageSizeListProject:       100,
			SnapshotJobContextTimeout: time.Minute,
		},
	}

	s.snapshotLeaderboardTask(context.Background())

	require.Len(t, leaderboardSvc.calls, 1)
	require.Equal(t, []string{"project-1"}, leaderboardSvc.calls[0])

	require.Len(t, projectSvc.requests, 1)
	require.Equal(t, int32(100), projectSvc.requests[0].PageSize)
	require.Equal(t, int32(0), projectSvc.requests[0].Offset)
	require.Equal(t, project.StatusActive, projectSvc.requests[0].Status)
}

func TestScheduler_snapshotLeaderboardTask_MultipleProjects(t *testing.T) {
	projectSvc := &projectServiceMock{
		responses: map[int32]*project.ListProjectsResponse{
			0: {
				Projects: []project.ProjectItem{
					{ProjectID: "project-1"},
					{ProjectID: "project-2"},
					{ProjectID: "project-3"},
				},
				TotalCount: 5,
			},
			3: {
				Projects: []project.ProjectItem{
					{ProjectID: "project-4"},
					{ProjectID: "project-5"},
				},
				TotalCount: 5,
			},
		},
	}

	leaderboardSvc := &leaderboardServiceMock{}

	s := Scheduler{
		project:        projectSvc,
		leaderboardSvc: leaderboardSvc,
		cfg: Config{
			PageSizeListProject:       3,
			SnapshotJobContextTimeout: time.Minute,
		},
	}

	s.snapshotLeaderboardTask(context.Background())

	require.Len(t, leaderboardSvc.calls, 1)
	require.Equal(t, []string{
		"project-1",
		"project-2",
		"project-3",
		"project-4",
		"project-5",
	}, leaderboardSvc.calls[0])

	require.Len(t, projectSvc.requests, 2)

	require.Equal(t, int32(3), projectSvc.requests[0].PageSize)
	require.Equal(t, int32(0), projectSvc.requests[0].Offset)
	require.Equal(t, project.StatusActive, projectSvc.requests[0].Status)

	require.Equal(t, int32(3), projectSvc.requests[1].PageSize)
	require.Equal(t, int32(3), projectSvc.requests[1].Offset)
	require.Equal(t, project.StatusActive, projectSvc.requests[1].Status)
}
