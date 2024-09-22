package worker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	workers       int
	jobs          chan Job
	results       chan JobRs
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	cancelOnError bool
	jobsInFlight  int32
	done          chan struct{}
}

func NewWorkerPool(workers int, cancelOnError bool) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:       workers,
		jobs:          make(chan Job),
		results:       make(chan JobRs),
		ctx:           ctx,
		wg:            sync.WaitGroup{},
		cancel:        cancel,
		cancelOnError: cancelOnError,
		jobsInFlight:  0,
		done:          make(chan struct{}),
	}
}

func (p *WorkerPool) Start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
	go p.monitorAndShutdown()
}

func (p *WorkerPool) initiateShutdown() {
	p.cancel()
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
	close(p.done)
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			atomic.AddInt32(&p.jobsInFlight, 1)
			result := job.Execute()
			atomic.AddInt32(&p.jobsInFlight, -1)
			select {
			case p.results <- result:
				if p.cancelOnError && result.Err != nil {
					p.cancel()
					return
				}
			case <-p.ctx.Done():
				return
			}
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *WorkerPool) monitorAndShutdown() {
	for {
		select {
		case <-p.ctx.Done():
			p.initiateShutdown()
			return
		default:
			if atomic.LoadInt32(&p.jobsInFlight) == 0 && len(p.jobs) == 0 {
				time.Sleep(100 * time.Millisecond) // Give a short delay to ensure no new jobs are incoming
				if atomic.LoadInt32(&p.jobsInFlight) == 0 && len(p.jobs) == 0 {
					p.initiateShutdown()
					return
				}
			}
			time.Sleep(4 * time.Second) // Check periodically
		}
	}
}

func (p *WorkerPool) AddJobNonBlocking(job Job) error {
	select {
	case p.jobs <- job:
		return nil
	case <-p.ctx.Done():
		return errors.New("worker pool is shutting down")
	default:
		// If the job channel is full, wait a bit and try again
		time.Sleep(10 * time.Millisecond)
		select {
		case p.jobs <- job:
			return nil
		case <-p.ctx.Done():
			return errors.New("worker pool is shutting down")
		default:
			return errors.New("job queue is full")
		}
	}
}

func (p *WorkerPool) Results() <-chan JobRs {
	return p.results
}

func (p *WorkerPool) Wait() {
	<-p.done
}
