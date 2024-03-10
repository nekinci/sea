#!/bin/sh
llc -filetype=obj plus.ll -o plus.o
llc -filetype=obj ./runtime/runtime.ll -o runtime.o
clang plus.o runtime.o -o plus
exec ./plus

