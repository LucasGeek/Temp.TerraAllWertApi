package errors

import (
	"context"
	"errors"
	"fmt"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Erros padrão da aplicação
var (
	// Erros de autenticação
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrInvalidToken       = errors.New("invalid token")
	
	// Erros de validação
	ErrInvalidInput    = errors.New("invalid input")
	ErrEmailExists     = errors.New("email already exists")
	ErrUsernameExists  = errors.New("username already exists")
	ErrRequiredField   = errors.New("required field")
	
	// Erros de banco de dados
	ErrNotFound         = errors.New("not found")
	ErrInternalServer   = errors.New("internal server error")
	ErrDatabaseError    = errors.New("database error")
	ErrDuplicateKey     = errors.New("duplicate key")
	
	// Erros de arquivo
	ErrFileNotFound     = errors.New("file not found")
	ErrFileUploadFailed = errors.New("file upload failed")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrFileTooLarge     = errors.New("file too large")
	
	// Erros de permissão
	ErrForbidden        = errors.New("forbidden")
	ErrNoPermission     = errors.New("no permission")
)

// NotFoundError retorna um erro customizado de não encontrado
func NotFoundError(entity string) error {
	return fmt.Errorf("%s not found", entity)
}

// ValidationError retorna um erro customizado de validação
func ValidationError(field, message string) error {
	return fmt.Errorf("validation error on field %s: %s", field, message)
}

// DatabaseError retorna um erro customizado de banco de dados
func DatabaseError(operation string, err error) error {
	return fmt.Errorf("database error during %s: %w", operation, err)
}

// FileError retorna um erro customizado de arquivo
func FileError(operation string, err error) error {
	return fmt.Errorf("file error during %s: %w", operation, err)
}

// IsNotFound verifica se o erro é de not found
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized verifica se o erro é de não autorizado
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden verifica se o erro é de forbidden
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsValidationError verifica se o erro é de validação
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidInput) || 
		   errors.Is(err, ErrEmailExists) || 
		   errors.Is(err, ErrUsernameExists) ||
		   errors.Is(err, ErrRequiredField)
}

// ErrorPresenter função para apresentar erros GraphQL
func ErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	return &gqlerror.Error{
		Message: err.Error(),
	}
}

// RecoverFunc função para recuperar de panics
func RecoverFunc(ctx context.Context, err interface{}) error {
	return fmt.Errorf("internal server error: %v", err)
}