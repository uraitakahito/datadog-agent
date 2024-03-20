package fxhelper

import (
	"reflect"

	"go.uber.org/fx"
)

// ProvideComponentConstructor takes as input a Component constructor function
// that uses plain (non-fx aware) structs as its argument and return value, and
// returns an fx.Provide'd Option that will properly include that Component
// into the fx constructor graph.
func ProvideComponentConstructor(ourConstructorFunc interface{}) fx.Option {
	// TODO: Validate that constructor is a function, with 1 input and 1 output

	// look at input argument and return value for our constructor
	ctorFuncVal := reflect.TypeOf(ourConstructorFunc)
	ctorInType := ctorFuncVal.In(0)
	ctorOutType := ctorFuncVal.Out(0)

	// create types that have fx-aware meta-fields
	// these are used to construct a function that can build the fx graph
	inFxType := constructFxInType(ctorInType)
	outFxType := constructFxOutType(ctorOutType)
	funcFxType := reflect.FuncOf([]reflect.Type{inFxType}, []reflect.Type{outFxType}, false)

	body := func(fxAwareDeps reflect.Value) interface{} {
		ctorValue := reflect.ValueOf(ourConstructorFunc)
		// get dependencies without fx.In
		myPlainDeps := constructPlainDeps(fxAwareDeps)
		ctorRes := ctorValue.Call([]reflect.Value{myPlainDeps})
		// TODO: assuming return value has only 1 element
		return fillFxProvideStruct(ctorRes[0], outFxType)
	}

	// construct a function value that will instruct fx what the Components are
	makeFuncValue := reflect.MakeFunc(funcFxType, func(args []reflect.Value) []reflect.Value {
		deps := args[0]
		res := body(deps)
		rval := reflect.ValueOf(res)
		return []reflect.Value{rval}
	})

	makeFunc := makeFuncValue.Interface()
	return fx.Provide(makeFunc)
}

// convert dependencies from fx-aware deps into a plain struct that
// can be used to construct the Component
func constructPlainDeps(fxAwareValue reflect.Value) reflect.Value {

	// TODO: assert that deps is a Struct, or else we panic!
	fxAwareType := fxAwareValue.Type()

	ourStructNumField := fxAwareValue.NumField() - 1

	// create an anonymous struct that matches our input,
	// except it removes the embedded "fx.In"
	newFields := make([]reflect.StructField, 0, ourStructNumField)
	for i := 0; i < fxAwareType.NumField(); i++ {
		// TODO: this is not strict enough, could get false positives
		if fxAwareType.Field(i).Name == "In" {
			continue
		}
		newF := reflect.StructField{
			Type: fxAwareType.Field(i).Type,
			Name: fxAwareType.Field(i).Name,
		}
		newFields = append(newFields, newF)
	}

	makeType := reflect.StructOf(newFields)
	makeResult := reflect.New(makeType).Elem()
	j := 0
	for i := 0; i < fxAwareType.NumField(); i++ {
		// TODO: this is not strict enough, could get false positives
		if fxAwareType.Field(i).Name == "In" {
			continue
		}
		makeResult.Field(j).Set(fxAwareValue.Field(i))
		j++
	}

	return makeResult
}

func constructFxInType(plainType reflect.Type) reflect.Type {
	return constructFxAwareStruct(plainType, false)
}

func constructFxOutType(plainType reflect.Type) reflect.Type {
	return constructFxAwareStruct(plainType, true)
}

func constructFxAwareStruct(plainType reflect.Type, isOut bool) reflect.Type {
	// create an anonymous struct that matches our input,
	// except it also has "fx.In" / "fx.Out" embedded
	fields := make([]reflect.StructField, 0, plainType.NumField()+1)
	var metaField reflect.StructField
	if isOut {
		metaField = reflect.StructField{Name: "Out", Type: reflect.TypeOf(fx.Out{}), Anonymous: true}
	} else {
		metaField = reflect.StructField{Name: "In", Type: reflect.TypeOf(fx.In{}), Anonymous: true}
	}
	fields = append(fields, metaField)
	for i := 0; i < plainType.NumField(); i++ {
		f := reflect.StructField{Type: plainType.Field(i).Type, Name: plainType.Field(i).Name}
		fields = append(fields, f)
	}
	return reflect.StructOf(fields)
}

func fillFxProvideStruct(provides reflect.Value, outFxType reflect.Type) interface{} {
	// `provides` argument is the fx-free struct returned by the Component's constructor
	// NOTE: assumes fxAwareResult[0] is the `fx.Out` embedded field
	fxAwareResult := reflect.New(outFxType).Elem()
	for i := 0; i < provides.NumField(); i++ {
		fxAwareResult.Field(i + 1).Set(provides.Field(i))
	}
	return fxAwareResult.Interface()
}
