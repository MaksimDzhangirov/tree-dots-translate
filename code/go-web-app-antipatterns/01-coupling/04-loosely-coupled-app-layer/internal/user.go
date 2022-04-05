package internal

import (
	"errors"
	"strings"
)

var (
	ErrNameRequired  = errors.New("either first name or last name is required")
	ErrEmailRequired = errors.New("email address is required")
	ErrInvalidEmail  = errors.New("invalid email address")
)

type User struct {
	id        int
	firstName string
	lastName  string
	emails    []Email
}

func NewUser(firstName string, lastName string, emailAddress string) (User, error) {
	if firstName == "" && lastName == "" {
		return User{}, ErrNameRequired
	}

	email, err := NewEmail(emailAddress, true)
	if err != nil {
		return User{}, err
	}

	return User{
		firstName: firstName,
		lastName:  lastName,
		emails:    []Email{email},
	}, nil
}

// UnmarshalUser загружает пользователя из базы данных. Функцию не следует использовать для чего-либо ещё
func UnmarshalUser(id int, firstName string, lastName string, emails []Email) User {
	return User{
		id:        id,
		firstName: firstName,
		lastName:  lastName,
		emails:    emails,
	}
}

func (u User) ID() int {
	return u.id
}

func (u User) FirstName() string {
	return u.firstName
}

func (u User) LastName() string {
	return u.lastName
}

func (u User) Emails() []Email {
	return u.emails
}

func (u User) PrimaryEmail() Email {
	for _, e := range u.emails {
		if e.primary {
			return e
		}
	}

	// Обычно вызов panic в коде логики - плохая практика.
	// Но поскольку мы уверены, что правильно созданный пользователь должен иметь основной адрес электронной почты, такой ситуации никогда не должно быть.
	// В редких случаях, когда это происходит, мы выбираем панику. middleware Recoverer перехватит её за нас.
	panic("no primary email found")
}

func (u *User) ChangeName(newFirstName *string, newLastName *string) error {
	if newFirstName == nil && newLastName == nil {
		return nil
	}

	firstName := u.firstName
	lastName := u.lastName

	if newFirstName != nil {
		firstName = *newFirstName
	}

	if newLastName != nil {
		lastName = *newLastName
	}

	if firstName == "" && lastName == "" {
		return ErrNameRequired
	}

	u.firstName = firstName
	u.lastName = lastName

	return nil
}

func (u User) DisplayName() string {
	if u.firstName != "" {
		if u.lastName != "" {
			return u.firstName + " " + u.lastName
		}

		return u.firstName
	}

	return u.lastName
}

type Email struct {
	address string
	primary bool
}

func NewEmail(address string, primary bool) (Email, error) {
	if address == "" {
		return Email{}, ErrEmailRequired
	}

	// Простейшая проверка, чтобы пример не был слишком длинным, но вы поняли идею
	if !strings.Contains(address, "@") {
		return Email{}, ErrInvalidEmail
	}

	return Email{
		address: address,
		primary: primary,
	}, nil
}

// UnmarshalEmail загружает e-mail из базы данных. Функция не должна использоваться для чего-либо ещё
func UnmarshalEmail(address string, primary bool) Email {
	return Email{
		address: address,
		primary: primary,
	}
}

func (e Email) Address() string {
	return e.address
}

func (e Email) Primary() bool {
	return e.primary
}
