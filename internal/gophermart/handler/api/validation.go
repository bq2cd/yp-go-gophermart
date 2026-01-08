package api

import (
	"github.com/gin-gonic/gin/binding"
	govalidator "github.com/go-playground/validator/v10"
)

const validateTagName = "validate"

func setupValidator() {
	validator, ok := binding.Validator.Engine().(*govalidator.Validate)
	if !ok {
		return
	}

	validator.SetTagName(validateTagName)
}
