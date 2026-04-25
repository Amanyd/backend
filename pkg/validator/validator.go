package validator

import (
	"reflect"
	"strings"

	"github.com/eulerbutcooler/wingman-backend/pkg/apierr"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

var validRanks = map[string]bool{
	"officer":  true,
	"agniveer": true,
}

func init() {
	validate = validator.New()
	validate.RegisterValidation("rank", func(fl validator.FieldLevel) bool {
		return validRanks[fl.Field().String()]
	})

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func ValidateStruct(s any) error {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	valErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return apierr.BadRequest("invalid input")
	}

	msgs := make([]string, 0, len(valErrs))
	for _, fe := range valErrs {
		field := strings.ToLower(fe.Field())
		var m string
		switch fe.Tag() {
		case "required":
			m = field + " is required"
		case "min":
			m = field + " must be at least " + fe.Param()
		case "oneof":
			m = field + " must be one of " + fe.Param()
		case "rank":
			m = field + " is not a valid rank"
		default:
			m = field + " is invalid"
		}
		msgs = append(msgs, m)
	}

	return apierr.BadRequest(strings.Join(msgs, "; "))
}
