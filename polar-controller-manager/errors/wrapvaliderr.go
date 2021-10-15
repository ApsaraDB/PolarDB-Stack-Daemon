/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package errors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/go-playground/validator.v9"
	log "k8s.io/klog"
)

// ValidationErrorToText
func ValidationErrorToText(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", e.Field())
	case "max":
		return fmt.Sprintf("%s cannot be longer than %s", e.Field(), e.Param())
	case "min":
		return fmt.Sprintf("%s must be longer than %s", e.Field(), e.Param())
	case "email":
		return fmt.Sprintf("%s is not the correct format", e.Field())
	case "len":
		return fmt.Sprintf("%s must be %s characters long", e.Field(), e.Param())
	case "phone":
		return fmt.Sprintf("%s is not the correct format", e.Field())
	case "pg":
		return fmt.Sprintf("%s 不能以pg开头", e.Field())
	case "comment":
		return fmt.Sprintf("%s 以大小字母或中文开头,可包含数字,_或-", e.Field())
	case "userId":
		// 用户名格式错误，用户名由小写字母、下划线、数字组成，必须以字母开头，字母或数字结尾
		return fmt.Sprintf("%s is not the correct format", e.Field())
	case "regExp":
		return fmt.Sprintf("%s 不符合当前正则规则，请重新输入", e.Field())
	}
	return fmt.Sprintf("%s is not valid", e.Field())
}

// CommonError error
type CommonError struct {
	Errors map[string]interface{} `json:"errors"`
}

// ThrowValidatorError ThrowValidatorError
func ThrowValidatorError(err error) CommonError {
	res := CommonError{}
	res.Errors = make(map[string]interface{})
	switch o := err.(type) {
	case *json.UnmarshalTypeError:
		res.Errors[o.Field] = o.Error()
	case validator.ValidationErrors:
		for _, e := range o {
			res.Errors[e.Field()] = ValidationErrorToText(e)
		}
	default:
		log.Error(err)
		res.Errors[""] = err.Error()
	}
	return res
}

// NewValidatorErrorStr NewValidatorErrorStr
func NewValidatorErrorStr(err error) string {
	b := new(bytes.Buffer)
	for _, value := range ThrowValidatorError(err).Errors {
		fmt.Fprintf(b, "%s ", value)
	}
	return b.String()
}

// NewValidatorError NewValidatorError
func NewValidatorError(err error) error {
	var errors []string

	for _, value := range ThrowValidatorError(err).Errors {
		errors = append(errors, fmt.Sprintf("%s", value))
	}

	return NewNormalError("ParamInvalidErr", strings.Join(errors, ","))
}
