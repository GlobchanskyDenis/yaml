package yaml

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"reflect"
	"strconv"
	"time"
)

type Configurator struct {
	dataMap       map[string]map[string]interface{}
	lastBlockName string
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

func (this *Configurator) ParseToStruct(packStruct interface{}, blockName string) error {
	structVal := reflect.ValueOf(packStruct).Elem()
	this.lastBlockName = blockName

	blockValue, exists := this.dataMap[blockName]
	if exists == false {
		return fmt.Errorf("Блок <%s> отсутствует в конфигурационном файле", blockName)
	}
	return this.switchSetType(structVal, blockValue, structVal.Type(), "", reflect.Value{})
}

// Рекурсивная функция заполнения полей конфига
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
			if tag == "" || tag == "-" {
				continue
			}
			value_child, exist := structValue[tag]
			if exist == false {
				return fmt.Errorf("Для поля %s не задано значение (блок %s)", tag, this.lastBlockName)
			}
			if err := this.switchSetType(field.Field(i), value_child, ftype.Field(i).Type, ftype.Field(i).Tag, reflect.Value{}); err != nil {
				return fmt.Errorf("%w (поле %s)", err, tag)
			}
		}
	case reflect.Ptr:
		if value != nil {
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
			return reflect.ValueOf(string("")), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Uint:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(uint(uint64Val)), nil
		case int:
			return reflect.ValueOf(uint(typedValue)), nil
		default:
			return reflect.ValueOf(uint(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Uint64:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(uint64Val), nil
		case int:
			return reflect.ValueOf(uint64(typedValue)), nil
		default:
			return reflect.ValueOf(uint64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Uint32:
		switch typedValue := value.(type) {
		case string:
			uint64Val, err := strconv.ParseUint(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(uint32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(uint32(uint64Val)), nil
		case int:
			return reflect.ValueOf(uint32(typedValue)), nil
		default:
			return reflect.ValueOf(uint32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Int:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(int64Val), nil
		case int:
			return reflect.ValueOf(int(typedValue)), nil
		default:
			return reflect.ValueOf(int(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Int64:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(int64(int64Val)), nil
		case int:
			return reflect.ValueOf(int64(typedValue)), nil
		default:
			return reflect.ValueOf(int64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Int32:
		switch typedValue := value.(type) {
		case string:
			int64Val, err := strconv.ParseInt(typedValue, 10, 64)
			if err != nil {
				return reflect.ValueOf(int32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(int32(int64Val)), nil
		case int:
			return reflect.ValueOf(int32(typedValue)), nil
		default:
			return reflect.ValueOf(int32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Bool:
		switch typedValue := value.(type) {
		case string:
			boolVal, err := strconv.ParseBool(typedValue)
			if err != nil {
				return reflect.ValueOf(bool(false)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(bool(boolVal)), nil
		case bool:
			return reflect.ValueOf(bool(typedValue)), nil
		default:
			return reflect.ValueOf(bool(false)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Float64:
		switch typedValue := value.(type) {
		case string:
			float64Val, err := strconv.ParseFloat(typedValue, 64)
			if err != nil {
				return reflect.ValueOf(float64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(float64(float64Val)), nil
		case float64:
			return reflect.ValueOf(float64(typedValue)), nil
		default:
			return reflect.ValueOf(float64(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	case reflect.Float32:
		switch typedValue := value.(type) {
		case string:
			float64Val, err := strconv.ParseFloat(typedValue, 64)
			if err != nil {
				return reflect.ValueOf(float32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
			}
			return reflect.ValueOf(float32(float64Val)), nil
		case float64:
			return reflect.ValueOf(float32(typedValue)), nil
		default:
			return reflect.ValueOf(float32(0)), typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
		}
	default:
		return reflect.Value{}, typeError(ftype, fmt.Sprintf("%T", value), this.lastBlockName)
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

func typeError(ftype reflect.Type, valueType, blockName string) error {
	return fmt.Errorf("Невозможно установить значение с типом %s в поле с типом %s (блок %s)", valueType, ftype.String(), blockName)
}
