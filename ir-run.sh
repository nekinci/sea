#!/bin/sh
llc -filetype=obj input.ll -o input.o
llc -filetype=obj ./runtime/runtime.ll -o runtime.o
clang input.o runtime.o -o input
exec ./input

