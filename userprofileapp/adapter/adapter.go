package adapter

type RPC struct {
	contributorRPC     ContributorRPC
	tasksRPC           TasksRPC
	contributorStatRPC ContributorStatRPC
}

func NewAdapter() RPC {
	return RPC{
		contributorRPC:     NewContributorRPC(),
		tasksRPC:           NewTasksRPC(),
		contributorStatRPC: NewContributorStatRPC(),
	}
}
