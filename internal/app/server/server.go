package server

import "context"

type server struct {
	ctx context.Context
	c   context.CancelFunc
}

func New(adr string) *server {
	ctx, cancel := context.WithCancel(context.Background())

	return &server{
		ctx: ctx,
		c:   cancel,
	}
}

func (s *server) Run() error {
	const fn = "server.Run"

	for {
		select {
		case <-s.ctx.Done():
			return nil
		}
	}
}

func (s *server) Close() error {
	const fn = "server.Close"

	s.c()

	return nil
}
