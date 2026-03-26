package indexer

import (
	"context"
	"fmt"
	"strings"

	"github.com/kurt4ins/taskmanager/internal/repo"
)

type IndexJob struct {
	TaskId      int
	Title       string
	Description string
}

type Indexer struct {
	queue chan IndexJob
	repo  *repo.TaskRepo
}

func New(r *repo.TaskRepo, queueSize int, workers int) *Indexer {
	idx := &Indexer{
		queue: make(chan IndexJob, queueSize),
		repo:  r,
	}

	for i := range workers {
		go idx.worker(i)
	}
	return idx
}

func (idx *Indexer) Submit(job IndexJob) bool {
	select {
	case idx.queue <- job:
		return true
	default:
		return false
	}
}

func (idx *Indexer) worker(id int) {
	for job := range idx.queue {
		text := fmt.Sprintf("%s %s", job.Title, job.Description)
		count := len(strings.Fields(text))

		if err := idx.repo.UpdateWordCount(context.Background(), job.TaskId, count); err != nil {
			fmt.Printf("indexer worker %d: %v", id, err)
		}
	}
}
