package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apimachyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	cust_validators "github.com/nXnus25/k8s-crd-custom-schema-validation/validators"
)

func main() {
	// validators := map[schema.GroupVersionKind]*validate.SchemaValidator{}
	validators := map[schema.GroupVersionKind]*cust_validators.CustomManager{}
	if err := fs.WalkDir(os.DirFS("./crds"), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := os.Open(filepath.Join("./crds", p))
		defer f.Close()
		yr := apimachyaml.NewYAMLReader(bufio.NewReader(f))
		for {
			b, err := yr.Read()
			if err != nil && err != io.EOF {
				return err
			}
			if err == io.EOF {
				break
			}
			if len(b) == 0 {
				continue
			}
			crd := &extv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(b, crd); err != nil {
				return err
			}
			internal := &ext.CustomResourceDefinition{}
			if err := extv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(crd, internal, nil); err != nil {
				return err
			}
			fmt.Println(crd)
			fmt.Println()
			fmt.Println(internal)
			for _, ver := range internal.Spec.Versions {
				// var sv *validate.SchemaValidator
				// var sv *validators.EdgeValidator
				var err error
				// sv, _, err = validation.NewSchemaValidator(ver.Schema)
				sv, err := cust_validators.NewCustomValidator(ver.Schema)
				if err != nil {
					return err
				}
				if internal.Spec.Validation != nil {
					// sv, _, err = validation.NewSchemaValidator(internal.Spec.Validation)
					sv, err = cust_validators.NewCustomValidator(internal.Spec.Validation)
					if err != nil {
						return err
					}
				}
				validators[schema.GroupVersionKind{
					Group:   internal.Spec.Group,
					Version: ver.Name,
					Kind:    internal.Spec.Names.Kind,
				}] = sv
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
	b, err := ioutil.ReadFile("cr.yaml")
	if err != nil {
		panic(err)
	}
	obj := &unstructured.Unstructured{}
	if err := yaml.Unmarshal(b, obj); err != nil {
		panic(err)
	}
	v, ok := validators[obj.GetObjectKind().GroupVersionKind()]
	if !ok {
		panic("could not find validator for: " + obj.GetObjectKind().GroupVersionKind().String())
	}
	re := v.Validate(obj)
	for i, e := range re.Errors {
		fmt.Printf("Validation Error %d (%s)(%s): %s\n", i, obj.GroupVersionKind().String(), obj.GetName(), e.Error())
	}
}
