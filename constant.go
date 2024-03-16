package main

type operationInfo struct {
	left  string
	op    Operation
	right string
}

var validOps map[operationInfo]string

/*
Add Operation = iota
Sub
Mul
Div
Mod
Eq
Neq
Lt
Lte
Gt
Gte
Not
And
Band
Or
Sizeof
*/

func init() {
	validOps = make(map[operationInfo]string)
	validOps[operationInfo{"*", Eq, "*"}] = "bool"
	validOps[operationInfo{"*", Neq, "*"}] = "bool"
	validOps[operationInfo{"string", Add, "string"}] = "string"
	validOps[operationInfo{"string", Add, "char"}] = "string"
	validOps[operationInfo{"char", Add, "string"}] = "string"

	validOps[operationInfo{"char", Add, "char"}] = "context_based<string,char,i8,i16,i32,i64>"
	validOps[operationInfo{"char", Sub, "char"}] = "char"
	validOps[operationInfo{"char", Gte, "char"}] = "bool"
	validOps[operationInfo{"char", Lte, "char"}] = "bool"
	validOps[operationInfo{"char", Add, "i8"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"char", Add, "i16"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"char", Add, "i32"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"char", Add, "i164"}] = "context_based<char,i8,i16,i32,i64>"

	validOps[operationInfo{"i8", Add, "char"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"i16", Add, "char"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"i32", Add, "char"}] = "context_based<char,i8,i16,i32,i64>"
	validOps[operationInfo{"i64", Add, "char"}] = "context_based<char,i8,i16,i32,i64>"

	validOps[operationInfo{"bool", Eq, "bool"}] = "bool"
	validOps[operationInfo{"bool", Neq, "bool"}] = "bool"
	validOps[operationInfo{"bool", Or, "bool"}] = "bool"
	validOps[operationInfo{"bool", And, "bool"}] = "bool"

	numericOps("i8", "i8", false)
	numericOps("i16", "i16", false)
	numericOps("i32", "i32", false)
	numericOps("i64", "i64", false)

	numericOps("f16", "f16", true)
	numericOps("f32", "f32", true)
	numericOps("f64", "f64", true)

}

func numericOps(left, right string, floatOp bool) {
	if floatOp {
		validOps[operationInfo{left, Add, right}] = left
	} else {
		validOps[operationInfo{left, Add, right}] = "context_based<char,i8,i16,i32,i64>"
	}
	validOps[operationInfo{left, Sub, right}] = left
	validOps[operationInfo{left, Mul, right}] = left
	validOps[operationInfo{left, Div, right}] = left
	validOps[operationInfo{left, Mod, right}] = left
	validOps[operationInfo{left, Eq, right}] = "bool"
	validOps[operationInfo{left, Neq, right}] = "bool"
	validOps[operationInfo{left, Lt, right}] = "bool"
	validOps[operationInfo{left, Lte, right}] = "bool"
	validOps[operationInfo{left, Gt, right}] = "bool"
	validOps[operationInfo{left, Gte, right}] = "bool"
}
