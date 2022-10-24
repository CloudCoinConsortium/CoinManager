# NFT CLI

NFT CLI command is used to upload NFT files and their metadata to a remote server

## Prerequisite

1. A remote server must run upload.php file

2. At least one PHP variable (UPLOAD_SECRET_KEY) must be defined in the beginning of the PHP file.

The other variable (NFT_FOLDER) is optional

```php
// Change The Secret Key to the one used in the Go Program
define('UPLOAD_SECRET_KEY', '630d96ae6fe44400566a781dfcb7f2cc');

// By default the files will be uploaded to the same folder where the script runs. Change it if necessary
define('NFT_FOLDER', dirname(__FILE__));
```

## CLI command accepts one parameter. The parameter must point to the config file

```console
nft_cli /home/user/config.txt
```

## Possible flags

[-version](README.md#-version)

[-help](README.md#-help)

## -version
example usage:
```console
nft_cli -version
```

Sample response:
```console
0.0.1
```

## -help
example usage:
```
nft_cli -help
```

Sample response:
```console
Usage of ./nft_cli:
./nft_cli <config-path>
./nft_cli [-help]
./nft_cli [-version]
<config-path> path to the TOML config file. Required
```

## Config
You can configure the behaviour of the NFT CLI by using a configuration file. The config file format must follow TOML syntax.


```toml
upload_url = "https://cloudcoin.digital/nft/upload.php"
upload_secret_key = "630d96ae6fe44400566a781dfcb7f2cc"
cf_api_key = "222233445566b119df51c67862c11a64f1111"
```

Mandatory parameters:

**upload_url** must point to the server's endpoint

**upload_secret_key** must match the one in the upload.php file

**cf_api_key** are the CloudFlare API credentials

The other parameters are optional





