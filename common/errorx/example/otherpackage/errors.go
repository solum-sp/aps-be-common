package otherpackage

import "github.com/solum-sp/aps-be-common/common/errorx"

type SysError struct {
	ErrIncorrectPassword string `json:"104001"`
	ErrInvalidJWTToken   string `json:"104002"`
	ErrTokenExpired      string `json:"104003"`
	ErrUserDoesNotExist  string `json:"104041"`
	ErrNoContent         string `json:"NoContent"`

	FieldToCode map[string]string `json:"-"`
}

func (e *SysError) GetCode(fieldPtr any) string {
	return errorx.GetCode(e, fieldPtr)
}
