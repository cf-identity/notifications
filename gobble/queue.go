package gobble

import (
	"database/sql"
	"math/rand"
	"time"

	"github.com/coopernurse/gorp"
)

var WaitMaxDuration = 5 * time.Second

type QueueInterface interface {
	Enqueue(Job) (Job, error)
	Reserve(string) <-chan Job
	Dequeue(Job)
	Requeue(Job)
	Len() (int, error)
	RetryQueueLengths() (map[int]int, error)
}

type Queue struct {
	config   Config
	database *DB
	closed   bool
}

func NewQueue(database DatabaseInterface, config Config) *Queue {
	if config.WaitMaxDuration == 0 {
		config.WaitMaxDuration = WaitMaxDuration
	}

	return &Queue{
		database: database.(*DB),
		config:   config,
	}
}

func (queue *Queue) Enqueue(job Job) (Job, error) {
	err := queue.database.Connection.Insert(&job)
	if err != nil {
		return job, err
	}

	return job, nil
}

func (queue *Queue) Requeue(job Job) {
	_, err := queue.database.Connection.Update(&job)
	if err != nil {
		panic(err)
	}
}

func (queue *Queue) Reserve(workerID string) <-chan Job {
	channel := make(chan Job)
	go queue.reserve(channel, workerID)

	return channel
}

func (queue *Queue) Len() (int, error) {
	length, err := queue.database.Connection.SelectInt("SELECT COUNT(*) FROM `jobs`")
	return int(length), err
}

func (queue *Queue) RetryQueueLengths() (map[int]int, error) {
	lengths := map[int]int{}

	type QueueLength struct {
		RetryCount int `db:"retry_count"`
		Count      int `db:"count"`
	}

	records, err := queue.database.Connection.Select(QueueLength{}, "SELECT retry_count, COUNT(*) AS count FROM `jobs` GROUP BY retry_count")
	if err != nil {
		return lengths, err
	}

	for _, value := range records {
		length := value.(*QueueLength)
		lengths[length.RetryCount] = length.Count
	}

	return lengths, nil
}

func (queue *Queue) Close() {
	queue.closed = true
}

func (queue *Queue) reserve(channel chan Job, workerID string) {
	job := Job{}
	for job.ID == 0 {
		var err error

		job = queue.findJob()
		if queue.closed {
			return
		}

		job, err = queue.updateJob(job, workerID)
		if err != nil {
			if _, ok := err.(gorp.OptimisticLockError); ok {
				job = Job{}
				continue
			} else {
				panic(err)
			}
		}
	}

	if queue.closed {
		queue.updateJob(job, "")
		return
	}

	channel <- job
}

func (queue *Queue) Dequeue(job Job) {
	_, err := queue.database.Connection.Delete(&job)
	if err != nil {
		panic(err)
	}
}

func (queue *Queue) findJob() Job {
	job := Job{}
	for job.ID == 0 {
		now := time.Now()
		expired := now.Add(-2 * time.Minute)
		err := queue.database.Connection.SelectOne(&job, "SELECT * FROM `jobs` WHERE ( `worker_id` = \"\" AND `active_at` <= ? ) OR `active_at` <= ? LIMIT 1", now, expired)
		if err != nil {
			if err == sql.ErrNoRows {
				job = Job{}
				queue.waitUpTo(queue.config.WaitMaxDuration)
				continue
			}
			panic(err)
		}
	}
	return job
}

func (queue *Queue) updateJob(job Job, workerID string) (Job, error) {
	job.WorkerID = workerID
	job.ActiveAt = time.Now()
	_, err := queue.database.Connection.Update(&job)
	if err != nil {
		return job, err
	}
	return job, nil
}

func (queue *Queue) waitUpTo(max time.Duration) {
	rand.Seed(time.Now().UnixNano())
	waitTime := rand.Int63n(int64(max))
	<-time.After(time.Duration(waitTime))
}
