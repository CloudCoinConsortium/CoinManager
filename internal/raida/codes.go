package raida


const (
 // RESPONSE_STATUS_SUCCESS = 0xF0
  RESPONSE_STATUS_INVALID_FRAME_COUNT = 0x0F
  RESPONSE_STATUS_INVALID_END_COMMAND = 0x21


  RESPONSE_STATUS_ALL_PASS = 0xF1
  RESPONSE_STATUS_ALL_FAIL = 0xF2
  RESPONSE_STATUS_MIX = 0xF3


  RESPONSE_STATUS_ECHO_SUCCESS = 0x28
  RESPONSE_STATUS_SUCCESS = 0xFA
  RESPONSE_STATUS_FAIL = 0xFB


  RESPONSE_STATUS_FAILED_AUTH = 0x40

  RESPONSE_STATUS_WORKING = 0x60

  RESPONSE_STATUS_PAGE_NOT_FOUND = 0x42

  RESPONSE_STATUS_NO_STATEMENTS = 0x78
  RESPONSE_STATUS_STATEMENTS_DELETED = 0xC3
)


const (
  COMMAND_POWN = 0x0
  COMMAND_DETECT = 0x1
  COMMAND_ECHO = 0x4
  COMMAND_SHOW_CHANGE = 0x74
  COMMAND_BREAK = 0x7A
  COMMAND_DEPOSIT = 0x64
  COMMAND_FREECOIN = 0x1E
  COMMAND_GETTICKET = 0x0B
  COMMAND_BALANCE = 0x6E
  COMMAND_SHOW_REGISTRY = 0x70
  COMMAND_FIX = 0x3
  COMMAND_VERSION = 0xF
  COMMAND_PANG = 0x15
  COMMAND_SHOW_STATEMENT = 0x82
  COMMAND_SHOW_PAYMENT = 0x84
  COMMAND_TRANSFER = 0x6C
  COMMAND_WITHDRAW = 0x68
  COMMAND_BREAK_IN_BANK = 0x7B
  COMMAND_SYNC_OWNER_ADD = 0x96
  COMMAND_SYNC_OWNER_DELETE = 0x98
  COMMAND_DELETE_ALL_STATEMENTS = 0x83
  COMMAND_SHOW_COINS_BY_DENOMINATION = 0x72
  COMMAND_FIX_V2 = 0x32
)

var CommandNames map[uint16]string = map[uint16]string{
  COMMAND_POWN : "pown",
  COMMAND_ECHO : "echo",
  COMMAND_DETECT: "detect",
  COMMAND_SHOW_CHANGE: "show_change",
  COMMAND_DEPOSIT: "deposit",
  COMMAND_BREAK: "break",
  COMMAND_FREECOIN: "freeid",
  COMMAND_GETTICKET: "getticket",
  COMMAND_BALANCE: "balance",
  COMMAND_SHOW_REGISTRY: "show_registry",
  COMMAND_FIX: "fix",
  COMMAND_VERSION: "version",
  COMMAND_SHOW_STATEMENT: "show_statement",
  COMMAND_SHOW_PAYMENT: "show_payment",
  COMMAND_TRANSFER: "transfer",
  COMMAND_WITHDRAW: "withdraw",
  COMMAND_BREAK_IN_BANK: "break_in_bank",
  COMMAND_SYNC_OWNER_ADD: "sync_owner_add",
  COMMAND_SYNC_OWNER_DELETE: "sync_owner_delete",
  COMMAND_DELETE_ALL_STATEMENTS: "delete_all_statements",
  COMMAND_SHOW_COINS_BY_DENOMINATION: "show_coins_by_denomination",
  COMMAND_FIX_V2: "fix2",
  COMMAND_PANG: "pang",
}


func (r *RAIDA) GetCommandName(code uint16) (string, bool) {
	commandName, ok := CommandNames[code]

  return commandName, ok
}
