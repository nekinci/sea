#include<stdio.h>
#include <stdarg.h>
#include<stdlib.h>


int r_runtime_printf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vprintf(fmt, args);
    va_end(args);
    return r;
}

int sum(int a, int b) {
    return a +b;
}

void r_runtime_exit(int status) {
    exit(status);
}

int t() {
    int r = 11;
    return 32;
}

int r_runtime_scanf(const char* fmt, ...) {
    va_list args;
    va_start(args, fmt);
    int r = vscanf(fmt, args);
    va_end(args);
    return r;
}
