#include<stdio.h>
#include<stdlib.h>
#include <stdarg.h>
#include<stdlib.h>
#include <string.h>

#pragma clang diagnostic ignored "-Wvarargs"


typedef struct {
    const char* buffer;
    size_t size;
} string;


string make_string(const char* buffer) {
    size_t size = strlen(buffer);
    string result;
    result.buffer = buffer;
    result.size = size;
    return result;
}


int r_runtime_scanf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vscanf(fmt, args);
    va_end(args);
    return r;
}


int r_runtime_printf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vprintf(fmt, args);
    va_end(args);
    return r;
}

int printf_internal(string str, ...) {
    va_list args;
    va_start(args, str.buffer);
    int r = vprintf(str.buffer, args);
    va_end(args);
    return r;
}

int scanf_internal(string s, ...) {
     va_list args;
     va_start(args, s.buffer);
     int r = vscanf(s.buffer, args);
     va_end(args);
     return r;
}


void* malloc_internal(size_t s) {
    return malloc(s);
}

size_t strlen_internal(string str) {
    return str.size;
}

int sum(int a, int b) {
    return a +b;
}

void r_runtime_exit(int status) {
    exit(status);
}
