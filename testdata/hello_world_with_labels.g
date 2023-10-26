// data
JUMP main
'H'
'e'
'l'
'l'
'o'
' '
'W'
'o'
'r'
'l'
'd'

// code
main:
SETI 2
// loop
up:
// load A with value at addr X
LDAI 0
OUTA
INCI
CMPI main
JNEQ up
halt
// and we're done looping