package adapters

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part9/internal/trainer/domain/hour"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type mysqlHour struct {
	ID           string    `db:"id"`
	Hour         time.Time `db:"hour"`
	Availability string    `db:"availability"`
}

type MySQLHourRepository struct {
	db          *sqlx.DB
	hourFactory hour.Factory
}

func NewMySQLHourRepository(db *sqlx.DB, hourFactory hour.Factory) *MySQLHourRepository {
	if db == nil {
		panic("missing db")
	}
	if hourFactory.IsZero() {
		panic("missing hourFactory")
	}

	return &MySQLHourRepository{db: db, hourFactory: hourFactory}
}

// sqlContextGetter - это интерфейс, который предоставляет как стандартное подключение к базе данных, так и использующее транзакции
type sqlContextGetter interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (m MySQLHourRepository) GetOrCreateHour(ctx context.Context, time time.Time) (*hour.Hour, error) {
	return m.getOrCreateHour(ctx, m.db, time, false)
}

func (m MySQLHourRepository) getOrCreateHour(
	ctx context.Context,
	db sqlContextGetter,
	hourTime time.Time,
	forUpdate bool,
) (*hour.Hour, error) {
	dbHour := mysqlHour{}

	query := "SELECT * FROM `hours` WHERE `hour` = ?"
	if forUpdate {
		query += " FOR UPDATE"
	}

	err := db.GetContext(ctx, &dbHour, query, hourTime.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		// на самом деле эта дата существует, даже если она не сохраняется
		return m.hourFactory.NewNotAvailableHour(hourTime)
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get hour from db")
	}

	availability, err := hour.NewAvailabilityFromString(dbHour.Availability)
	if err != nil {
		return nil, err
	}

	domainHour, err := m.hourFactory.UnmarshalHourFromDatabase(dbHour.Hour.Local(), availability)
	if err != nil {
		return nil, err
	}

	return domainHour, nil
}

func (m MySQLHourRepository) UpdateHour(
	ctx context.Context,
	hourTime time.Time,
	updateFn func(h *hour.Hour) (*hour.Hour, error),
) (err error) {
	tx, err := m.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "unable to start transaction")
	}

	// Defer в функции выполняется непосредственно перед выходом.
	// Используя defer, мы можем быть уверены, что мы закроем нашу транзакцию соответствующим образом.
	defer func() {
		// В `UpdateHour` мы используем именованный return - `(err error)`.
		// Благодаря этому можно проверить, завершается ли функция с ошибкой.
		//
		// Даже если функция завершается без ошибок, фиксация транзакции может вернуть ошибку.
		// В этом случае мы можем заменить nil на err `err = m.finish ...`.
		err = m.finishTransaction(err, tx)
	}()

	existingHour, err := m.getOrCreateHour(ctx, tx, hourTime, true)
	if err != nil {
		return err
	}

	updatedHour, err := updateFn(existingHour)
	if err != nil {
		return err
	}

	if err := m.upsertHour(tx, updatedHour); err != nil {
		return err
	}

	return nil
}

// upsertHour обновляет hour, если он уже существует в базе данных.
// Если не существует, он вставляется.
func (m MySQLHourRepository) upsertHour(tx *sqlx.Tx, hourToUpdate *hour.Hour) error {
	updatedDbHour := mysqlHour{
		Hour:         hourToUpdate.Time().UTC(),
		Availability: hourToUpdate.Availability().String(),
	}

	_, err := tx.NamedExec(
		`INSERT INTO
					  hours (hour, availability)
			   VALUES
					  (:hour, :availability)
			   ON DUPLICATE	KEY UPDATE
					  availability = :availability`,
		updatedDbHour,
	)
	if err != nil {
		return errors.Wrap(err, "unable to upsert hour")
	}

	return nil
}

// finishTransaction откатывает транзакцию, если указана ошибка.
// Если ошибка равна нулю, транзакция фиксируется.
//
// Если откат не удастся, мы используем библиотеку multierr, чтобы добавить ошибку об ошибке отката.
// Если фиксация не удалась, возвращается ошибка фиксации.
func (m MySQLHourRepository) finishTransaction(err error, tx *sqlx.Tx) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return multierr.Combine(err, rollbackErr)
		}

		return err
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.Wrap(err, "failed to commit tx")
		}

		return nil
	}
}

func NewMySQLConnection() (*sqlx.DB, error) {
	config := mysql.Config{
		Addr:      os.Getenv("MYSQL_ADDR"),
		User:      os.Getenv("MYSQL_USER"),
		Passwd:    os.Getenv("MYSQL_PASSWORD"),
		DBName:    os.Getenv("MYSQL_DATABASE"),
		ParseTime: true, // с этим параметром мы можем использовать time.Time в mysqlHour.Hour
	}

	db, err := sqlx.Connect("mysql", config.FormatDSN())
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to MySQL")
	}

	return db, nil
}
