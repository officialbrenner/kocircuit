package symbol

import (
	"reflect"

	. "github.com/kocircuit/kocircuit/lang/circuit/model"
	"github.com/kocircuit/kocircuit/lang/go/gate"
	. "github.com/kocircuit/kocircuit/lang/go/kit/tree"
)

func Deconstruct(span *Span, v reflect.Value) (Symbol, error) {
	ctx := &typingCtx{Span: span}
	return ctx.Deconstruct(v)
}

func (ctx *typingCtx) Deconstruct(v reflect.Value) (Symbol, error) {
	if v.IsValid() {
		if typeName := TypeName(v.Type()); typeName != "" {
			return &NamedSymbol{Value: v}, nil
		}
	}
	return ctx.DeconstructKind(v)
}

func (ctx *typingCtx) DeconstructKind(v reflect.Value) (Symbol, error) {
	switch v.Kind() {
	case reflect.Invalid:
		return EmptySymbol{}, nil
	case reflect.String: // string
		return BasicSymbol{string(v.String())}, nil
	case reflect.Bool: // bool
		return BasicSymbol{bool(v.Bool())}, nil
	case reflect.Int: // signed integers
		return BasicSymbol{int64(v.Int())}, nil
	case reflect.Int8:
		return BasicSymbol{int8(v.Int())}, nil
	case reflect.Int16:
		return BasicSymbol{int16(v.Int())}, nil
	case reflect.Int32:
		return BasicSymbol{int32(v.Int())}, nil
	case reflect.Int64:
		return BasicSymbol{int64(v.Int())}, nil
	case reflect.Uint: // unsigned integers
		return BasicSymbol{uint64(v.Uint())}, nil
	case reflect.Uint8:
		return BasicSymbol{uint8(v.Uint())}, nil
	case reflect.Uint16:
		return BasicSymbol{uint16(v.Uint())}, nil
	case reflect.Uint32:
		return BasicSymbol{uint32(v.Uint())}, nil
	case reflect.Uint64:
		return BasicSymbol{uint64(v.Uint())}, nil
	case reflect.Float32: // floating point
		return BasicSymbol{float32(v.Float())}, nil
	case reflect.Float64:
		return BasicSymbol{float64(v.Float())}, nil
	case reflect.Uintptr:
		return nil, ctx.Errorf(nil, "go uintptr type not supported")
	case reflect.Complex64:
		return nil, ctx.Errorf(nil, "go complex64 type not supported")
	case reflect.Complex128:
		return nil, ctx.Errorf(nil, "go complex128 type not supported")
	case reflect.Array:
		return nil, ctx.Errorf(nil, "go array type not supported")
	case reflect.Chan:
		return nil, ctx.Errorf(nil, "go chan type not supported")
	case reflect.UnsafePointer:
		return nil, ctx.Errorf(nil, "go unsafe pointer type not supported")
	case reflect.Func:
		return nil, ctx.Errorf(nil, "go func type not supported")
	case reflect.Map:
		return nil, ctx.Errorf(nil, "go map type not supported")
	case reflect.Interface:
		return ctx.DeconstructInterface(v)
	case reflect.Ptr:
		if v.IsNil() {
			return EmptySymbol{}, nil
		} else {
			return ctx.Deconstruct(v.Elem())
		}
	case reflect.Slice:
		return ctx.DeconstructSlice(v)
	case reflect.Struct:
		return ctx.DeconstructStruct(v)
	}
	panic("o")
}

func (ctx *typingCtx) DeconstructInterface(v reflect.Value) (Symbol, error) {
	return &OpaqueSymbol{Value: v}, nil
}

func (ctx *typingCtx) DeconstructSlice(v reflect.Value) (Symbol, error) {
	ds := make(Symbols, 0, v.Len())
	dt := make(Types, 0, v.Len())
	ctx2 := ctx.Refine("()")
	for i := 0; i < v.Len(); i++ {
		if elem, err := ctx2.Deconstruct(v.Index(i)); err != nil {
			return nil, err
		} else {
			if !IsEmptySymbol(elem) {
				ds = append(ds, elem)
				dt = append(dt, elem.Type())
			}
		}
	}
	if len(ds) == 0 {
		return EmptySymbol{}, nil
	} else {
		if unified, err := ctx.UnifyTypes(dt); err != nil {
			panic("o")
		} else {
			return &SeriesSymbol{
				Type_: &SeriesType{unified},
				Elem:  ds,
			}, nil
		}
	}
}

func (ctx *typingCtx) DeconstructStruct(v reflect.Value) (Symbol, error) {
	fields := make(FieldSymbols, 0, v.NumField())
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		name, hasKoName := gate.StructFieldKoName(t.Field(i))
		if !hasKoName {
			continue // skip
		}
		monadic := gate.IsStructFieldMonadic(t.Field(i))
		if value, err := ctx.Refine(name).Deconstruct(v.Field(i)); err != nil {
			return nil, err
		} else {
			fields = append(fields,
				&FieldSymbol{Name: name, Monadic: monadic, Value: value},
			)
		}
	}
	return MakeStructSymbol(fields), nil
}
