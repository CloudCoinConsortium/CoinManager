package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *Ilogger

type Ilogger struct {
  L *zap.SugaredLogger
  lastCaller string
  lineNo int
  lastTs int64
  Ls string
  nlogger *zap.SugaredLogger
}


func SetContext(ctx context.Context, prefix string) context.Context {
  if globalLogger == nil {
    return ctx
  }

  rqId := GenerateRequestID(prefix)

  zlogger := globalLogger.With("rqid", rqId)
	ctx = context.WithValue(ctx, "logger", zlogger)

  return ctx
}

func Init(rootPath string, needConsole bool) error {
  logDir := rootPath
  _, err := os.Stat(logDir)
  if err != nil {
    if os.IsNotExist(err) {
      return perror.New(perror.ERROR_DIRECTORY_NOT_EXIST, "Log directory doesn't exist")
    }

    return perror.New(perror.ERROR_FILESYSTEM, "Failed to Stat() logdir: " + err.Error())
  }

  logFilePath := logDir + string(os.PathSeparator) + config.LOG_FILENAME
	_, err = os.Stat(logFilePath)
  if err != nil {
    if os.IsNotExist(err) {
      _, err = os.Create(logFilePath)
      if err != nil {
        return perror.New(perror.ERROR_FILESYSTEM, "Failed to create empty logfile: " + err.Error())
      }
    	_, err = os.Stat(logFilePath)
      if err != nil {
        return perror.New(perror.ERROR_FILESYSTEM, "Failed to Stat() and Create() logfile: " + err.Error())
      }
    } else {
      return perror.New(perror.ERROR_FILESYSTEM, "Failed to Stat() logfile: " + err.Error())
    }
  }

  globalLogger = &Ilogger{}
  if runtime.GOOS == "windows" {
    globalLogger.Ls = "\r\n"
  } else {
    globalLogger.Ls = "\n"
  }

  lj := &lumberjack.Logger{
    Filename: logFilePath,
    MaxSize: config.MAX_LOG_SIZE,
    MaxBackups: 5,
    MaxAge: 30,
    Compress: false,
  }

  w := zapcore.AddSync(lj)

  cfgEncoder := zap.NewProductionEncoderConfig()
  cfgEncoder.EncodeTime = TimeEncoder
  cfgEncoder.EncodeLevel = zapcore.CapitalLevelEncoder
	cfgEncoder.EncodeCaller = zapcore.ShortCallerEncoder

//  e := zapcore.NewConsoleEncoder(cfgEncoder)
  e := NewSRConsoleEncoder(cfgEncoder)
  core := zapcore.NewCore(e, w, zapcore.DebugLevel)

  var ncore zapcore.Core
  if needConsole {
    consoleCore := zapcore.NewCore(e, os.Stdout, zapcore.DebugLevel)
    ncore = zapcore.NewTee(core, consoleCore)
  } else {
    ncore = core
  }

  
  globalLogger.L = zap.New(ncore, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()

  return nil
}

func LevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
  /*
  if ilogger.lineNo == 0 || ilogger.lastCaller == "" {
    return
  }

  enc.AppendString("\t")
  */
  if (runtime.GOOS != "windows") {
    zapcore.CapitalColorLevelEncoder(level, enc)
  } else {
    zapcore.CapitalLevelEncoder(level, enc)
  }

}

func CallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
  enc.AppendString(caller.TrimmedPath())
}

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
  /*
  if ilogger.lineNo == 0 || ilogger.lastCaller == "" {
    return
  }

  if t.Unix() == ilogger.lastTs {
    return
  }
  */

  enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func newWinFileSink(u *url.URL) (zap.Sink, error) {
  return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func L(ctx context.Context) *Ilogger {
  if ctx == nil {
    return globalLogger
  }

  logger, ok := ctx.Value("logger").(*Ilogger)
  if !ok {
    return globalLogger
  }

  return logger
}

func (v *Ilogger) With(k, val string) *Ilogger {
  newLogger := &Ilogger{}
  newLogger.L = v.L.With(zap.String(k, val))

  return newLogger
}

func (v *Ilogger) Debug(fmt string) {
  v.DoWrapBefore()
  v.L.Debug(fmt)
  v.DoWrapAfter()
}

func (v *Ilogger) Info(fmt string) {
  v.DoWrapBefore()
  v.L.Info(fmt)
  v.DoWrapAfter()
}

func (v *Ilogger) Debugf(fmt string, a ...interface{}) {
  v.DoWrapBefore()
  v.L.Debugf(fmt, a...)
  v.DoWrapAfter()
}

func (v *Ilogger) Infof(fmt string, a ...interface{}) {
  v.DoWrapBefore()
  v.L.Infof(fmt, a...)
  v.DoWrapAfter()
}

func (v *Ilogger) Warnf(fmt string, a ...interface{}) {
  v.DoWrapBefore()
  v.L.Warnf(fmt, a...)
  v.DoWrapAfter()
}

func (v *Ilogger) Errorf(fmt string, a ...interface{}) {
  v.DoWrapBefore()
  v.L.Errorf(fmt, a...)
  v.DoWrapAfter()
}

func (v *Ilogger) DoWrapBefore() {
  /*
  _, file, no, ok := runtime.Caller(2)
  if ok {
    path := filepath.Base(file)
    dir := filepath.Base(filepath.Dir(file))

    key := dir + config.Ps() + path

    if v.lastCaller != "" && v.lastCaller != key {
      v.lastCaller = ""
      v.lineNo = 0
      v.slogger.Info("")
    }
    v.lastCaller = key
    v.lineNo = no
  }

  nowTime := time.Now()
  nowTs := nowTime.Unix()

  savedLC := v.lastCaller
  savedLN := v.lineNo
  if nowTs != v.lastTs {
      v.lastCaller = ""
      v.lineNo = 0
      v.slogger.Info(nowTime.Format("2006-01-02 15:04:05"))
  }

  v.lastCaller = savedLC
  v.lineNo = savedLN
*/



}

func (v *Ilogger) DoWrapAfter() {
}

// can't import from utils
func GenerateRequestID(prefix string) (string) {
	bytes := make([]byte, 6)

	if _, err := rand.Read(bytes); err != nil {
    return "???"
	}

	return prefix + "-" + hex.EncodeToString(bytes)
}
