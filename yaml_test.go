package yaml

import (
	"testing"
)

func TestYamlConfigurator(t *testing.T) {
	/*	Требуется присутствие файла sample_test.yaml текущей директории  */
	t.Run("read file", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.ReadFile("sample_test.yaml"); err != nil {
			t.Errorf("Error while reading file: %s", err)
			t.FailNow()
		}
	})

	t.Run("TypedPrimitives", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.setNewSource([]byte(`
            TypedPrimitives:
                Int64Val: !!int64 100500
                Int64PtrVal: !!int64 100500
                Int32Val: !!int32 100500
                Int32PtrVal: !!int32 100500
                IntVal: !!int 100500
                IntPtrVal: !!int 100500
                Float64Val: !!float64 3.1415
                Float64PtrVal: !!float64 3.1415
                Float32Val: !!float32 3.1415
                Float32PtrVal: !!float32 3.1415
                Uint64Val: !!uint64 42
                Uint64PtrVal: !!uint64 42
                Uint32Val: !!uint32 42
                Uint32PtrVal: !!uint32 42
                UintVal: !!uint 42
                UintPtrVal: !!uint 42
                StringVal: !!string "some string"
                StringPtrVal: !!string "some string"
                BoolVal: !!bool true
                BoolPtrVal: !!bool true
        `)); err != nil {
			t.Errorf("Error while reading source yaml: %s", err)
			t.FailNow()
		}
		type TypedPrimitivesType struct {
			Int64Val      int64    `conf:"Int64Val"`
			Int64PtrVal   *int64   `conf:"Int64PtrVal"`
			Int32Val      int32    `conf:"Int32Val"`
			Int32PtrVal   *int32   `conf:"Int32PtrVal"`
			IntVal        int      `conf:"IntVal"`
			IntPtrVal     *int     `conf:"IntPtrVal"`
			Float64Val    float64  `conf:"Float64Val"`
			Float64PtrVal *float64 `conf:"Float64PtrVal"`
			Float32Val    float32  `conf:"Float32Val"`
			Float32PtrVal *float32 `conf:"Float32PtrVal"`
			Uint64Val     uint64   `conf:"Uint64Val"`
			Uint64PtrVal  *uint64  `conf:"Uint64PtrVal"`
			Uint32Val     uint32   `conf:"Uint32Val"`
			Uint32PtrVal  *uint32  `conf:"Uint32PtrVal"`
			UintVal       uint     `conf:"UintVal"`
			UintPtrVal    *uint    `conf:"UintPtrVal"`
			StringVal     string   `conf:"StringVal"`
			StringPtrVal  *string  `conf:"StringPtrVal"`
			BoolVal       bool     `conf:"BoolVal"`
			BoolPtrVal    *bool    `conf:"BoolPtrVal"`
		}

		var dto TypedPrimitivesType
		if err := config.ParseToStruct(&dto, "TypedPrimitives"); err != nil {
			t.Errorf("Error while filling config: %s", err)
			t.FailNow()
		}

		if dto.Int64Val != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Int64Val", 100500, dto.Int64Val)
		}
		if dto.Int64PtrVal == nil || *dto.Int64PtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Int64PtrVal", 100500, dto.Int64PtrVal)
		}

		if dto.Int32Val != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Int32Val", 100500, dto.Int32Val)
		}
		if dto.Int32PtrVal == nil || *dto.Int32PtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Int32PtrVal", 100500, dto.Int32PtrVal)
		}

		if dto.IntVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.IntVal", 100500, dto.IntVal)
		}
		if dto.IntPtrVal == nil || *dto.IntPtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.IntPtrVal", 100500, dto.IntPtrVal)
		}

		if dto.Float64Val != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %f", "dto.Float64Val", 3.1415, dto.Float64Val)
		}
		if dto.Float64PtrVal == nil || *dto.Float64PtrVal != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %#v", "dto.Float64PtrVal", 3.1415, dto.Float64PtrVal)
		}

		if dto.Float32Val != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %f", "dto.Float32Val", 3.1415, dto.Float32Val)
		}
		if dto.Float32PtrVal == nil || *dto.Float32PtrVal != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %#v", "dto.Float32PtrVal", 3.1415, dto.Float32PtrVal)
		}

		if dto.Uint64Val != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Uint64Val", 42, dto.Uint64Val)
		}
		if dto.Uint64PtrVal == nil || *dto.Uint64PtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Uint64PtrVal", 42, dto.Uint64PtrVal)
		}

		if dto.Uint32Val != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Uint32Val", 42, dto.Uint32Val)
		}
		if dto.Uint32PtrVal == nil || *dto.Uint32PtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Uint32PtrVal", 42, dto.Uint32PtrVal)
		}

		if dto.UintVal != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.UintVal", 42, dto.UintVal)
		}
		if dto.UintPtrVal == nil || *dto.UintPtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.UintPtrVal", 42, dto.UintPtrVal)
		}

		if dto.BoolVal != true {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.BoolVal", true, dto.BoolVal)
		}
		if dto.BoolPtrVal == nil || *dto.BoolPtrVal != true {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.BoolPtrVal", true, dto.BoolPtrVal)
		}

		if dto.StringVal != "some string" {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.StringVal", "some string", dto.StringVal)
		}
		if dto.StringPtrVal == nil || *dto.StringPtrVal != "some string" {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.StringPtrVal", "some string", dto.StringPtrVal)
		}
	})

	t.Run("NotTypedPrimitives", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.setNewSource([]byte(`
            NotTypedPrimitives:
                Int64Val: 100500
                Int64PtrVal: 100500
                Int64NilVal: null
                Int32Val: 100500
                Int32PtrVal: 100500
                Int32NilVal: null
                IntVal: 100500
                IntPtrVal: 100500
                IntNilVal: null
                Float64Val: 3.1415
                Float64PtrVal: 3.1415
                Float64NilVal: null
                Float32Val: 3.1415
                Float32PtrVal: 3.1415
                Float32NilVal: null
                Uint64Val: 42
                Uint64PtrVal: 42
                Uint64NilVal: null
                Uint32Val: 42
                Uint32PtrVal: 42
                Uint32NilVal: null
                UintVal: 42
                UintPtrVal: 42
                UintNilVal: null
                StringVal: "some string"
                StringPtrVal: "some string"
                StringNilVal: null
                BoolVal: true
                BoolPtrVal: true
                BoolNilVal: null
        `)); err != nil {
			t.Errorf("Error while reading source yaml: %s", err)
			t.FailNow()
		}
		type NotTypedPrimitivesType struct {
			Int64Val      int64    `conf:"Int64Val"`
			Int64PtrVal   *int64   `conf:"Int64PtrVal"`
			Int64NilVal   *int64   `conf:"Int64NilVal"`
			Int32Val      int32    `conf:"Int32Val"`
			Int32PtrVal   *int32   `conf:"Int32PtrVal"`
			Int32NilVal   *int32   `conf:"Int32NilVal"`
			IntVal        int      `conf:"IntVal"`
			IntPtrVal     *int     `conf:"IntPtrVal"`
			IntNilVal     *int     `conf:"IntNilVal"`
			Float64Val    float64  `conf:"Float64Val"`
			Float64PtrVal *float64 `conf:"Float64PtrVal"`
			Float64NilVal *float64 `conf:"Float64NilVal"`
			Float32Val    float32  `conf:"Float32Val"`
			Float32PtrVal *float32 `conf:"Float32PtrVal"`
			Float32NilVal *float32 `conf:"Float32NilVal"`
			Uint64Val     uint64   `conf:"Uint64Val"`
			Uint64PtrVal  *uint64  `conf:"Uint64PtrVal"`
			Uint64NilVal  *uint64  `conf:"Uint64NilVal"`
			Uint32Val     uint32   `conf:"Uint32Val"`
			Uint32PtrVal  *uint32  `conf:"Uint32PtrVal"`
			Uint32NilVal  *uint32  `conf:"Uint32NilVal"`
			UintVal       uint     `conf:"UintVal"`
			UintPtrVal    *uint    `conf:"UintPtrVal"`
			UintNilVal    *uint    `conf:"UintNilVal"`
			StringVal     string   `conf:"StringVal"`
			StringPtrVal  *string  `conf:"StringPtrVal"`
			StringNilVal  *string  `conf:"StringNilVal"`
			BoolVal       bool     `conf:"BoolVal"`
			BoolPtrVal    *bool    `conf:"BoolPtrVal"`
			BoolNilVal    *bool    `conf:"BoolNilVal"`
		}

		var dto NotTypedPrimitivesType
		if err := config.ParseToStruct(&dto, "NotTypedPrimitives"); err != nil {
			t.Errorf("Error while filling config: %s", err)
			t.FailNow()
		}

		if dto.Int64Val != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Int64Val", 100500, dto.Int64Val)
		}
		if dto.Int64PtrVal == nil || *dto.Int64PtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Int64PtrVal", 100500, dto.Int64PtrVal)
		}
		if dto.Int64NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Int64NilVal", nil, dto.Int64NilVal)
		}

		if dto.Int32Val != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Int32Val", 100500, dto.Int32Val)
		}
		if dto.Int32PtrVal == nil || *dto.Int32PtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Int32PtrVal", 100500, dto.Int32PtrVal)
		}
		if dto.Int32NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Int32NilVal", nil, dto.Int32NilVal)
		}

		if dto.IntVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.IntVal", 100500, dto.IntVal)
		}
		if dto.IntPtrVal == nil || *dto.IntPtrVal != 100500 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.IntPtrVal", 100500, dto.IntPtrVal)
		}
		if dto.IntNilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.IntNilVal", nil, dto.IntNilVal)
		}

		if dto.Float64Val != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %f", "dto.Float64Val", 3.1415, dto.Float64Val)
		}
		if dto.Float64PtrVal == nil || *dto.Float64PtrVal != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %#v", "dto.Float64PtrVal", 3.1415, dto.Float64PtrVal)
		}
		if dto.Float64NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Float64NilVal", nil, dto.Float64NilVal)
		}

		if dto.Float32Val != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %f", "dto.Float32Val", 3.1415, dto.Float32Val)
		}
		if dto.Float32PtrVal == nil || *dto.Float32PtrVal != 3.1415 {
			t.Errorf("Fail: field %s expected %f got %#v", "dto.Float32PtrVal", 3.1415, dto.Float32PtrVal)
		}
		if dto.Float32NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Float32NilVal", nil, dto.Float32NilVal)
		}

		if dto.Uint64Val != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Uint64Val", 42, dto.Uint64Val)
		}
		if dto.Uint64PtrVal == nil || *dto.Uint64PtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Uint64PtrVal", 42, dto.Uint64PtrVal)
		}
		if dto.Uint64NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Uint64NilVal", nil, dto.Uint64NilVal)
		}

		if dto.Uint32Val != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.Uint32Val", 42, dto.Uint32Val)
		}
		if dto.Uint32PtrVal == nil || *dto.Uint32PtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.Uint32PtrVal", 42, dto.Uint32PtrVal)
		}
		if dto.Uint32NilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.Uint32NilVal", nil, dto.Uint32NilVal)
		}

		if dto.UintVal != 42 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.UintVal", 42, dto.UintVal)
		}
		if dto.UintPtrVal == nil || *dto.UintPtrVal != 42 {
			t.Errorf("Fail: field %s expected %d got %#v", "dto.UintPtrVal", 42, dto.UintPtrVal)
		}
		if dto.UintNilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.UintNilVal", nil, dto.UintNilVal)
		}

		if dto.BoolVal != true {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.BoolVal", true, dto.BoolVal)
		}
		if dto.BoolPtrVal == nil || *dto.BoolPtrVal != true {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.BoolPtrVal", true, dto.BoolPtrVal)
		}
		if dto.BoolNilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.BoolNilVal", nil, dto.BoolNilVal)
		}

		if dto.StringVal != "some string" {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.StringVal", "some string", dto.StringVal)
		}
		if dto.StringPtrVal == nil || *dto.StringPtrVal != "some string" {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.StringPtrVal", "some string", dto.StringPtrVal)
		}
		if dto.StringNilVal != nil {
			t.Errorf("Fail: field %s expected %#v got %#v", "dto.StringNilVal", nil, dto.StringNilVal)
		}
	})

	t.Run("InternalSlice", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.setNewSource([]byte(`
            InternalSlice:
                SliceInt:
                - 1
                - 2
                - 3
                SliceUint:
                - 1
                - 2
                - 3
                SliceString:
                - 1
                - 3.1415
                - three
                - true
        `)); err != nil {
			t.Errorf("Error while reading source yaml: %s", err)
			t.FailNow()
		}
		type SliceType struct {
			SliceInt    []int    `conf:"SliceInt"`
			SliceUint   []uint   `conf:"SliceUint"`
			SliceString []string `conf:"SliceString"`
		}

		var dto SliceType
		if err := config.ParseToStruct(&dto, "InternalSlice"); err != nil {
			t.Errorf("Error while filling config: %s", err)
			t.FailNow()
		}

		if len(dto.SliceInt) != 3 {
			t.Errorf("Fail: field %s expected %d elements got %d elements", "dto.SliceInt", 3, len(dto.SliceInt))
		}
		if len(dto.SliceUint) != 3 {
			t.Errorf("Fail: field %s expected %d elements got %d elements", "dto.SliceUint", 3, len(dto.SliceUint))
		}
		if len(dto.SliceString) != 4 {
			t.Errorf("Fail: field %s expected %d elements got %d elements", "dto.SliceString", 4, len(dto.SliceString))
		}
	})

	t.Run("NotExistingFields", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.setNewSource([]byte(`
            TypedPrimitives:
                Int64Val: !!int64 100500
                Int64PtrVal: !!int64 100500

            InternalSlice:
                SliceInt:
                - 1
                - 2
                - 3
                SliceUint:
                - 1
                - 2
                - 3
        `)); err != nil {
			t.Errorf("Error while reading source yaml: %s", err)
			t.FailNow()
		}
		/*	Данное поле (SliceNotExist) не должно существовать в yaml файле  */
		type TypedPrimitivesType struct {
			NotExistValue int64 `conf:"NotExistValue"`
		}
		if err := config.ParseToStruct(&TypedPrimitivesType{}, "TypedPrimitives"); err == nil {
			t.Errorf("Fail, not existing value not prevent to error, field NotExistValue in TypedPrimitivesType")
		} else {
			t.Logf("Success. %s", err)
		}

		/*	Данное поле (SliceNotExist) не должно существовать в yaml файле  */
		type SliceThatNotExist struct {
			SliceNotExist []int `conf:"SliceNotExist"`
		}
		if err := config.ParseToStruct(&SliceThatNotExist{}, "InternalSlice"); err == nil {
			t.Errorf("Fail, not existing slice not prevent to error, field SliceNotExist in InternalSlice")
		} else {
			t.Logf("Success. %s", err)
		}
	})

	t.Run("Fields without tags", func(t *testing.T) {
		config := NewConfigurator()
		if err := config.setNewSource([]byte(`
            BlockAlias:
                ValByTag: 100500
                ValWithoutTag: 42
        `)); err != nil {
			t.Errorf("Error while reading source yaml: %s", err)
			t.FailNow()
		}
		/*	Данное поле (SliceNotExist) не должно существовать в yaml файле  */
		type BlockAliasType struct {
			ValByTag      int `conf:"ValByTag"`
			ValWithoutTag int
		}
		var dto BlockAliasType
		dto.ValWithoutTag = 21
		if err := config.ParseToStruct(&dto, "BlockAlias"); err != nil {
			t.Errorf("Error while filling config: %s", err)
			t.FailNow()
		}

		if dto.ValByTag != 100500 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.ValByTag", 100500, dto.ValByTag)
		}
		if dto.ValWithoutTag != 21 {
			t.Errorf("Fail: field %s expected %d got %d", "dto.ValWithoutTag", 21, dto.ValWithoutTag)
		}
	})
}
