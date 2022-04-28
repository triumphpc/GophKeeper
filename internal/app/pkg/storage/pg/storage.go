package pg

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/triumphpc/GophKeeper/internal/app/pkg/storage"
	"github.com/triumphpc/GophKeeper/internal/app/service/user"
	ud "github.com/triumphpc/GophKeeper/internal/app/service/userdata"
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

// CreateUser implement save user to storage
func (s *Pg) CreateUser(u *user.User) error {
	err := s.db.QueryRow(
		"INSERT INTO users (id, login, password, role) VALUES (DEFAULT, $1, $2, $3) RETURNING id",
		u.Username,
		u.HashedPassword,
	).Scan(&u.Id)

	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			if err.Code == pgerrcode.UniqueViolation {
				return errors.New("login already exist")
			}
			return err
		}
	}

	return nil
}

// Find user by login  in storage
func (s *Pg) Find(login string) (*user.User, error) {
	var usr user.User
	err := s.db.QueryRow(
		"SELECT id, login, password, role FROM users WHERE login=$1", login).
		Scan(&usr.Id, &usr.Username, &usr.HashedPassword, &usr.Role)
	if err != nil {
		return nil, err
	}

	return &usr, nil
}

// SaveText save text data to store
func (s *Pg) SaveText(ctx context.Context, data *ud.DataText, userId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastId int
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO user_data_text (id, text, meta, name) VALUES (DEFAULT, $1, $2, $3) RETURNING id",
		data.Text,
		data.Meta,
		data.Name,
	).Scan(&lastId)

	if err != nil {
		return err
	}

	data.Id = lastId

	// Create relation
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO user_data (id, type_id, entity_id, user_id) VALUES (DEFAULT, $1, $2, $3)",
		ud.TypeText,
		lastId,
		userId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveCard save card data to store
func (s *Pg) SaveCard(ctx context.Context, data *ud.DataCard, userId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastId int
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO user_data_cards (id, number, meta) VALUES (DEFAULT, $1, $2) RETURNING id",
		data.Number,
		data.Meta,
	).Scan(&lastId)

	if err != nil {
		return err
	}

	data.Id = lastId

	// Create relation
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO user_data (id, type_id, entity_id, user_id) VALUES (DEFAULT, $1, $2, $3)",
		ud.TypeCard,
		lastId,
		userId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// SaveFile save file relation to store
func (s *Pg) SaveFile(ctx context.Context, fileInfo storage.FileInfo, userId int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var lastId int
	err = s.db.QueryRowContext(ctx,
		"INSERT INTO user_data_files (id, file_id, meta, path) VALUES (DEFAULT, $1, $2, $3) RETURNING id",
		fileInfo.ID,
		fileInfo.Meta,
		fileInfo.Path,
	).Scan(&lastId)

	if err != nil {
		return err
	}

	// Create relation
	_, err = s.db.ExecContext(ctx,
		"INSERT INTO user_data (id, type_id, entity_id, user_id) VALUES (DEFAULT, $1, $2, $3)",
		ud.TypeFile,
		lastId,
		userId,
	)

	if err != nil {
		return err
	}

	return tx.Commit()
}
