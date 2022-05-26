package data

import (
	"database/sql"
	"github.com/lib/pq"
	"github.com/xuzheng465/greenlight/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	RunTime   Runtime   `json:"runtime,omitempty,string"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.RunTime != 0, "runtime", "must be provided")
	v.Check(movie.RunTime > 0, "runtime", "must be positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	stmt := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`
	args := []interface{}{movie.Title, movie.Year, movie.RunTime, pq.Array(movie.Genres)}

	return m.DB.QueryRow(stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get returns a movie by its ID.
func (m MovieModel) Get(id int64) (*Movie, error) {
	stmt := `
		SELECT id, created_at, title, year, runtime, version
		FROM movies
		WHERE id = $1
	`
	movie := &Movie{}
	err := m.DB.QueryRow(stmt, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.RunTime, &movie.Version)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

// Update movie
func (m MovieModel) Update(movie *Movie) error {
	stmt := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, version = $4
		WHERE id = $5
	`
	_, err := m.DB.Exec(stmt, movie.Title, movie.Year, movie.RunTime, movie.Version, movie.ID)
	return err
}

// Delete movie by ID
func (m MovieModel) Delete(id int64) error {
	stmt := `
		DELETE FROM movies
		WHERE id = $1
	`
	_, err := m.DB.Exec(stmt, id)
	return err
}
