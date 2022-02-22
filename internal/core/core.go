package core

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	//"go.uber.org/zap/zapcore"
)

var wh unsafe.Pointer

func CreateRootFolder() error {
	rootDir := config.ROOT_PATH

  err := MkDir(nil, rootDir)
  if err != nil {
    return err
  }

  err = MkDir(nil, rootDir + config.Ps() + config.DIR_TEMPLATES)
  if err != nil {
    return err
  }

  return nil
}

func ReadIcon() ([]byte, error) {
  exPath, err := os.Executable()
  if err != nil {
    logger.L(nil).Debugf("Failed to get exec path: %s", err.Error())
    return nil, perror.New(perror.ERROR_FILESYSTEM, "Failed to get exec path")
  }

  exDir := filepath.Dir(exPath)
  file := exDir + config.Ps() + config.ICON_FILENAME

  bytes, err := ioutil.ReadFile(file)
  if err != nil {
    logger.L(nil).Debugf("Failed to read icon file %s", file)
    return nil, err
  }

  return bytes, nil
}


func GetUIPath(path string) (string, error) {
  var file string

  if path == "" {
    exPath, err := os.Executable()
    if err != nil {
      logger.L(nil).Debugf("Failed to get exec path: %s", err.Error())
      return "", perror.New(perror.ERROR_FILESYSTEM, "Failed to get exec path")
    }

    exDir := filepath.Dir(exPath)
    file = exDir + config.Ps() + "index.html"
  } else {
    file = path
  }

  _, err := os.Stat(file)
  if err != nil {
    return "", perror.New(perror.ERROR_FILESYSTEM, "Failed to access UI file " + file + ": " + err.Error())
  }

  return file, nil

}

func UpdateAssets() error {
  tDir := GetTemplateDirPath()

  exPath, err := os.Executable()
  if err != nil {
    logger.L(nil).Debugf("Failed to get exec path: %s", err.Error())
    return perror.New(perror.ERROR_FILESYSTEM, "Failed to get exec path")
  }

  exDir := filepath.Dir(exPath)
  assetsDir := exDir + config.Ps() + config.INSTALL_DIR_ASSETS

  logger.L(nil).Debugf("Exec path %s: %s", exDir, tDir)

  // No errors
  files, err := ioutil.ReadDir(assetsDir)
  if err != nil {
    logger.L(nil).Warnf("Can't read original assets dir %s: %s", assetsDir, err.Error())
    return nil
  }


  for _, f := range(files) {
    fname := assetsDir + config.Ps() + f.Name()
    iname := tDir + config.Ps() + f.Name()
    logger.L(nil).Debugf("Found %s. Dst %s", fname, iname)



    file, err := os.Open(fname)
    if err != nil {
      logger.L(nil).Warnf("Unable to read %s, skipping it: %s", fname, err.Error())
      continue
    }

    defer file.Close()

  	_, err = os.Stat(iname)
	  if os.IsNotExist(err) {
      logger.L(nil).Debugf("Copying file %s to Templates", fname)


      newdstFile, err := os.Create(iname)
      if err != nil {
        logger.L(nil).Warnf("Failed to create file %s: %s", iname, err.Error())
        continue
      }

      _, err = io.Copy(newdstFile, file)
      if err != nil {
        logger.L(nil).Warnf("Failed to copy file %s: %s", fname, err.Error())
        continue
      }

      logger.L(nil).Debugf("Copied %s", fname)
      continue
  	}

    ofile, err := os.Open(iname)
    if err != nil {
      logger.L(nil).Warnf("Unable to read %s, skipping it: %s", iname, err.Error())
      continue
    }

    defer ofile.Close()


    newFileHash, err := utils.GetMD5File(file)
    if err != nil {
      logger.L(nil).Warnf("Cant get md5 for %s", fname)
      continue
    }

    myFileHash, err := utils.GetMD5File(ofile)
    if err != nil {
      logger.L(nil).Warnf("Cant get md5 for %s", iname)
      continue
    }


    if myFileHash != newFileHash {
      logger.L(nil).Debugf("Updating %s", fname)



      os.Remove(iname)

      
      newdstFile, err := os.Create(iname)
      if err != nil {
        logger.L(nil).Warnf("Failed to create file %s: %s", iname, err.Error())
        continue
      }

      defer newdstFile.Close()

      _, err = file.Seek(0, io.SeekStart)
      if err != nil {
        logger.L(nil).Warnf("Failed to seek file %s", iname, err.Error())
        continue
      }

      _, err = io.Copy(newdstFile, file)
      if err != nil {
        logger.L(nil).Warnf("Failed to copy file %s: %s", fname, err.Error())
        continue
      }
      logger.L(nil).Debugf("Replaced %s", fname)
    }



    logger.L(nil).Debugf("File hash %s:%s", newFileHash, myFileHash)

  }



  return nil
}


func GetTemplateDirPath() string {
  return config.ROOT_PATH + config.Ps() + config.DIR_TEMPLATES
}


func MkDir(ctx context.Context, path string) error {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) {
		return nil
	}

	err = os.Mkdir(path, 0700)
	if err != nil {
    logger.L(ctx).Debugf("Created %s", path)
    return perror.New(perror.ERROR_FAILED_TO_CREATE_DIRECTORY, "Failed to create folder: " + err.Error())
	}

  return nil
}

func ExitError(err error) {
  switch err.(type) {
    case *perror.ProgramError:
      perr := err.(*perror.ProgramError)
      json := perr.ToJson()
      fmt.Printf("%s\n" , json)
    default:
      perr := perror.New(perror.ERROR_INTERNAL, "Internal Error: " + err.Error())
      json := perr.ToJson()
      fmt.Printf("%s\n" , json)
  }

  os.Exit(1)
}

func GetUIHandle() unsafe.Pointer {
  return wh
}

func SetUIHandle(swh unsafe.Pointer) {
  wh = swh
  logger.L(nil).Debugf("Set UI pointer %v", wh)
}
