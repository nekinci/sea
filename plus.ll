@_val = global [39 x i8] c"bu abc fonksiyonunun stringi: %d \5Cn\5Cx00"

declare external i32 @r_runtime_printf(i8* %0, ...)

define i32 @abc(i32 %a) {
entry:
	%0 = call i32 (i8*, ...) @r_runtime_printf(i8* getelementptr ([39 x i8], [39 x i8]* @_val, i8 0, i8 0), i32 %a)
	ret i32 6
}

define i32 @main() {
entry:
	%0 = mul i32 5, 2
	%1 = mul i32 %0, 2
	%2 = call i32 @abc(i32 %1)
	%3 = call i32 @abc(i32 7)
	ret i32 0
}
