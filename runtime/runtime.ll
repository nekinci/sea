; ModuleID = './runtime/runtime.c'
source_filename = "./runtime/runtime.c"
target datalayout = "e-m:o-i64:64-i128:128-n32:64-S128"
target triple = "arm64-apple-macosx14.0.0"

%struct.string = type { ptr, i64 }

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define ptr @memcpy_internal(ptr noundef %0, ptr noundef %1) #0 {
  %3 = alloca ptr, align 8
  %4 = alloca ptr, align 8
  %5 = alloca i64, align 8
  store ptr %0, ptr %3, align 8
  store ptr %1, ptr %4, align 8
  store i64 10, ptr %5, align 8
  %6 = load ptr, ptr %3, align 8
  %7 = load ptr, ptr %4, align 8
  %8 = load i64, ptr %5, align 8
  %9 = load ptr, ptr %3, align 8
  %10 = call i64 @llvm.objectsize.i64.p0(ptr %9, i1 false, i1 true, i1 false)
  %11 = call ptr @__memcpy_chk(ptr noundef %6, ptr noundef %7, i64 noundef %8, i64 noundef %10) #7
  ret ptr %11
}

; Function Attrs: nounwind
declare ptr @__memcpy_chk(ptr noundef, ptr noundef, i64 noundef, i64 noundef) #1

; Function Attrs: nocallback nofree nosync nounwind readnone speculatable willreturn
declare i64 @llvm.objectsize.i64.p0(ptr, i1 immarg, i1 immarg, i1 immarg) #2

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define [2 x i64] @make_string(ptr noundef %0) #0 {
  %2 = alloca %struct.string, align 8
  %3 = alloca ptr, align 8
  %4 = alloca i64, align 8
  store ptr %0, ptr %3, align 8
  %5 = load ptr, ptr %3, align 8
  %6 = call i64 @strlen(ptr noundef %5)
  store i64 %6, ptr %4, align 8
  %7 = load ptr, ptr %3, align 8
  %8 = getelementptr inbounds %struct.string, ptr %2, i32 0, i32 0
  store ptr %7, ptr %8, align 8
  %9 = load i64, ptr %4, align 8
  %10 = getelementptr inbounds %struct.string, ptr %2, i32 0, i32 1
  store i64 %9, ptr %10, align 8
  %11 = load [2 x i64], ptr %2, align 8
  ret [2 x i64] %11
}

declare i64 @strlen(ptr noundef) #3

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @r_runtime_scanf(ptr noundef %0, ...) #0 {
  %2 = alloca ptr, align 8
  %3 = alloca ptr, align 8
  %4 = alloca i32, align 4
  store ptr %0, ptr %2, align 8
  call void @llvm.va_start(ptr %3)
  %5 = load ptr, ptr %2, align 8
  %6 = load ptr, ptr %3, align 8
  %7 = call i32 @vscanf(ptr noundef %5, ptr noundef %6)
  store i32 %7, ptr %4, align 4
  call void @llvm.va_end(ptr %3)
  %8 = load i32, ptr %4, align 4
  ret i32 %8
}

; Function Attrs: nocallback nofree nosync nounwind willreturn
declare void @llvm.va_start(ptr) #4

declare i32 @vscanf(ptr noundef, ptr noundef) #3

; Function Attrs: nocallback nofree nosync nounwind willreturn
declare void @llvm.va_end(ptr) #4

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @r_runtime_printf(ptr noundef %0, ...) #0 {
  %2 = alloca ptr, align 8
  %3 = alloca ptr, align 8
  %4 = alloca i32, align 4
  store ptr %0, ptr %2, align 8
  call void @llvm.va_start(ptr %3)
  %5 = load ptr, ptr %2, align 8
  %6 = load ptr, ptr %3, align 8
  %7 = call i32 @vprintf(ptr noundef %5, ptr noundef %6)
  store i32 %7, ptr %4, align 4
  call void @llvm.va_end(ptr %3)
  %8 = load i32, ptr %4, align 4
  ret i32 %8
}

declare i32 @vprintf(ptr noundef, ptr noundef) #3

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @printf_internal([2 x i64] %0, ...) #0 {
  %2 = alloca %struct.string, align 8
  %3 = alloca ptr, align 8
  %4 = alloca i32, align 4
  store [2 x i64] %0, ptr %2, align 8
  call void @llvm.va_start(ptr %3)
  %5 = getelementptr inbounds %struct.string, ptr %2, i32 0, i32 0
  %6 = load ptr, ptr %5, align 8
  %7 = load ptr, ptr %3, align 8
  %8 = call i32 @vprintf(ptr noundef %6, ptr noundef %7)
  store i32 %8, ptr %4, align 4
  call void @llvm.va_end(ptr %3)
  %9 = load i32, ptr %4, align 4
  ret i32 %9
}

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @scanf_internal([2 x i64] %0, ...) #0 {
  %2 = alloca %struct.string, align 8
  %3 = alloca ptr, align 8
  %4 = alloca i32, align 4
  store [2 x i64] %0, ptr %2, align 8
  call void @llvm.va_start(ptr %3)
  %5 = getelementptr inbounds %struct.string, ptr %2, i32 0, i32 0
  %6 = load ptr, ptr %5, align 8
  %7 = load ptr, ptr %3, align 8
  %8 = call i32 @vscanf(ptr noundef %6, ptr noundef %7)
  store i32 %8, ptr %4, align 4
  call void @llvm.va_end(ptr %3)
  %9 = load i32, ptr %4, align 4
  ret i32 %9
}

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define ptr @malloc_internal(i64 noundef %0) #0 {
  %2 = alloca i64, align 8
  store i64 %0, ptr %2, align 8
  %3 = load i64, ptr %2, align 8
  %4 = call ptr @malloc(i64 noundef %3) #8
  ret ptr %4
}

; Function Attrs: allocsize(0)
declare ptr @malloc(i64 noundef) #5

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i64 @strlen_internal([2 x i64] %0) #0 {
  %2 = alloca %struct.string, align 8
  store [2 x i64] %0, ptr %2, align 8
  %3 = getelementptr inbounds %struct.string, ptr %2, i32 0, i32 1
  %4 = load i64, ptr %3, align 8
  ret i64 %4
}

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @sum(i32 noundef %0, i32 noundef %1) #0 {
  %3 = alloca i32, align 4
  %4 = alloca i32, align 4
  store i32 %0, ptr %3, align 4
  store i32 %1, ptr %4, align 4
  %5 = load i32, ptr %3, align 4
  %6 = load i32, ptr %4, align 4
  %7 = add nsw i32 %5, %6
  ret i32 %7
}

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define void @r_runtime_exit(i32 noundef %0) #0 {
  %2 = alloca i32, align 4
  store i32 %0, ptr %2, align 4
  %3 = load i32, ptr %2, align 4
  call void @exit(i32 noundef %3) #9
  unreachable
}

; Function Attrs: noreturn
declare void @exit(i32 noundef) #6

; Function Attrs: noinline nounwind optnone ssp uwtable(sync)
define i32 @main2() #0 {
  ret i32 0
}

attributes #0 = { noinline nounwind optnone ssp uwtable(sync) "frame-pointer"="non-leaf" "min-legal-vector-width"="0" "no-trapping-math"="true" "probe-stack"="__chkstk_darwin" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+sha3,+sm4,+v8.1a,+v8.2a,+v8.3a,+v8.4a,+v8.5a,+v8a,+zcm,+zcz" }
attributes #1 = { nounwind "frame-pointer"="non-leaf" "no-trapping-math"="true" "probe-stack"="__chkstk_darwin" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+sha3,+sm4,+v8.1a,+v8.2a,+v8.3a,+v8.4a,+v8.5a,+v8a,+zcm,+zcz" }
attributes #2 = { nocallback nofree nosync nounwind readnone speculatable willreturn }
attributes #3 = { "frame-pointer"="non-leaf" "no-trapping-math"="true" "probe-stack"="__chkstk_darwin" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+sha3,+sm4,+v8.1a,+v8.2a,+v8.3a,+v8.4a,+v8.5a,+v8a,+zcm,+zcz" }
attributes #4 = { nocallback nofree nosync nounwind willreturn }
attributes #5 = { allocsize(0) "frame-pointer"="non-leaf" "no-trapping-math"="true" "probe-stack"="__chkstk_darwin" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+sha3,+sm4,+v8.1a,+v8.2a,+v8.3a,+v8.4a,+v8.5a,+v8a,+zcm,+zcz" }
attributes #6 = { noreturn "frame-pointer"="non-leaf" "no-trapping-math"="true" "probe-stack"="__chkstk_darwin" "stack-protector-buffer-size"="8" "target-cpu"="apple-m1" "target-features"="+aes,+crc,+crypto,+dotprod,+fp-armv8,+fp16fml,+fullfp16,+lse,+neon,+ras,+rcpc,+rdm,+sha2,+sha3,+sm4,+v8.1a,+v8.2a,+v8.3a,+v8.4a,+v8.5a,+v8a,+zcm,+zcz" }
attributes #7 = { nounwind }
attributes #8 = { allocsize(0) }
attributes #9 = { noreturn }

!llvm.module.flags = !{!0, !1, !2, !3, !4}
!llvm.ident = !{!5}

!0 = !{i32 2, !"SDK Version", [2 x i32] [i32 14, i32 2]}
!1 = !{i32 1, !"wchar_size", i32 4}
!2 = !{i32 8, !"PIC Level", i32 2}
!3 = !{i32 7, !"uwtable", i32 1}
!4 = !{i32 7, !"frame-pointer", i32 1}
!5 = !{!"Apple clang version 15.0.0 (clang-1500.1.0.2.5)"}
