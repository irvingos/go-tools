package selectdb

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/irvingos/go-tools/logx"
	gmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

// NewForDDL opens SelectDB (Doris) with a MySQL dialector whose Migrator.ColumnTypes does not emit
// "SELECT ... LIMIT 1". Doris/SelectDB may reject that introspection query for some objects
// ("mismatched input 'LIMIT' expecting {<EOF>, ';'}"). Use for cmd/ddl (gorm gen) only; runtime
// code can keep using New.
func NewForDDL(cfg Config) (*gorm.DB, error) {
	dia, err := newDDLMySQLDialector(cfg.buildDSN())
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(dia, &gorm.Config{
		Logger: logx.NewDBLogger(&logx.DBLoggerOptions{
			SlowSQLThreshold: cfg.SlowSQLThreshold,
		}).LogMode(cfg.LogLevel),
		PrepareStmt:            false,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error when init selectdb (ddl), err: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.maxOpenConns())
	sqlDB.SetMaxIdleConns(cfg.maxIdleConns())
	sqlDB.SetConnMaxLifetime(cfg.connMaxLifetime())
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

type ddlMySQLDialector struct {
	*gmysql.Dialector
}

func newDDLMySQLDialector(dsn string) (*ddlMySQLDialector, error) {
	d, ok := gmysql.Open(dsn).(*gmysql.Dialector)
	if !ok {
		return nil, fmt.Errorf("selectdb ddl: mysql.Open did not return *mysql.Dialector")
	}
	return &ddlMySQLDialector{Dialector: d}, nil
}

func (d *ddlMySQLDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return ddlMySQLMigrator{
		Migrator: gmysql.Migrator{
			Migrator: migrator.Migrator{
				Config: migrator.Config{
					DB:        db,
					Dialector: d,
				},
			},
			Dialector: *d.Dialector,
		},
	}
}

type ddlMySQLMigrator struct {
	gmysql.Migrator
}

// ColumnTypes is copied from gorm.io/driver/mysql.Migrator.ColumnTypes with the only change:
// use Where("1 = ?", 0) instead of Limit(1) for the probe query.
func (m ddlMySQLMigrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columnTypes := make([]gorm.ColumnType, 0)
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var (
			currentDatabase, table = m.CurrentSchema(stmt, stmt.Table)
			columnTypeSQL          = "SELECT column_name, column_default, is_nullable = 'YES', data_type, character_maximum_length, column_type, column_key, extra, column_comment, numeric_precision, numeric_scale "
			rows, err              = m.DB.Session(&gorm.Session{}).Table(table).Where("1 = ?", 0).Rows()
		)

		if err != nil {
			return err
		}

		rawColumnTypes, err := rows.ColumnTypes()

		if err != nil {
			return err
		}

		if err := rows.Close(); err != nil {
			return err
		}

		if !m.DisableDatetimePrecision {
			columnTypeSQL += ", datetime_precision "
		}
		columnTypeSQL += "FROM information_schema.columns WHERE table_schema = ? AND table_name = ? ORDER BY ORDINAL_POSITION"

		columns, rowErr := m.DB.Table(table).Raw(columnTypeSQL, currentDatabase, table).Rows()
		if rowErr != nil {
			return rowErr
		}

		defer columns.Close()

		for columns.Next() {
			var (
				column            migrator.ColumnType
				datetimePrecision sql.NullInt64
				extraValue        sql.NullString
				columnKey         sql.NullString
				values            = []interface{}{
					&column.NameValue, &column.DefaultValueValue, &column.NullableValue, &column.DataTypeValue, &column.LengthValue, &column.ColumnTypeValue, &columnKey, &extraValue, &column.CommentValue, &column.DecimalSizeValue, &column.ScaleValue,
				}
			)

			if !m.DisableDatetimePrecision {
				values = append(values, &datetimePrecision)
			}

			if scanErr := columns.Scan(values...); scanErr != nil {
				return scanErr
			}

			column.PrimaryKeyValue = sql.NullBool{Bool: false, Valid: true}
			column.UniqueValue = sql.NullBool{Bool: false, Valid: true}
			switch columnKey.String {
			case "PRI":
				column.PrimaryKeyValue = sql.NullBool{Bool: true, Valid: true}
			case "UNI":
				column.UniqueValue = sql.NullBool{Bool: true, Valid: true}
			}

			if strings.Contains(extraValue.String, "auto_increment") {
				column.AutoIncrementValue = sql.NullBool{Bool: true, Valid: true}
			}

			s := column.DefaultValueValue.String
			for (len(s) >= 3 && s[0] == '\'' && s[len(s)-1] == '\'' && s[len(s)-2] != '\\') ||
				(len(s) == 2 && s == "''") {
				s = s[1 : len(s)-1]
			}
			column.DefaultValueValue.String = s
			if m.Dialector.DontSupportNullAsDefaultValue {
				if column.DefaultValueValue.Valid && column.DefaultValueValue.String == "NULL" {
					column.DefaultValueValue.Valid = false
					column.DefaultValueValue.String = ""
				}
			}

			if datetimePrecision.Valid {
				column.DecimalSizeValue = datetimePrecision
			}

			for _, c := range rawColumnTypes {
				if c.Name() == column.NameValue.String {
					column.SQLColumnType = c
					break
				}
			}

			columnTypes = append(columnTypes, column)
		}

		return nil
	})

	return columnTypes, err
}
