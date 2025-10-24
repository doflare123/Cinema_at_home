package utils

import (
	"cinema/internal/errors"
	stdErrors "errors"
	"net/http"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterPasswordValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
			password := fl.Field().String()

			var (
				hasMinLen  = false
				hasUpper   = false
				hasLower   = false
				hasNumber  = false
				hasSpecial = false
			)

			if len(password) >= 12 {
				hasMinLen = true
			}

			for _, c := range password {
				switch {
				case unicode.IsUpper(c):
					hasUpper = true
				case unicode.IsLower(c):
					hasLower = true
				case unicode.IsDigit(c):
					hasNumber = true
				case unicode.IsPunct(c) || unicode.IsSymbol(c):
					hasSpecial = true
				}
			}

			return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
		})
	}
}

type ValidationErrorResponse struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

func HandleValidationError(c *gin.Context, err error) bool {
	var ve validator.ValidationErrors
	if stdErrors.As(err, &ve) {
		out := make([]ValidationErrorResponse, len(ve))
		for i, fe := range ve {
			switch fe.Tag() {
			case "required":
				out[i] = ValidationErrorResponse{Field: fe.Field(), Error: errors.ErrEmptyPassword.Error()}
			case "password":
				out[i] = ValidationErrorResponse{Field: fe.Field(), Error: errors.ErrInvalidPassword.Error()}
			default:
				out[i] = ValidationErrorResponse{Field: fe.Field(), Error: errors.ErrInvalidServer.Error()}
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": out})
		return true
	}

	return false
}
