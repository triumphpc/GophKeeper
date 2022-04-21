package pg

import (
	"context"
	"database/sql"
	"github.com/pressly/goose"
	"github.com/triumphpc/GophKeeper/internal/app/service"
	"github.com/triumphpc/GophKeeper/migrations"
	"go.uber.org/zap"
)

// Pg type for postgresql storage
type Pg struct {
	db  *sql.DB
	lgr *zap.Logger
}

// New create postgres storage with not null fields
func New(ctx context.Context, lgr *zap.Logger, DSN string) (*Pg, error) {
	// Database init
	connect, err := sql.Open("postgres", DSN)
	if err != nil {
		return nil, err
	}
	// Ping
	if err := connect.PingContext(ctx); err != nil {
		return nil, err
	}
	// Run migrations
	goose.SetBaseFS(migrations.EmbedMigrations)
	if err := goose.Up(connect, "."); err != nil {
		panic(err)
	}

	return &Pg{connect, lgr}, nil
}

// Close connection
func (s *Pg) Close() {
	err := s.db.Close()
	if err != nil {
		s.lgr.Info("Closing don't close")
	}
}

func (s *Pg) Save(u *service.User) error {
	//TODO implement me
	panic("implement me")
}

func (s *Pg) Find(username string) (*service.User, error) {
	//TODO implement me
	panic("implement me")
}
