package utils

import (
	"github.com/google/uuid"
)

type UUID uuid.UUID

func (u *UUID) New() UUID {
	return UUID(uuid.New())
}

func (u *UUID) String() string {
	return uuid.UUID(*u).String()
}

func (u *UUID) Parse(s string) (UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return UUID{}, err
	}
	return UUID(id), nil
}

func (u *UUID) MustParse(s string) UUID {
	return UUID(uuid.MustParse(s))
}

func (u *UUID) FromBytes(b []byte) (UUID, error) {
	uid, err := uuid.FromBytes(b)
	if err != nil {
		return UUID{}, err
	}
	return UUID(uid), nil
}
