// data
JUMP 13
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
SETI 2
// load A with value at addr X
LDAI 0
OUTA
INCI
CMPI 13
JNEQ 15
halt
// and we're done looping