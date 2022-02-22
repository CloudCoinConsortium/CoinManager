package wallets

import "github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"



type WalletInterface interface {
  GetCoinsByDenominations() map[int][]*cloudcoin.CloudCoin 
}
