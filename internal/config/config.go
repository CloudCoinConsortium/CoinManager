package config

import (
	"os"
)

const VERSION = "0.1.127"

// For various HTTP requests (not RAIDA requests)
var HTTP_TIMEOUT = 3
var DEFAULT_DOMAIN = "73.66.181.49"

//73.66.181.49

//Ports 30,001 - 30,025


const DEFAULT_BASE_TIMEOUT = 10


// millisecods
const (
  GLOBAL_FIRST_TIMEOUT_MS = 300
  GLOBAL_NEXT_TIMEOUT_MS = 200
)

var GLOBAL_NEXT_2ND_DEVIATION_TIMEOUT_MS = 0

var ROOT_PATH = ""

// 10 minutes
const COMPLETED_TASK_LIFETIME = 600
const TASK_TIMEOUT = 600

const RAIDA_REQUEST_HEADER_SIZE = 12
const RAIDA_RESPONSE_HEADER_SIZE = 12

const LOG_LEVEL_DEBUG = 3
const LOG_LEVEL_INFO = 2
const LOG_LEVEL_ERROR = 1

const STORAGE_DRIVER = "filesystem"

const DEFAULT_WALLET_NAME = "Default"

const LISTEN_PORT = 8888

const REMOTE_RESULT_ERROR_NONE = 0
const REMOTE_RESULT_ERROR_TIMEOUT = 1
const REMOTE_RESULT_ERROR_COMMON = 2
const REMOTE_RESULT_ERROR_SKIPPED = 3


const WALLET_NAME_MIN_LENGTH = 1
const WALLET_NAME_MAX_LENGTH = 64

const SKYWALLET_NAME_MIN_LENGTH = 6
const SKYWALLET_NAME_MAX_LENGTH = 64

const PASSWORD_MIN_LENGTH = 4
const PASSWORD_MAX_LENGTH = 64


const COINID_TYPE_SKY = 0
const COINID_TYPE_CLOUD = 1

const MAX_TAG_LENGTH = 64

const TOTAL_COINS = 16777216

const MIN_FILE_LENGTH = 3
const MAX_FILE_LENGTH = 256

const WALLET_CONFIG_NAME = "config.toml"

const IMPORT_TYPE_INLINE = "inline"
const IMPORT_TYPE_FILE = "file"
const IMPORT_TYPE_SUSPECT = "suspect"



const MIN_QUORUM_COUNT = 13

var PUBLIC_CHANGE_SN = 2
var ENCRYPTION_DISABLED = false
var USE_LOCAL_RAIDAS = false


const COIN_ID1_MAIN = 0
const (
  COIN_ID2_ID = 0
  COIN_ID2_CLOUDCOIN = 1
  COIN_ID2_NFT = 2
)

const (
  TASK_STATUS_RUNNING = "running"
  TASK_STATUS_COMPLETED = "completed"
  TASK_STATUS_ERROR = "error"
)

const DEFAULT_NN = COIN_ID2_CLOUDCOIN
const DEFAULT_CLOUDID = 0
const (
  COIN_FORMAT_TYPE_STANDARD = 0
  COIN_FORMAT_TYPE_PANG = 1
  COIN_FORMAT_TYPE_STORE_IN_MIND = 2
)

const (
  COIN_ENCRYPTION_TYPE_NONE = 0
)


const PUBLIC_CHANGE_MAKER_ID = 2



var EXPORT_BACKGROUND = "#02203D"
var BRAND_COLOR = "#7FA8FF"


var CmdCli bool
var CmdNoUI bool
var CmdDebug bool
var CmdHelp bool
var CmdVersion bool
var CmdCommand string
var CmdLogfile string
var CmdUIPath string

var LogDesc *os.File

const TOTAL_RAIDA_NUMBER = 25
const NEED_VOTERS_FOR_BALANCE = 17

const (
	RAIDA_STATUS_UNTRIED = 0
	RAIDA_STATUS_PASS = 1
	RAIDA_STATUS_ERROR = 2
	RAIDA_STATUS_FAIL = 3
	RAIDA_STATUS_NORESPONSE = 4
)

const DIR_TEMPLATES = "Templates"

const (
  COIN_STATUS_UNKNOWN = iota
  COIN_STATUS_AUTHENTIC
  COIN_STATUS_FRACKED
  COIN_STATUS_COUNTERFEIT
  COIN_STATUS_LIMBO
)

const (
  COIN_LOCATION_STATUS_UNKNOWN = iota
  COIN_LOCATION_STATUS_IMPORT 
  COIN_LOCATION_STATUS_SUSPECT
  COIN_LOCATION_STATUS_BANK
  COIN_LOCATION_STATUS_FRACKED
  COIN_LOCATION_STATUS_COUNTERFEIT
  COIN_LOCATION_STATUS_SENT
  COIN_LOCATION_STATUS_LIMBO
  COIN_LOCATION_STATUS_TRASH
)

const FIX_MAX_REGEXPS = 56

const MAX_DGRAM_SIZE = 65535

const TOPDIR = "cloudcoin_manager"
const DIR_LOG = "Logs"
const DIR_BANK = "Bank"
const DIR_FRACKED = "Fracked"
const DIR_SENT = "Sent"
const DIR_TRASH = "Trash"
const DIR_COUNTERFEIT = "Counterfeit"
const DIR_ID = "ID"
const DIR_IMPORT = "Import"
const DIR_IMPORTED = "Imported"
const DIR_SUSPECT = "Suspect"
const DIR_LIMBO = "Limbo"

const DIR_RECEIPTS = "Receipts"

const TYPE_STACK = 1
const TYPE_PNG = 4


const LOG_FILENAME = "main.log"
const CONFIG_FILENAME = "config.toml"

const MAX_LOG_SIZE = 50000000

const CHANGE_METHOD_250F = 1
const CHANGE_METHOD_100E = 2
const CHANGE_METHOD_25B = 3
const CHANGE_METHOD_5A = 4

const META_ENV_SEPARATOR = "*"

//const MAX_NOTES_TO_SEND = 100
//var MAX_NOTES_TO_SEND = 400
var MAX_NOTES_TO_SEND = 40

const MIN_PASSED_NUM_TO_BE_AUTHENTIC = 14
const MAX_FAILED_NUM_TO_BE_COUNTERFEIT = 12


const EXPORT_TYPE_PNG = "png"
const EXPORT_TYPE_ZIP = "zip"
const EXPORT_TYPE_BIN = "bin"

const GUARDIAN_HTTP_TIMEOUT = 3

// 5 minutes
const MONITOR_PERIOD = 600

const MAX_FREECOIN_ATTEMPTS = 3

const MIN_FREECOIN_SN = 26000
const MAX_FREECOIN_SN = 100000


var LocalRaidas = []string{
}

var Guardians = []string{
  "raida-guardian-tx.us",
  "g2.cloudcoin.asia",
  "guardian.ladyjade.cc",
  "watchdog.guardwatch.cc",
  "g5.raida-guardian.us",
  "goodguardian.xyz",
  "g7.ccraida.com",
  "raidaguardian.nz",
  "g9.guardian9.net",
  "g10.guardian25.com",
  "g11.raidacash.com",
  "g12.aeroflightcb300.com",
  "g13.stomarket.co",
  "guardian14.gsxcover.com",
  "guardian.keilagd.cc",
  "g16.guardianstl.us",
  "raida-guardian.net",
  "g18.raidaguardian.al",
  "g19.paolobui.com",
  "g20.cloudcoins.asia",
  "guardian21.guardian.al",
  "rg.encrypting.us",
  "g23.cuvar.net",
  "guardian24.rsxcover.com",
  "g25.mattyd.click",
  "g26.cloudcoinconsortium.art",
}


var ECHO_TIMEOUT_MULT = 100
var DEFAULT_TIMEOUT_MULT = 100

//const DNSSERVICE_URL = "https://ddns.cloudcoin.global/service/ddns"
const DNSSERVICE_URL = "http://209.205.66.11/service/ddns"

const MAX_ATTEMPTS_TO_PICK_RANDOM_RAIDA = 5
const CC_BINARY_HEADER_SIZE = 32
const CC_FILE_BINARY_EXTENSION = ".bin"
const CC_FILE_PNG_EXTENSION = ".png"

const (
  STATEMENT_TYPE_DEPOSIT = 0x0
  STATEMENT_TYPE_WITHDRAW = 0x1
  STATEMENT_TYPE_TRANSFER = 0x2
  STATEMENT_TYPE_SENT = 0x3
  STATEMENT_TYPE_BREAK = 0x4
  STATEMENT_TYPE_JOIN = 0x5
)


const (
  STATEMENT_RAID_TYPE_STRIPE = 0x0
)

const MAX_MEMO_LENGTH = 50
const MEMO_SEPARATOR = "*"



const INSTALL_DIR_ASSETS = "backassets"


const (
  HEALTH_CHECK_STATUS_PRESENT = 0x1
  HEALTH_CHECK_STATUS_NOT_PRESENT = 0x2
  HEALTH_CHECK_STATUS_NETWORK = 0x3
  HEALTH_CHECK_STATUS_COUNTERFEIT = 0x4
  HEALTH_CHECK_STATUS_UNKNOWN = 0x5
)

const DEFAULT_YEAR = 2000


const LEGACY_RAIDA_DOMAIN_NAME = "cloudcoin.global"


const (
  FILEPICKER_TYPE_FILE = "file"
  FILEPICKER_TYPE_FOLDER = "folder"
)

const MAX_RAIDA_DATAGRAM_SIZE = 1024

const SKYVAULT_TYPE_CARD = "card"
const SKYVAULT_TYPE_BIN = "bin"

const CARD_EXPIRATION_IN_YEARS = 5

const NEWSFEED_URL = "https://cloudcoinconsortium.com/newsfeed.js"

const ICON_FILENAME = "256x256.ico"
