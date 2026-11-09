package agent

import (
	"errors"
	"time"

	bolt "go.etcd.io/bbolt"
)

func (h historyStore) markTaskRunning(task queuedTask) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		row, found, err := queueRecordFrom(tx, task.result.TaskID)
		if err != nil {
			return err
		}
		if !found {
			return errors.New("任务尚未持久接收")
		}
		row.Status = "running"
		if err := saveQueueRecord(tx, row); err != nil {
			return err
		}
		session, err := sessionFrom(tx, row.SessionID)
		if err != nil {
			return err
		}
		if session.TaskID != row.ID || session.TaskStartedAt == 0 {
			session.TaskStartedAt = time.Now().UnixMilli()
		}
		session.TaskID, session.TaskStatus, session.TaskError = row.ID, "running", ""
		session.TaskAcceptedAt, session.TaskCompletedAt = row.CreatedAt, 0
		return saveSession(tx, session)
	})
}

func (h historyStore) finishQueuedTask(task queuedTask, status string, runErr error) error {
	return h.transact(true, func(tx *bolt.Tx) error {
		row, found, err := queueRecordFrom(tx, task.result.TaskID)
		if err != nil {
			return err
		}
		if found {
			row.Status = status
			if err := saveQueueRecord(tx, row); err != nil {
				return err
			}
		}
		session, err := sessionFrom(tx, task.result.Session.ID)
		if err != nil || session.TaskID != task.result.TaskID {
			return err
		}
		session.TaskStatus, session.TaskError = status, ""
		session.TaskCompletedAt = time.Now().UnixMilli()
		if runErr != nil {
			session.TaskError = runErr.Error()
		}
		return saveSession(tx, session)
	})
}
