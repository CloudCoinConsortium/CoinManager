package config

import (
	"io/ioutil"
	"os"
	"os/user"

	"github.com/BurntSushi/toml"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
)

type MainSection struct {
	HTTPTimeout             int    `toml:"http_timeout"`
	Domain              string `toml:"domain"`
	MaxNotesToSend int    `toml:"max_notes_to_send"`
  ExportBackground    string `toml:"export_background"`
	BrandColor					string `toml:"brand_color"`
  Guardians []string `toml:"guardians"`
  LocalRaidas []string `toml:"private_raidas"`
  

  DefaultTimeoutMult int `toml:"default_timeout_mult"`
  EchoTimeoutMult int `toml:"echo_timeout_mult"`
  ChangeServerSN int `toml:"change_server_sn"`
  EncryptionDisabled bool `toml:"encryption_disabled"`
  UseLocalRaidas bool `toml:"use_local_raidas"`
}



type RConfig struct {
	Title string      `toml:"title"`
	Help  string      `toml:"help"`
	Main  MainSection `toml:"main"`
}

func Ps() string {
	return string(os.PathSeparator)
}

func SetRootPath(home string) error {

  if home != "" {
    _, err := os.Stat(home)
    if err != nil {
      return err
    }

    ROOT_PATH = home  + Ps() + TOPDIR
    return nil
  }

	root, err := user.Current()
	if err != nil {
    return err
	}

	ROOT_PATH = root.HomeDir + Ps() + TOPDIR

  return nil
}

func GetConfigPath() string {
  return ROOT_PATH + Ps() + CONFIG_FILENAME
}


func ApplyString(data string) (*RConfig, error) {
	var conf RConfig
	if _, err := toml.Decode(data, &conf); err != nil {
		return nil, err
	}

  Apply(&conf)

  return &conf, nil
}

func Apply(conf *RConfig) {
  SetMainVar(&conf.Main.HTTPTimeout, &HTTP_TIMEOUT)
  SetMainVar(&conf.Main.Domain, &DEFAULT_DOMAIN)
  SetMainVar(&conf.Main.ExportBackground, &EXPORT_BACKGROUND)
  SetMainVar(&conf.Main.BrandColor, &BRAND_COLOR)
  SetMainVar(&conf.Main.MaxNotesToSend, &MAX_NOTES_TO_SEND)
  SetMainVar(&conf.Main.Guardians, &Guardians)
  SetMainVar(&conf.Main.LocalRaidas, &LocalRaidas)
  SetMainVar(&conf.Main.EchoTimeoutMult, &ECHO_TIMEOUT_MULT)
  SetMainVar(&conf.Main.DefaultTimeoutMult, &DEFAULT_TIMEOUT_MULT)
  SetMainVar(&conf.Main.ChangeServerSN, &PUBLIC_CHANGE_SN)
  SetMainVar(&conf.Main.EncryptionDisabled, &ENCRYPTION_DISABLED)
  SetMainVar(&conf.Main.UseLocalRaidas, &USE_LOCAL_RAIDAS)
}

func ApplySave(conf *RConfig) error {
	configFilePath := GetConfigPath()
  f, err := os.OpenFile(configFilePath, os.O_RDWR, 0644)
  if err != nil {
    return err
  }

  defer f.Close()

  Apply(conf)
  if err := toml.NewEncoder(f).Encode(&conf); err != nil {
    return err
  }

  return nil
}

func SetMainVar(l interface{}, r interface{}) {

  switch l.(type) {
  case *int:
    leftValue := l.(*int)
    rightValue := r.(*int)
    if *leftValue != 0 {
      *rightValue = *leftValue
    } else {
      *leftValue = *rightValue
    }
  case *string:
    leftValue := l.(*string)
    rightValue := r.(*string)
    if *leftValue != "" {
      *rightValue = *leftValue
    } else {
      *leftValue = *rightValue
    }
  case *[]string:
    leftValue := l.(*[]string)
    rightValue := r.(*[]string)
    if *leftValue != nil {
      *rightValue = *leftValue
    } else {
      *leftValue = *rightValue
    }
  case *bool:
    leftValue := l.(*bool)
    rightValue := r.(*bool)
    if leftValue != nil {
      *rightValue = *leftValue
    } else {
      *leftValue = *rightValue
    }



  default:
    return
  }
}

func ReadApplyConfig() (*RConfig, error) {
	configFilePath := GetConfigPath()
	var content []byte

	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
    if err := SaveDefaultConfig(); err != nil {
      return nil, perror.New(perror.ERROR_CONFIG_SAVE, "Failed to save default config: " + err.Error())
    }
	}

	content, err = ioutil.ReadFile(configFilePath)
	if err != nil {
    return nil, perror.New(perror.ERROR_CONFIG_READ, "Failed to read config file: " + err.Error())
	}

  conf, err := ApplyString(string(content))
	if err != nil {
		return nil, perror.New(perror.ERROR_CONFIG_PARSE, "Failed to parse config: " + err.Error())
	}

  return conf, nil
}

func SaveDefaultConfig() error {
	var conf RConfig
 
	configFilePath := GetConfigPath()
  f, err := os.Create(configFilePath)
  if err != nil {
    return err
  }

  defer f.Close()

  conf.Title = "CloudCoin Manager"
  conf.Help = "support@cloudcoin.global"
  conf.Main.HTTPTimeout = HTTP_TIMEOUT
  conf.Main.ExportBackground = EXPORT_BACKGROUND
  conf.Main.BrandColor = BRAND_COLOR
  conf.Main.MaxNotesToSend = MAX_NOTES_TO_SEND
  conf.Main.EchoTimeoutMult = ECHO_TIMEOUT_MULT
  conf.Main.DefaultTimeoutMult = DEFAULT_TIMEOUT_MULT
  conf.Main.ChangeServerSN = PUBLIC_CHANGE_SN
  conf.Main.EncryptionDisabled = ENCRYPTION_DISABLED
  conf.Main.UseLocalRaidas = USE_LOCAL_RAIDAS

  conf.Main.Guardians = Guardians
  conf.Main.LocalRaidas = LocalRaidas


  if err := toml.NewEncoder(f).Encode(&conf); err != nil {
    return err
  }

  return nil

}
