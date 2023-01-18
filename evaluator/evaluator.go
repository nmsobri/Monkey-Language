package evaluator

import (
	"Monkey/ast"
	"Monkey/object"
	"fmt"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	case *ast.Program:
		return evalProgram(node.Statements, env)

	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)

		// Prevent error object being pass around.. If its error, return immdediately
		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := Eval(node.Left, env)

		// Prevent error object being pass around.. If its error, return immdediately
		if isError(left) {
			return left
		}

		right := Eval(node.Right, env)

		// Prevent error object being pass around.. If its error, return immdediately
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.ReturnStatement:
		// Evaluate the return value expression
		val := Eval(node.ReturnValue, env)

		// Prevent error object being pass around.. If its error, return immdediately
		if isError(val) {
			return val
		}

		// Wrap inside this `object.ReturnValue` object so we can check later to determine wether to stop execution or not
		// We check in `evalStatements`
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := Eval(node.Value, env)

		if isError(val) {
			return val
		}

		env.Set(node.Name.Value, val)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}

	// let add = fn(x + y) { return x + y; }
	// add(1,2)
	case *ast.CallExpression:
		fn := Eval(node.Function, env) // This will return `object.Function` if there is no error

		if isError(fn) {
			return fn
		}

		args := evalExpressions(node.Arguments, env)

		// Stop executing if got any error with the params
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(fn, args)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.ArrayLiteral:
		elements := evalExpressions(node.Elements, env)

		if len(elements) == 1 && isError(elements[0]) {
			return elements[0] // If there is an error, return an error object
		}

		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := Eval(node.Left, env)

		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)

		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)

	case *ast.AssignmentExpression:
		val := Eval(node.Value, env)

		if isError(val) {
			return val
		}

		if !env.IsKey(node.Name.Value) {
			return newError("identifier not found `%s`", node.Name.Value)
		}

		env.Set(node.Name.Value, val)
		return nil

	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}

	return nil
}

func evalProgram(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range statements {
		result = Eval(stmt, env)

		switch result := result.(type) {
		// Check for early return statement, if found, return now!
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalStatements(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range statements {
		result = Eval(stmt, env)
		// Just check if this is `object.ReturnValue`, return early
		// but dont unwrap it, else, early return wouldnt be possible
		// cause its type already change to whatever wrapped value that
		// the `object.ReturnValue` contain and cant be detected in `evalProgram`

		// `object.ERROR_OBJ` just to terminate execution in case there is any error

		if result != nil {
			resultType := result.Type()

			if resultType == object.RETURN_VALUE_OBJ || resultType == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperator(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperator(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	case operator == "==": // For boolean comoparison
		return nativeBoolToBooleanObject(left == right) // Pointer comparison

	case operator == "!=": // For boolean comoparison
		return nativeBoolToBooleanObject(left != right) // Pointer comparison

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}

	case "-":
		return &object.Integer{Value: leftVal - rightVal}

	case "*":
		return &object.Integer{Value: leftVal * rightVal}

	case "/":
		return &object.Integer{Value: leftVal / rightVal}

	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)

	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)

	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)

	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)

	// Prevent error object being pass around.. If its error, return immdediately
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	// First search the identifier in current environment and its outer environment and etc
	// If its still not found, try search from builtins, if still not found, return and error
	// indicating identifier is not found

	obj, ok := env.Get(node.Value)

	if ok {
		return obj
	}

	obj, ok = builtins[node.Value]

	if ok {
		return obj
	}

	return newError("identifier not found: " + node.Value)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	args := []object.Object{}

	for _, arg := range exps {
		evaluated := Eval(arg, env)

		if isError(evaluated) {
			args = append(args, evaluated)
			return args
		}

		args = append(args, evaluated)
	}

	return args
}

func applyFunction(_fn object.Object, args []object.Object) object.Object {

	// Build function params
	// Cannot used top level environment cause in Monkey,
	// `function` can be nested at arbitarily depth and
	// each function depth have its own environment

	// let add = fn(x,y) { return x + y}
	// add (1, 2)

	// let newAdder = fn(x) { <--- New environment ( extend global environment )
	// 	  fn(y) { <--- New environment ( extend its parent environment )
	//      x + y;
	//    }
	// }
	//
	// let addTwo = newAdder(2);
	// addTwo(3);

	// extended function environment cause this function might be nested inside
	// another function.. ( each function have their own environment )
	switch fn := _fn.(type) {

	case *object.Function:
		extendedEnv := extendedFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		// Call directly since this builtin is `golang` code
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendedFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	// Bind function arguments to function parameters name
	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	returnVal, ok := obj.(*object.ReturnValue)

	if ok {
		return returnVal.Value
	}

	return obj
}

func evalStringInfixExpression(operator string, left object.Object, right object.Object) object.Object {
	if operator != "+" {
		return &object.Error{Message: fmt.Sprintf("unknown operator: %s %s %s", left.Type(), operator, right.Type())}
	}

	leftStr := left.(*object.String)
	rightStr := right.(*object.String)
	return &object.String{Value: leftStr.Value + rightStr.Value}
}

func evalIndexExpression(left object.Object, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(left object.Object, index object.Object) object.Object {

	arr := left.(*object.Array).Elements
	idx := index.(*object.Integer).Value
	max := len(arr) - 1

	if idx < 0 || int(idx) > max {
		return NULL
	}

	return arr[idx]
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	hash := &object.Hash{}
	pairs := make(map[object.HashKey]object.HashPair)

	for k, v := range node.Pairs {
		// Get key
		key := Eval(k, env)

		if isError(key) {
			return key
		}

		// Check if this key implement `Hashable` interface
		hashKey, ok := key.(object.Hashable)

		if !ok {
			return newError("unusable as hash key %s", key.Type())
		}

		val := Eval(v, env)

		if isError(val) {
			return val
		}

		pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: val}
	}

	hash.Pairs = pairs
	return hash
}

func evalHashIndexExpression(left object.Object, index object.Object) object.Object {
	hash := left.(*object.Hash)

	key, ok := index.(object.Hashable)

	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	hashPair, ok := hash.Pairs[key.HashKey()]

	if !ok {
		return NULL
	}

	return hashPair.Value
}
