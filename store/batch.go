package store

import (
	"database/sql"
	"fmt"
	"strings"
)

func (db *DB) batchInsert(template string, batchSize int, fieldNum int, args []any) error {
	var stmtN *sql.Stmt
	for len(args) > 0 {
		rowNum := len(args) / fieldNum
		if rowNum >= batchSize {
			if stmtN == nil {
				var err error
				stmtN, err = db.prepareStmtN(template, fieldNum, batchSize)
				if err != nil {
					return err
				}
			}
			_, err := stmtN.Exec(args[:batchSize*fieldNum]...)
			if err != nil {
				return err
			}
			args = args[batchSize*fieldNum:]
			continue
		}
		stmt, err := db.prepareStmtN(template, fieldNum, rowNum)
		if err != nil {
			return err
		}
		_, err = stmt.Exec(args...)
		if err != nil {
			return err
		}
		break
	}
	return nil
}

func (db *DB) prepareStmtN(template string, fieldNum, rowNum int) (*sql.Stmt, error) {
	fields := strings.Repeat("?,", fieldNum)
	fields = fields[:len(fields)-1]
	rows := strings.Repeat("("+fields+"),", rowNum)
	rows = rows[:len(rows)-1]
	return db.db.Prepare(fmt.Sprintf(template, rows))
}
