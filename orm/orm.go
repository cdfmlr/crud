package orm

import (
	"github.com/cdfmlr/crud/log"
	"gorm.io/gorm"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
)

type DBDriver string

// available database drivers
const (
	DBDriverMySQL    = "mysql"
	DBDriverSqlite   = "sqlite"
	DBDriverPostgres = "postgres"
)

// DB is the global database instance
var DB *gorm.DB

var logger = log.ZoneLogger("crud/orm")

// OpenDB initializes the global DB instance
func OpenDB(driver DBDriver, dsn string) (*gorm.DB, error) {
	var err error

	driverOpen := getDBOpener(driver)

	DB, err = gorm.Open(driverOpen(dsn), &gorm.Config{
		Logger: log.Logger4Gorm,
	})
	return DB, err
}

// region dbOpener

// DBOpener opens a gorm Dialector.
//
// See:
// 	- gorm.io/driver/mysql:    https://github.com/go-gorm/mysql/blob/f46a79cf94a9d67edcc7d5f6f2606e21bf6525fe/mysql.go#L52
// 	- gorm.io/driver/postgres: https://github.com/go-gorm/postgres/blob/c2cfceb161687324cb399c9f60ec775428335957/postgres.go#L31
// 	- gorm.io/driver/sqlite:   https://github.com/go-gorm/sqlite/blob/1d1e7723862758a6e6a860f90f3e7a3bea9cc94a/sqlite.go#L28
type DBOpener func(dsn string) gorm.Dialector

// 获取连接某种数据库的驱动
func getDBOpener(driver DBDriver) DBOpener {
	switch driver {
	case DBDriverMySQL:
		return mysql.Open
	case DBDriverSqlite:
		return sqlite.Open
	case DBDriverPostgres:
		return postgres.Open
	default:
		//panic("unknown database driver: " + driver)
		logger.WithField("driver", driver).
			Fatal("unknown database driver")
	}
	return nil // unreachable, make compiler happy
}

// endregion dbOpener

func RegisterModel(m ...any) error {
	err := DB.AutoMigrate(m...)
	if err != nil {
		logger.WithError(err).
			Errorf("RegisterModel: AutoMigrate failed")
		return err
	}
	return nil
}
