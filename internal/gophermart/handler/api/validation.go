package api

import (
	"reflect"
	"strconv"

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

	registerCustomValidations(validator)
}

func registerCustomValidations(validator *govalidator.Validate) {
	err := validator.RegisterValidation("orderID", validateOrderIDField)
	if err != nil {
		panic(err)
	}
}

func validateOrderIDField(field govalidator.FieldLevel) bool {
	var orderID OrderID

	switch field.Field().Kind() { //nolint:exhaustive
	case reflect.String:
		orderNum, err := strconv.ParseUint(field.Field().String(), 10, 64)
		if err != nil {
			return false
		}

		orderID = OrderID(orderNum)
	case reflect.Uint, reflect.Uint64:
		orderID = OrderID(field.Field().Uint())
	default:
		return false
	}

	err := orderID.Validate()

	return err == nil
}
