package validation

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("slug", validateSlug)
	validate.RegisterValidation("state_code", validateStateCode)
	validate.RegisterValidation("sun_position", validateSunPosition)
	validate.RegisterValidation("suite_status", validateSuiteStatus)
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	// Slug should contain only lowercase letters, numbers, and hyphens
	match, _ := regexp.MatchString("^[a-z0-9-]+$", slug)
	return match && len(slug) >= 2 && len(slug) <= 255
}

func validateStateCode(fl validator.FieldLevel) bool {
	state := fl.Field().String()
	// Brazilian state codes (UF)
	validStates := []string{
		"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA",
		"MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN",
		"RS", "RO", "RR", "SC", "SP", "SE", "TO",
	}
	
	for _, validState := range validStates {
		if strings.ToUpper(state) == validState {
			return true
		}
	}
	return false
}

func validateSunPosition(fl validator.FieldLevel) bool {
	position := fl.Field().String()
	validPositions := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	
	for _, validPosition := range validPositions {
		if strings.ToUpper(position) == validPosition {
			return true
		}
	}
	return false
}

func validateSuiteStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	validStatuses := []string{"available", "reserved", "sold", "unavailable"}
	
	for _, validStatus := range validStatuses {
		if strings.ToLower(status) == validStatus {
			return true
		}
	}
	return false
}

type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, ve := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   ve.Field(),
				Tag:     ve.Tag(),
				Message: getErrorMessage(ve),
			})
		}
	}
	
	return errors
}

func getErrorMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return ve.Field() + " is required"
	case "email":
		return ve.Field() + " must be a valid email address"
	case "min":
		return ve.Field() + " must be at least " + ve.Param() + " characters"
	case "max":
		return ve.Field() + " must be at most " + ve.Param() + " characters"
	case "len":
		return ve.Field() + " must be exactly " + ve.Param() + " characters"
	case "slug":
		return ve.Field() + " must be a valid slug (lowercase letters, numbers, and hyphens only)"
	case "state_code":
		return ve.Field() + " must be a valid Brazilian state code"
	case "sun_position":
		return ve.Field() + " must be a valid sun position (N, NE, E, SE, S, SW, W, NW)"
	case "suite_status":
		return ve.Field() + " must be a valid status (available, reserved, sold, unavailable)"
	default:
		return ve.Field() + " is invalid"
	}
}