# Publishing NFTs So They Can Be Verified
NOTE: Regardless if you are tech savvy or not, our support is here to help you publish your NFTs. If you have any problems, your first step should be to call support. 

It is very easy to create NFTs using the CloudCoin Manager once you have everything set up.

Publishing NFTs means that you will post a copy of the image and NFT description in a place that only you have access to.
It is assumed that your domain name is owned and controlled by you. The only people who have access to your domains are the people you grant access to. 
Any art or descriptions that are found on your DNS or web servers could only be put there by you. Therefore, NFT collectors can know that their NFTs are
authentic by going to your domain and comparing what they own with what you published. This should make your NFTs more valuable.  

When you publish your NFTs, the CloudCoin Manager will contact your DNS server and command it to create a TXT record that cointains the location of the folder on your web page where the image and description of your NFTs are kept. Your NFTs will be given a FQDN (Fully Qaulified Domain Name) and this will become the name of your NFT. Suppose that you create an NFT and give it the title of "Jill-Having-Fun". Suppose the domain name you own is "happy-acres.com". Your NFT will become "Jill-Having-Fun.happy-acres.com". Your NFT becomes a PNG file with the name of "Jill-Having-Fun.happy-acres.com.ccnft.png". 

When a user wants to verify that you are the creator of your NFT, their CloudCoin Manager software will contact your DNS server and perform a TXT lookup for Jill-Having-Fun.happy-acres.com. Your DNS server will respond with something like "https://happy-acres.com/nft/Jill-Having-Fun". Then their CloudCoin Manager software will go to that URL and find a PNG and description file. The CloudCoin Manager will show your collector these files so they can compare them to what they have and verifty that the NFTs are authentic. 

To publish your NFTs, you will need: 
1. A Domain name and a DNS host that allows it to be controlled by an API.
2. A web server that can use PHP.

### DNS Host Setup
Coin Manager is set up to talk to CloudFlare, the most used DNS Hosting company. We are able to integrate many more DNS host companies, but will not do so unless asked. Call support to make that request. Otherwise, you may move your DNS over to CloudFlare for free. You will need to enter the API token for adding records to your DNS server. You can find this API by going to CLoudFlare.com.

1. Log in to CloudFlare. 
2. At the home screen, pick the Domain that you want to add NFTs to. 
3. At the Overview for that domain, scroll down and click on "Get your API token" on the lower right-hand side of the page.
![Overview](0.png?raw=true "CloudFlare Overview")
4. At the API Tokens page, click the "Create Token" button.
![Create API](1.png?raw=true "CloudFlare Create Token")
5. In the list of API Token templates, choose the top one (Edit Zone DNS) and click the "Use template" button. 
6. On the Create Token form, set the Permissions to: Zone, DNS, Edit.  Set Zone Resouces to: Include, Specific zone, yourdomain.com. Leave Client IP Address Filtering and TTL empty. Click on the "Contiue to summary" button. Then click "Create Token"
![API Token](1.png?raw=true "CloudFlare API Token Create Token")
7. The token is a long code with capital letters, lowercase letters, and numbers. Copy that and put it into the CloudCoin Manager: Tools > Settings > NFT Publishing > "CloudFlare DNS API Key".

### Web Server Upload Page 


You will need a webserver that supports PHP. You can use the PHP page below. You will need to change the password in the page, then put this password in the CloudCoin Manager: Tools > Settings > NFT Publishing > "CloudFlare DNS API Key".

Here is the PHP page you can put on your server: 
https://github.com/CloudCoinConsortium/CoinManager/blob/main/nft/upload.php


# NFT CLI


NFT CLI command is used to upload NFT files and their metadata to a remote server.

## Prerequisite

1. A remote server must run upload.php file.

2. At least one PHP variable (UPLOAD_SECRET_KEY) must be defined in the beginning of the PHP file.

The other variable (NFT_FOLDER) is optional.

```php
// Change The Secret Key to the one used in the Go Program.
define('UPLOAD_SECRET_KEY', '630d96ae6fe44400566a781dfcb7f2cc');

// By default, the files will be uploaded to the same folder where the script runs. Change it if necessary.
define('NFT_FOLDER', dirname(__FILE__));
```

## CLI command accepts one parameter. The parameter must point to the config file.

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

**upload_url** must point to the server's endpoint.

**upload_secret_key** must match the one in the upload.php file.

**cf_api_key** are the CloudFlare API credentials.

The other parameters are optional.





