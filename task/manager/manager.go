package manager

import (
	"mesos-framework-sdk/include/mesos_v1"
	"mesos-framework-sdk/task"
	"time"
	"mesos-framework-sdk/task/retry"
)

// Consts for mesos states.
const (
	RUNNING          = mesos_v1.TaskState_TASK_RUNNING
	KILLED           = mesos_v1.TaskState_TASK_KILLED
	LOST             = mesos_v1.TaskState_TASK_LOST
	GONE             = mesos_v1.TaskState_TASK_GONE
	STAGING          = mesos_v1.TaskState_TASK_STAGING
	STARTING         = mesos_v1.TaskState_TASK_STARTING // Default executor never sends this, it sends RUNNING directly.
	UNKNOWN          = mesos_v1.TaskState_TASK_UNKNOWN
	UNREACHABLE      = mesos_v1.TaskState_TASK_UNREACHABLE
	FINISHED         = mesos_v1.TaskState_TASK_FINISHED
	DROPPED          = mesos_v1.TaskState_TASK_DROPPED
	FAILED           = mesos_v1.TaskState_TASK_FAILED
	ERROR            = mesos_v1.TaskState_TASK_ERROR
	GONE_BY_OPERATOR = mesos_v1.TaskState_TASK_GONE_BY_OPERATOR
	KILLING          = mesos_v1.TaskState_TASK_KILLING
)

// Task manager holds information about tasks coming into the framework from the API
// It can set the state of a task.  How the implementation holds/handles those tasks
// is up to the end user.
type TaskManager interface {
	Add(...*Task) error
	Delete(...*Task) error
	Get(*string) (*Task, error)
	GetById(id *mesos_v1.TaskID) (*Task, error)
	HasTask(*mesos_v1.TaskInfo) bool
	Update(...*Task) error
	AllByState(state mesos_v1.TaskState) ([]*Task, error)
	TotalTasks() int
	All() ([]Task, error)
}

// Used to hold information about task states in the task manager.
// Task and its fields should be public so that we can encode/decode this.
type Task struct {
	Info      *mesos_v1.TaskInfo
	State     mesos_v1.TaskState
	Filters   []task.Filter
	Retry     *retry.TaskRetry
	Instances int
	Retries   int
	GroupInfo GroupInfo
}

type GroupInfo struct {
	GroupName string
	InGroup   bool
}

// TODO (tim): Create a serialize/deserialize mechanism from string <-> struct to avoid costly encoding.

func (t *Task) Reschedule() {
	t.Retry.TotalRetries += 1 // Increment retry counter.

	// Minimum is 1 seconds, max is 60.
	if t.Retry.RetryTime < 1*time.Second {
		t.Retry.RetryTime = 1 * time.Second
	} else if t.Retry.RetryTime > time.Minute {
		t.Retry.RetryTime = time.Minute
	}

	delay := t.Retry.RetryTime + t.Retry.RetryTime

	// Total backoff can't be greater than 5 minutes.
	if delay > 5*time.Minute {
		delay = 5 * time.Minute
	}

	t.Retry.RetryTime = delay // update with new time.
	t.State = mesos_v1.TaskState_TASK_UNKNOWN
}
