package edge_validators

import (
	"fmt"
	"reflect"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

type CustomManager struct {
	SchemaValidator *validate.SchemaValidator
}

func NewCustomValidator(customResourceValidation *apiextensions.CustomResourceValidation) (*CustomManager, error) {
	sv, _, err := validation.NewSchemaValidator(customResourceValidation)
	if err != nil {
		return nil, err
	}

	return &CustomManager{
		SchemaValidator: sv,
	}, nil
}

func (e *CustomManager) SetPath(path string) {
	e.SchemaValidator.SetPath(path)
}

func (e *CustomManager) Applies(source interface{}, kind reflect.Kind) bool {
	return e.SchemaValidator.Applies(source, kind)
}

func (e *CustomManager) Validate(data interface{}) *validate.Result {
	result := e.SchemaValidator.Validate(data)

	result.AddErrors(fmt.Errorf("namespace [%s] already exists, but not owned by edge custom resource named [%s]", "edge-sample-core", "edge-sample"))

	return result
}
