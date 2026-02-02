package handlers

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// decodeRequest декодирует тело запроса из JSON или form data
func (h *EventHandler) decodeRequest(r *http.Request, v interface{}) error {
	contentType := r.Header.Get("Content-Type")
	
	// Проверить, является ли content type JSON (может включать charset)
	if len(contentType) >= 16 && contentType[:16] == "application/json" {
		return json.NewDecoder(r.Body).Decode(v)
	}
	
	// Обработать form data
	if err := r.ParseForm(); err != nil {
		return err
	}
	
	// Использовать рефлексию для заполнения структуры из значений формы
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		formTag := field.Tag.Get("form")
		if formTag == "" {
			formTag = field.Tag.Get("json")
		}
		if formTag == "" || formTag == "-" {
			continue
		}
		
		value := r.FormValue(formTag)
		if value != "" {
			fieldValue := rv.Field(i)
			if fieldValue.CanSet() && fieldValue.Kind() == reflect.String {
				fieldValue.SetString(value)
			}
		}
	}
	
	return nil
}

