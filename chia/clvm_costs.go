package chia

const IF_COST = 33
const CONS_COST = 50
const FIRST_COST = 30
const REST_COST = 30
const LISTP_COST = 19

const MALLOC_COST_PER_BYTE = 10

const ARITH_BASE_COST = 99
const ARITH_COST_PER_BYTE = 3
const ARITH_COST_PER_ARG = 320

const LOG_BASE_COST = 100
const LOG_COST_PER_BYTE = 3
const LOG_COST_PER_ARG = 264

const GRS_BASE_COST = 117
const GRS_COST_PER_BYTE = 1

const EQ_BASE_COST = 117
const EQ_COST_PER_BYTE = 1

const GR_BASE_COST = 498
const GR_COST_PER_BYTE = 2

const DIVMOD_BASE_COST = 1116
const DIVMOD_COST_PER_BYTE = 6

const DIV_BASE_COST = 988
const DIV_COST_PER_BYTE = 4

const SHA256_BASE_COST = 87
const SHA256_COST_PER_ARG = 134
const SHA256_COST_PER_BYTE = 2

const POINT_ADD_BASE_COST = 101094
const POINT_ADD_COST_PER_ARG = 1343980

const PUBKEY_BASE_COST = 1325730
const PUBKEY_COST_PER_BYTE = 38

const MUL_BASE_COST = 92
const MUL_COST_PER_OP = 885
const MUL_LINEAR_COST_PER_BYTE = 6
const MUL_SQUARE_COST_PER_BYTE_DIVIDER = 128

const STRLEN_BASE_COST = 173
const STRLEN_COST_PER_BYTE = 1

const PATH_LOOKUP_BASE_COST = 40
const PATH_LOOKUP_COST_PER_LEG = 4
const PATH_LOOKUP_COST_PER_ZERO_BYTE = 4

const CONCAT_BASE_COST = 142
const CONCAT_COST_PER_ARG = 135
const CONCAT_COST_PER_BYTE = 3

const BOOL_BASE_COST = 200
const BOOL_COST_PER_ARG = 300

const ASHIFT_BASE_COST = 596
const ASHIFT_COST_PER_BYTE = 3

const LSHIFT_BASE_COST = 277
const LSHIFT_COST_PER_BYTE = 3

const LOGNOT_BASE_COST = 331
const LOGNOT_COST_PER_BYTE = 3

const APPLY_COST = 90
const QUOTE_COST = 20
