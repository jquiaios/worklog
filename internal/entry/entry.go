package entry

import "time"

type Type string

const (
	Highlight Type = "highlight"
	Lowlight  Type = "lowlight"
)

type Entry struct {
	ID        int64
	Type      Type
	Body      string
	CreatedAt time.Time
}
