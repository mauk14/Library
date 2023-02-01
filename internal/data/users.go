package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"mauk14.library/internal/validator"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	AnonymousUser     = &User{}
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   uuid.UUID `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB DB
}

func (m UserModel) Insert(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	id, err := m.DB.GetLastId(ctx, "", "users")
	fmt.Println(err)

	if err != nil {
		return err
	}

	user.ID = id + 1
	user.Version = uuid.New()
	user.CreatedAt = time.Now()

	err = m.DB.Insert(ctx, "", "users", user)
	fmt.Println(err)

	if err != nil {
		switch {
		case err.Error() == `write exception: write errors: [E11000 duplicate key error collection: Library.users index: email_1 dup key: { email: "`+user.Email+`" }]`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil

}

func (m UserModel) GetByEmail(email string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var user User
	var result interface{}

	result, err := m.DB.Get(ctx, "", email, "users", "")
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	user = result.(User)

	return &user, nil
}

func (m *UserModel) Update(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.Update(ctx, "", user)

	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.Get(ctx, "", tokenHash, "tokens", tokenScope)
	if err != nil {
		switch {
		case errors.Is(err, mongo.ErrNoDocuments):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}

	}

	user = result.(User)

	return &user, nil
}
