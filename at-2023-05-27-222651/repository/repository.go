package repository

type File struct {
	Name    string `db:"name"`
	Offset  int64  `db:"foffset"`
	Size    int64  `db:"size"`
	Content []byte `db:"data"`
}

type Repository interface {
	Save([]*File) error
	Load(*File) error
}
