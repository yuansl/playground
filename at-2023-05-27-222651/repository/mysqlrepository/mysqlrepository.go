package mysqlrepository

import (
	"errors"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"playground/at-2023-05-27-222651/repository"
)

var ErrDatabase = errors.New("repository.mysql: database fatal")

type mysqlRepository struct {
	db *sqlx.DB
}

func (repo *mysqlRepository) Save(files []*repository.File) error {
	_, err := repo.db.NamedExec("insert into file(name,foffset,size,data) values(:name,:foffset,:size,:data) ON DUPLICATE KEY UPDATE name=values(name), foffset=values(foffset), size=values(size), data=values(data)",
		files,
	)
	if err != nil {
		return fmt.Errorf("%w: db.NamedExec: %v", ErrDatabase, err)
	}
	return nil
}

func (repo *mysqlRepository) Load(file *repository.File) error {
	rows, err := repo.db.NamedQuery("select name,foffset,size,data from file where name=:name and foffset=:offset",
		map[string]any{
			"name":   file.Name,
			"offset": file.Offset,
		})
	if err != nil {
		return fmt.Errorf("%w: db.NamedQuery: %v", ErrDatabase, err)
	}
	for rows.Next() {
		err = rows.StructScan(file)
		if err != nil {
			return fmt.Errorf("%w: rows.StructScan: %v", ErrDatabase, err)
		}
	}
	return rows.Close()
}

func fatal(v ...any) {
	fmt.Fprintln(os.Stderr, "fatal error: ", v)
	os.Exit(1)
}

func NewRepository(mysqluri string) repository.Repository {
	db, err := sqlx.Open("mysql", mysqluri)
	if err != nil {
		fatal("sql.Open:", err)
	}

	return &mysqlRepository{db: db}
}
