package storage

import (
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/storage/filesystem"
)


var driver StorageInterface 

func Init() error {
  logger.L(nil).Debugf("Initializing storage driver \"%s\"", config.STORAGE_DRIVER)

  switch (config.STORAGE_DRIVER) {
  case "filesystem":
    driver = filesystem.New()
  default:
    return perror.New(perror.ERROR_STORAGE_DRIVER_NOT_FOUND, "Storage Driver is Not Found")
  }


  err := driver.InitWallets(nil)
  if err != nil {
    return perror.New(perror.ERROR_STORAGE_DRIVER_INIT, "Storage Driver is Not Found")
  }

  logger.L(nil).Debugf("Storage driver \"%s\" initialized", config.STORAGE_DRIVER)

  return nil
}


func GetDriver() StorageInterface {
  return driver
}
