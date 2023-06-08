// data
JUMP 13
72
101
108
108
111
32
87
111
114
108
100

// code
SETI 2
// load A with value at addr X
LDAI 0
OUTA
INCI
CMPI 13
JNEQ 15
halt
// and we're done looping