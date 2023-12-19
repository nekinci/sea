@str = global [9 x i8] c"hello %d\0A"

declare external i32 @r_runtime_printf(i8* %0, ...)

define i32 @main() {
entry:
	%0 = add i32 0, 1
	%1 = call i32 @ExampleFunc(i32 %0, i32 1)
	%2 = call i32 (i8*, ...) @r_runtime_printf([9 x i8]* @str, i32 %1)
	%3 = alloca i32
	store i32 111, i32* %3
	%4 = load i32, i32* %3
	%5 = call i32 (i8*, ...) @r_runtime_printf([9 x i8]* @str, i32 %4)
	ret i32 %4
}

define i32 @ExampleFunc(i32 %a, i32 %b) {
entry:
	%0 = alloca i32
	store i32 992, i32* %0
	%1 = load i32, i32* %0
	%2 = add i32 %a, %1
	ret i32 %2
}
