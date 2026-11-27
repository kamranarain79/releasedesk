package user

import "strings"

// Person is the name and photo shown for someone in an audit row.
type Person struct {
	UserID int
	Name   string
}

// Directory resolves a stored actor string to a console user.
type Directory map[string]User

func NewDirectory(accounts []User) Directory {
	directory := Directory{}
	for _, account := range accounts {
		directory[account.Username] = account
	}
	return directory
}

// Person turns a username or ci: token prefix into a display name and avatar id.
func (d Directory) Person(actor string) Person {
	actor = strings.TrimSpace(actor)
	if actor == "" || actor == "unknown" {
		return Person{Name: "Unknown"}
	}
	if strings.HasPrefix(actor, "ci:") {
		return Person{Name: actor}
	}
	account, ok := d[actor]
	if !ok {
		return Person{Name: actor}
	}
	name := strings.TrimSpace(account.FirstName.String + " " + account.LastName.String)
	if name == "" {
		name = account.Username
	}
	return Person{UserID: account.ID, Name: name}
}
