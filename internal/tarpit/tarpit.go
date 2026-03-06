package tarpit

import (
	"context"
	"net"
	"time"
)

type Tarpit struct {
	jobs chan Job
	ctx  context.Context
	sem  chan struct{}
}

type Job struct {
	Conn  net.Conn
	Delay time.Duration
}

func NewTarpit(ctx context.Context, maxWorkers int) *Tarpit {
	t := &Tarpit{
		jobs: make(chan Job, 1000),
		ctx:  ctx,
		sem:  make(chan struct{}, maxWorkers),
	}

	go t.run()

	return t
}

func (t *Tarpit) run() {
	for {
		select {
		case <-t.ctx.Done():
			return

		case job := <-t.jobs:
			select {
			case t.sem <- struct{}{}:
				go t.handle(job)

			case <-t.ctx.Done():
				return
			}
		}
	}
}

func (t *Tarpit) handle(job Job) {
	defer func() { <-t.sem }()

	timer := time.NewTimer(job.Delay)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-t.ctx.Done():
	}

	job.Conn.Close()
}

func (t *Tarpit) Add(conn net.Conn, delay time.Duration) {
	select {
	case t.jobs <- Job{Conn: conn, Delay: delay}:
	default:
		// queue full
		conn.Close()
	}
}
