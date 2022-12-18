package yaml

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Configurator struct {
	dataMap       map[string]map[string]interface{}
	lastAliasName string
}

func NewConfigurator() *Configurator {
	return &Configurator{}
}

func (this *Configurator) ReadFile(fileName string) error {
	body, err := readFile(fileName)
	if err != nil {
		return err
	}
	return this.setNewSource(body)
}

func readFile(fileName string) ([]byte, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

func (this *Configurator) setNewSource(src []byte) error {
	this.dataMap = nil
	if err := yaml.Unmarshal(src, &this.dataMap); err != nil {
		return err
	}
	return nil
}

func (this *Configurator) ParseToStruct(packStruct interface{}, aliasName string) error {
	structVal := reflect.ValueOf(packStruct).Elem()
	this.lastAliasName = aliasName

	aliasValue, exists := this.dataMap[aliasName]
	if exists == false {
		return fmt.Errorf("Алиас <%s> отсутствует в конфигурационном файле", aliasName)
	}
	return this.switchSetType(structVal, aliasValue, structVal.Type(), "", reflect.Value{})
}

/*	Рекурсивная функция заполнения полей конфига  */
func (this *Configurator) switchSetType(field reflect.Value, value interface{}, ftype reflect.Type, ftag reflect.StructTag, map_key reflect.Value) error {
	switch ftype.Kind() {
	case reflect.Slice:
		if value != nil {
			v_slice := reflect.ValueOf(value)
			t_slice := ftype.Elem()
			slice := reflect.MakeSlice(reflect.SliceOf(t_slice), v_slice.Len(), v_slice.Cap())
			for j := 0; j < v_slice.Len(); j++ {
				slice.Index(j).Set(reflect.Zero(t_slice))
				if err := this.switchSetType(slice.Index(j), v_slice.Index(j).Interface(), t_slice, ftag, reflect.Value{}); err != nil {
					return fmt.Errorf("%w (поле номер %d)", err, j)
				}
			}
			if field.Type().Kind() == reflect.Map {
				field.SetMapIndex(map_key, slice)
			} else {
				field.Set(slice)
			}
		}
	case reflect.Map:
		if value != nil {
			v_map := reflect.ValueOf(value)
			field.Set(reflect.MakeMap(ftype))

			for _, k := range v_map.MapKeys() {
				val_json := v_map.MapIndex(k)
				n_key := reflect.ValueOf(k.Interface().(string))
				field.SetMapIndex(n_key, reflect.Zero(ftype.Elem()))
				if err := this.switchSetType(field, val_json.Interface(), ftype.Elem(), ftag, n_key); err != nil {
					return err
				}
			}
		}
	/*	Обработка структуры. Если нет тега conf - поле не обрабатывается
	**	Для исчислимых можно добавлять тэги min max
	**	Для строковых можно добавлять тэг env (заполнить поле значением из переменной окружения)
	**	Для строковых и исчислимых можно добавлять тэг enum - выбор из допустимых значений  */
	case reflect.Struct:
		var structValue map[string]interface{}
		switch typedValue := value.(type) {
		case map[string]interface{}:
			structValue = typedValue
		case map[interface{}]interface{}:
			structValue = cleanupInterfaceMap(typedValue)
		default:
			return fmt.Errorf("Тело структуры невозможно заполнить так как попался необрабатываемый тип %T", value)
		}
		for i := 0; i < ftype.NumField(); i++ {
			tag := ftype.Field(i).Tag.Get("conf")
			minTag := ftype.Field(i).Tag.Get("min")
			maxTag := ftype.Field(i).Tag.Get("max")
			envTag := ftype.Field(i).Tag.Get("env")
			enumTag := ftype.Field(i).Tag.Get("enum")
			if tag == "" || tag == "-" {
				continue
			}
			value_child, exist := structValue[tag]
			if exist == false {
				return fmt.Errorf("Для поля %s не задано значение (алиас %s)", tag, this.lastAliasName)
			}
			if minTag != "" && minTag != "-" {
				if isCountableType(ftype.Field(i).Type, field.Field(i)) == false {
					return fmt.Errorf("Поле %s имеет тэг min но при этом не является исчислимым (алиас %s)", tag, this.lastAliasName)
				}
				if err := this.checkMinFieldValue(ftype.Field(i).Type, field.Field(i), value_child, minTag); err != nil {
					return fmt.Errorf("%w (поле %s, алиас %s)", err, tag, this.lastAliasName)
				}
			}
			if maxTag != "" && maxTag != "-" {
				if isCountableType(ftype.Field(i).Type, field.Field(i)) == false {
					return fmt.Errorf("Поле %s имеет тэг max но при этом не является исчислимым (алиас %s)", tag, this.lastAliasName)
				}
				if err := this.checkMaxFieldValue(ftype.Field(i).Type, field.Field(i), value_child, maxTag); err != nil {
					return fmt.Errorf("%w (поле %s, алиас %s)", err, tag, this.lastAliasName)
				}
			}
			if envTag != "" && envTag != "-" {
				result, err := strconv.ParseBool(envTag)
				if err != nil || result != true {
					return fmt.Errorf("Поле %s имеет тэг env но при этом не установлено в true (алиас %s)", tag, this.lastAliasName)
				}
				if isStringType(ftype.Field(i).Type, field.Field(i)) == false {
					return fmt.Errorf("Поле %s имеет тэг env но при этом не является строкой (алиас %s)", tag, this.lastAliasName)
				}
				string_value_child, ok := value_child.(string)
				if ok == false {
					return fmt.Errorf("Поле %s имеет тэг env но при этом не является строкой (алиас %s)", tag, this.lastAliasName)
				}
				path, exists := os.LookupEnv(string_value_child)
				if exists == false {
					return fmt.Errorf("Поле %s имеет тэг env но переменная окружения %s не обнаружена в системе (алиас %s)", tag, value_child, this.lastAliasName)
				}
				value_child = path
			}
			if enumTag != "" {
				if isCountableType(ftype.Field(i).Type, field.Field(i)) == true || isStringType(ftype.Field(i).Type, field.Field(i)) == true {
					if err := this.checkEnum(ftype.Field(i).Type, field.Field(i), value_child, enumTag); err != nil {
						return fmt.Errorf("%w (поле %s, алиас %s)", err, tag, this.lastAliasName)
					}
				} else {
					return fmt.Errorf("Поле %s имеет тэг enum но при этом не является ни исчислимым ни строкой (алиас %s)", tag, this.lastAliasName)
				}
			}
			/*	Рекурсия  */
			if err := this.switchSetType(field.Field(i), value_child, ftype.Field(i).Type, ftype.Field(i).Tag, reflect.Value{}); err != nil {
				return fmt.Errorf("%w (поле %s, алиас %s)", err, tag, this.lastAliasName)
			}
		}
	case reflect.Ptr:
		if value != nil {
			/*	Рекурсия. В случае nil из конфигурационника - ошибкой не считается  */
			field_type_child := field.Type()
			field.Set(reflect.New(field_type_child.Elem()))
			if err := this.switchSetType(field.Elem(), value, field_type_child.Elem(), ftag, reflect.Value{}); err != nil {
				return err
			}
		}
	default:
		val, err := this.primitiveType(ftype, value, ftag)
		if err != nil {
			return err
		}
		if field.CanAddr() && field.Type().Kind() != reflect.Map {
			field.Set(val)
		} else if field.Type().Kind() == reflect.Map {
			field.SetMapIndex(map_key, val)
		}
	}
	return nil
}

func cleanupInterfaceMap(in map[interface{}]interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range in {
		res[fmt.Sprintf("%v", k)] = cleanupMapValue(v)
	}
	return res
}

func cleanupMapValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanupInterfaceArray(v)
	case map[interface{}]interface{}:
		return cleanupInterfaceMap(v)
	case string:
		return v
	default:
		return v
	}
}
func cleanupInterfaceArray(in []interface{}) []interface{} {
	res := make([]interface{}, len(in))
	for i, v := range in {
		res[i] = cleanupMapValue(v)
	}
	return res
}

func isCountableType(ftype reflect.Type, field reflect.Value) bool {
	switch ftype.Kind() {
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32:
		return true
	case reflect.Ptr:
		switch field.Type().Elem().Kind() {
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int32, reflect.Int64, reflect.Float64, reflect.Float32:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func isStringType(ftype reflect.Type, field reflect.Value) bool {
	switch ftype.Kind() {
	case reflect.String:
		return true
	case reflect.Ptr:
		switch field.Type().Elem().Kind() {
		case reflect.String:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func (this *Configurator) checkEnum(ftype reflect.Type, field reflect.Value, value interface{}, enumTag string) error {
	parts := strings.Split(enumTag, ";")
	var wasFound bool
	for _, enumItem := range parts {
		switch ftype.Kind() {
		case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int32, reflect.Int64:
			typedTag, err := strconv.ParseInt(enumItem, 10, 64)
			if err != nil {
				fmt.Errorf("Не смог распарсить часть тэга enum (%s) структуры в целочисленный тип (%w)", enumItem, err)
			}
			switch typedValue := value.(type) {
			case uint:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			case uint64:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			case uint32:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			case int:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			case int64:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			case int32:
				if int64(typedValue) == int64(typedTag) {
					wasFound = true
				}
			default:
				return fmt.Errorf("Невозможно сравнить значение целочисленного типа и нецелочисленного для валидации enum")
			}
		case reflect.Float64, reflect.Float32:
			typedTag, err := strconv.ParseFloat(enumItem, 64)
			if err != nil {
				fmt.Errorf("Не смог распарсить часть тэга enum (%s) структуры в тип float (%w)", enumItem, err)
			}
			switch typedValue := value.(type) {
			case uint:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case uint64:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case uint32:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case int:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case int64:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case int32:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case float32:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			case float64:
				if float64(typedValue) == float64(typedTag) {
					wasFound = true
				}
			default:
				return fmt.Errorf("Невозможно сравнить значение вещественного (float) типа и невещественного для валидации enum")
			}
		case reflect.String:
			switch typedValue := value.(type) {
			case string:
				if typedValue == enumItem {
					wasFound = true
				}
			default:
				return fmt.Errorf("Невозможно сравнить значение строкового типа и нестрокового для валидации enum")
			}
		case reflect.Ptr:
			return this.checkEnum(field.Type().Elem(), field, value, enumTag)
		}
	}
	if wasFound == false {
		return fmt.Errorf("Поле не соответствует ни одному из перечисленный в enum значений (%s)", enumTag)
	}
	return nil
}

func (this *Configurator) checkMinFieldValue(ftype reflect.Type, field reflect.Value, value interface{}, tagMinValue string) error {
	switch ftype.Kind() {
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int32, reflect.Int64:
		minValue, err := strconv.ParseInt(tagMinValue, 10, 64)
		if err != nil {
			fmt.Errorf("Не смог распарсить тэг min структуры в целочисленный тип (%w)", err)
		}
		switch typedValue := value.(type) {
		case uint:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		case uint64:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		case uint32:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		case int:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		case int64:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		case int32:
			if int64(typedValue) < int64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d меньше значения %d заданного тэгом min заполняемой структуры", int(typedValue), int(minValue))
			}
		default:
			return fmt.Errorf("Невозможно сравнить значение целочисленного типа и нецелочисленного для проверки минимального значения")
		}
	case reflect.Float64, reflect.Float32:
		minValue, err := strconv.ParseFloat(tagMinValue, 64)
		if err != nil {
			fmt.Errorf("Не смог распарсить тэг min структуры в тип float (%w)", err)
		}
		switch typedValue := value.(type) {
		case uint:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case uint64:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case uint32:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case int:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case int64:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case int32:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case float64:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		case float32:
			if float64(typedValue) < float64(minValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f меньше значения %f заданного тэгом min заполняемой структуры", float64(typedValue), float64(minValue))
			}
		default:
			return fmt.Errorf("Невозможно сравнить значение вещественного (float) типа и невещественного для проверки минимального значения")
		}
	case reflect.Ptr:
		return this.checkMinFieldValue(field.Type().Elem(), field, value, tagMinValue)
	}
	return nil
}

func (this *Configurator) checkMaxFieldValue(ftype reflect.Type, field reflect.Value, value interface{}, tagMaxValue string) error {
	switch ftype.Kind() {
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int32, reflect.Int64:
		maxValue, err := strconv.ParseInt(tagMaxValue, 10, 64)
		if err != nil {
			fmt.Errorf("Не смог распарсить тэг max структуры в целочисленный тип (%w)", err)
		}
		switch typedValue := value.(type) {
		case uint:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		case uint64:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		case uint32:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		case int:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		case int64:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		case int32:
			if int64(typedValue) > int64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %d больше значения %d заданного тэгом max заполняемой структуры", int(typedValue), int(maxValue))
			}
		default:
			return fmt.Errorf("Невозможно сравнить значение целочисленного типа и нецелочисленного для проверки максимального значения")
		}
	case reflect.Float64, reflect.Float32:
		maxValue, err := strconv.ParseFloat(tagMaxValue, 64)
		if err != nil {
			fmt.Errorf("Не смог распарсить тэг max структуры в тип float (%w)", err)
		}
		switch typedValue := value.(type) {
		case uint:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case uint64:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case uint32:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case int:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case int64:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case int32:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case float64:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		case float32:
			if float64(typedValue) > float64(maxValue) {
				return fmt.Errorf("Значение поля в конфигурационном файле %f больше значения %f заданного тэгом max заполняемой структуры", float64(typedValue), float64(maxValue))
			}
		default:
			return fmt.Errorf("Невозможно сравнить значение вещественного (float) типа и невещественного для проверки максимального значения")
		}
	case reflect.Ptr:
		return this.checkMaxFieldValue(field.Type().Elem(), field, value, tagMaxValue)
	}
	return nil
}

// Получение значений для простых типов
func (this *Configurator) primitiveType(ftype reflect.Type, value interface{}, tag reflect.StructTag) (reflect.Value, error) {
	switch ftype.Kind() {
	case reflect.String:
		switch typedValue := value.(type) {
		case string:
			return reflect.ValueOf(string(typedValue)), nil
		case int:
			return reflect.ValueOf(strconv.Itoa(typedValue)), nil
		case float64:
			return reflect.ValueOf(strconv.FormatFloat(typedValue, 'E', -1, 64)), nil
		case bool:
			return reflect.ValueOf(strconv.FormatBool(typedValue)), nil
		default:
			return reflect.ValueOf(string("")), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Uint:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(uint(uint64Val)), nil
		case int:
			return reflect.ValueOf(uint(typedValue)), nil
		default:
			return reflect.ValueOf(uint(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Uint64:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(uint64Val), nil
		case int:
			return reflect.ValueOf(uint64(typedValue)), nil
		default:
			return reflect.ValueOf(uint64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Uint32:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(uint32(uint64Val)), nil
		case int:
			return reflect.ValueOf(uint32(typedValue)), nil
		default:
			return reflect.ValueOf(uint32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Int:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(int64Val), nil
		case int:
			return reflect.ValueOf(int(typedValue)), nil
		default:
			return reflect.ValueOf(int(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Int64:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(int64(int64Val)), nil
		case int:
			return reflect.ValueOf(int64(typedValue)), nil
		default:
			return reflect.ValueOf(int64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Int32:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(int32(int64Val)), nil
		case int:
			return reflect.ValueOf(int32(typedValue)), nil
		default:
			return reflect.ValueOf(int32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Bool:
		switch typedValue := value.(type) {
		case string:
			boolVal, err := strconv.ParseBool(typedValue)
			if err != nil {
				return reflect.ValueOf(bool(false)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(bool(boolVal)), nil
		case bool:
			return reflect.ValueOf(bool(typedValue)), nil
		default:
			return reflect.ValueOf(bool(false)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Float64:
		switch typedValue := value.(type) {
		case string:
			float64Val, err := strconv.ParseFloat(typedValue, 64)
			if err != nil {
				return reflect.ValueOf(float64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(float64(float64Val)), nil
		case float64:
			return reflect.ValueOf(float64(typedValue)), nil
		default:
			return reflect.ValueOf(float64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	case reflect.Float32:
		switch typedValue := value.(type) {
		case string:
			float64Val, err := strconv.ParseFloat(typedValue, 64)
			if err != nil {
				return reflect.ValueOf(float32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
			}
			return reflect.ValueOf(float32(float64Val)), nil
		case float64:
			return reflect.ValueOf(float32(typedValue)), nil
		default:
			return reflect.ValueOf(float32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
		}
	default:
		return reflect.Value{}, typeError(ftype, fmt.Sprintf("%T", value), this.lastAliasName)
	}
}

// Возможность в конфиге задавать время в разных единицах.
func multiplyDuration(duration time.Duration, multiplier string) time.Duration {
	switch multiplier {
	case "Microsecond":
		return duration * time.Microsecond
	case "Millisecond":
		return duration * time.Millisecond
	case "Second":
		return duration * time.Second
	case "Minute":
		return duration * time.Minute
	case "Hour":
		return duration * time.Hour
	default:
		return duration
	}
}

func typeError(ftype reflect.Type, valueType, aliasName string) error {
	return fmt.Errorf("Невозможно установить значение с типом %s в поле с типом %s (алиас %s)", valueType, ftype.String(), aliasName)
}
