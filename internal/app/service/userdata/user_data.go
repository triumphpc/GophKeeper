// Package userdata contain logic for user data working
package userdata

const (
	TypeText int = 1 + iota
	TypeCard
	TypeFile
)

// UserData service struct
type UserData struct {
	Id       int
	TypeId   int
	EntityId int
	UserId   int
}

// DataText data for text
type DataText struct {
	Id   int
	Name string
	Text string
	Meta string
}

// DataCard data for card
type DataCard struct {
	Id     int
	Number string
	Meta   string
}

// NewDataText returns a new user data text instance
func NewDataText(name string, text string, meta string) *DataText {
	return &DataText{
		Name: name,
		Text: text,
		Meta: meta,
	}
}

// NewDataCard returns a new user data card instance
func NewDataCard(number string, meta string) *DataCard {
	return &DataCard{
		Number: number,
		Meta:   meta,
	}
}
