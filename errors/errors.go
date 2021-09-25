package errors

import (
	"github.com/genjidb/genji/internal/errors"
	"github.com/genjidb/genji/internal/stringutil"
)

var (
	// ErrDocumentNotFound is returned when no document is associated with the provided key.
	ErrDocumentNotFound = errors.New("document not found")

	// ErrDuplicateDocument is returned when another document is already associated with a given key, primary key,
	// or if there is a unique index violation.
	ErrDuplicateDocument = errors.New("duplicate document")
)

// AlreadyExistsError is returned when to create a table, an index or a sequence
// with a name that is already used by another resource.
type AlreadyExistsError struct {
	Name string
}

func (a AlreadyExistsError) Error() string {
	return stringutil.Sprintf("%q already exists", a.Name)
}

// NotFoundError is returned when the requested table, index or sequence
// doesn't exist.
type NotFoundError struct {
	Name string
}

func (a NotFoundError) Error() string {
	return stringutil.Sprintf("%q not found", a.Name)
}
