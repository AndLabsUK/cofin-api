package models

type MessageAuthor string

const (
	UserAuthor MessageAuthor = "user"
	AIAuthor   MessageAuthor = "ai"
)

type Message struct {
	Generic

	UserID uint `gorm:"index;not null"`
	User   User
	Author MessageAuthor
	Kind   SourceKind `gorm:"index;not null"`
	Text   string
}
