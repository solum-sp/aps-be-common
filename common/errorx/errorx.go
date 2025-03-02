package errorx

import (
	"fmt"
	"reflect"

	"github.com/solum-sp/aps-be-common/common/loader"
)

// **Default error file path**
const defaultPath = "config/errors.json"

// **Load function: Loads errors.json into the struct and maps field pointers to error codes**
func Load(cfg any, filePath ...string) error {
	// Use provided filePath or fallback to default
	path := defaultPath
	if len(filePath) > 0 && filePath[0] != "" {
		path = filePath[0]
	}

	err := loader.Load(cfg, path)
	if err != nil {
		return fmt.Errorf("failed to load errors config: %w", err)
	}

	// **Dynamically populate fieldToCode mapping**
	structVal := reflect.ValueOf(cfg).Elem()
	structType := structVal.Type()

	fieldToCodeField := structVal.FieldByName("FieldToCode")
	if !fieldToCodeField.IsValid() || fieldToCodeField.Kind() != reflect.Map {
		return fmt.Errorf("cfg must have a 'FieldToCode' map[string]string field")
	}

	// **Initialize the map if it's nil**
	if fieldToCodeField.IsNil() {
		fieldToCodeField.Set(reflect.MakeMap(fieldToCodeField.Type()))
	}

	// **Loop over struct fields and store field names in `fieldToCode`**
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		errorCode := field.Tag.Get("json") // Extract error code from struct tag
		if errorCode == "" {
			continue // Skip fields without JSON tags
		}

		// **Store field name as key in fieldToCode map**
		fieldToCodeField.SetMapIndex(reflect.ValueOf(field.Name), reflect.ValueOf(errorCode))
	}

	return nil
}

// **GetCode retrieves the error code for a given struct field pointer**
func GetCode(cfg any, fieldPtr any) string {
	structVal := reflect.ValueOf(cfg).Elem()

	// **Find the FieldToCode map inside the struct**
	fieldToCodeMap := structVal.FieldByName("FieldToCode")

	// **Ensure it's valid before using it**
	if !fieldToCodeMap.IsValid() || fieldToCodeMap.Kind() != reflect.Map {
		return "unknown_error"
	}

	// **Look up the error code using the field name**
	fieldName := ""
	structType := structVal.Type()
	for i := 0; i < structType.NumField(); i++ {
		if structVal.Field(i).Addr().Interface() == fieldPtr {
			fieldName = structType.Field(i).Name
			break
		}
	}

	// **Check if the field name exists in the map**
	if fieldName == "" {
		return "unknown_error"
	}

	if code := fieldToCodeMap.MapIndex(reflect.ValueOf(fieldName)); code.IsValid() {
		return code.String()
	}

	return "unknown_error"
}
