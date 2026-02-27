package tarpit

import (
	"context"
	"net"
	"time"
)

type Tarpit struct {
	jobs chan Job
	ctx  context.Context
}

type Job struct {
	Conn  net.Conn
	Delay time.Duration
}

func NewTarpit(ctx context.Context, workers int) *Tarpit {
	t := &Tarpit{
		jobs: make(chan Job, 1000),
		ctx:  ctx,
	}

	// start worker goroutines
	//
	// TODO: maybe flip it around and have one thread that spawns workers up to a limit instead
	for range workers {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case job := <-t.jobs:
					select {
					case <-time.After(job.Delay):
					case <-ctx.Done():
					}
					job.Conn.Close()
				}
			}
		}()
	}

	return t
}

func (t *Tarpit) Add(conn net.Conn, delay time.Duration) {
	select {
	case t.jobs <- Job{Conn: conn, Delay: delay}:
		// added to worker queue
	default:
		// queue full, drop immediately
		conn.Close()
	}
}
