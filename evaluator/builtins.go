package evaluator

import (
	"Monkey/object"
	"fmt"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", len(args), 1)
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}

			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}

			default:
				return newError("argument to `len` not supported, got=%s", args[0].Type())
			}
		},
	},
	"first": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", len(args), 1)
			}

			if args[0].Type() != object.ARRAY_OBJ {
				// return newError(string(args[0].Type()))
				return newError("argument to `first` must be an ARRAY, got=%s", args[0].Type())
			}

			arr := args[0].(*object.Array)

			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}

			return NULL
		},
	},
	"last": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", len(args), 1)
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `last` must be an ARRAY, got=%s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			if length > 0 {
				return arr.Elements[length-1]
			}

			return NULL

		},
	},
	"rest": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=%d", len(args), 1)
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("argument to `rest` must be an ARRAY, got=%s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)

			if len(arr.Elements) > 0 {
				newArr := make([]object.Object, length-1)
				copy(newArr, arr.Elements[1:])
				return &object.Array{Elements: newArr}
			}

			return NULL
		},
	},
	"push": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("argument to push should be 2")
			}

			if args[0].Type() != object.ARRAY_OBJ {
				return newError("first argument to `push` must be an ARRAY, got=%s", args[0].Type())
			}

			arr := args[0].(*object.Array)
			length := len(arr.Elements)
			newArr := make([]object.Object, length+1)

			copy(newArr, arr.Elements)
			newArr[length] = args[1]

			return &object.Array{Elements: newArr}
		},
	},
	"puts": {
		Fn: func(args ...object.Object) object.Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return NULL
		},
	},
}
