package repository

import (
	"cinema/config"
	"database/sql"

	"cinema/internal/logger"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Model(value interface{}) *gorm.DB
	Select(query interface{}, args ...interface{}) *gorm.DB
	Find(out interface{}, where ...interface{}) *gorm.DB
	Exec(sql string, values ...interface{}) *gorm.DB
	First(out interface{}, where ...interface{}) *gorm.DB
	Raw(sql string, values ...interface{}) *gorm.DB
	Create(value interface{}) *gorm.DB
	Save(value interface{}) *gorm.DB
	Updates(value interface{}) *gorm.DB
	Delete(value interface{}) *gorm.DB
	Where(query interface{}, args ...interface{}) *gorm.DB
	Preload(column string, conditions ...interface{}) *gorm.DB
	Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB
	ScanRows(rows *sql.Rows, result interface{}) error
	Transaction(fc func(tx Repository) error) (err error)
	Close() error
	DropTableIfExists(value interface{}) error
	GetSQLDB() (*sql.DB, error)
	Clauses(conds ...clause.Expression) *gorm.DB
	AutoMigrate(value interface{}) error
}

type repository struct {
	db *gorm.DB
}

type filmRepository struct {
	*repository
}

func NewFilmRepository(logger logger.Logger, conf config.Config) Repository {
	logger.Info("Try database connection")
	db, err := initDB(conf.DBDsn, logger)
	if err != nil {
		logger.Error("Failure database connection", "error", err)
		os.Exit(config.ErrExitStatus)
	}
	logger.Info("Success database connection")
	return &filmRepository{
		repository: &repository{
			db: db,
		},
	}
}

func initDB(dsn string, logger logger.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	logger.Info("connected to DB")
	return db, nil
}

func (rep *repository) Model(value interface{}) *gorm.DB {
	return rep.db.Model(value)
}

func (rep *repository) Select(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Select(query, args...)
}

func (rep *repository) Find(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.Find(out, where...)
}

func (rep *repository) Exec(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Exec(sql, values...)
}

func (rep *repository) First(out interface{}, where ...interface{}) *gorm.DB {
	return rep.db.First(out, where...)
}

func (rep *repository) Raw(sql string, values ...interface{}) *gorm.DB {
	return rep.db.Raw(sql, values...)
}

func (rep *repository) Create(value interface{}) *gorm.DB {
	return rep.db.Create(value)
}

func (rep *repository) Save(value interface{}) *gorm.DB {
	return rep.db.Save(value)
}

func (rep *repository) Updates(value interface{}) *gorm.DB {
	return rep.db.Updates(value)
}

func (rep *repository) Delete(value interface{}) *gorm.DB {
	return rep.db.Delete(value)
}

func (rep *repository) Where(query interface{}, args ...interface{}) *gorm.DB {
	return rep.db.Where(query, args...)
}

func (rep *repository) Preload(column string, conditions ...interface{}) *gorm.DB {
	return rep.db.Preload(column, conditions...)
}

func (rep *repository) Scopes(funcs ...func(*gorm.DB) *gorm.DB) *gorm.DB {
	return rep.db.Scopes(funcs...)
}

func (rep *repository) ScanRows(rows *sql.Rows, result interface{}) error {
	return rep.db.ScanRows(rows, result)
}

func (rep *repository) Close() error {
	sqlDB, _ := rep.db.DB()
	return sqlDB.Close()
}

func (rep *repository) DropTableIfExists(value interface{}) error {
	return rep.db.Migrator().DropTable(value)
}

func (rep *repository) AutoMigrate(value interface{}) error {
	return rep.db.AutoMigrate(value)
}

func (rep *repository) GetSQLDB() (*sql.DB, error) {
	return rep.db.DB()
}

func (rep *repository) Clauses(conds ...clause.Expression) *gorm.DB {
	return rep.db.Clauses(conds...)
}

func (rep *repository) Transaction(fc func(tx Repository) error) (err error) {
	panicked := true
	tx := rep.db.Begin()
	defer func() {
		if panicked || err != nil {
			tx.Rollback()
		}
	}()

	txrep := &repository{}
	txrep.db = tx
	err = fc(txrep)

	if err == nil {
		err = tx.Commit().Error
	}

	panicked = false
	return
}
