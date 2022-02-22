package wallets

import (
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
)


type Wallet struct {
  Name string `json:"name"`
  Email string `json:"email"`
  PasswordHash string `json:"password_hash"`
  Balance int `json:"balance"`
  Denominations map[int]int `json:"denominations"`
  Contents []uint32 `json:"contents"`
  CoinsByDenomination map[int][]*cloudcoin.CloudCoin `json:"-"`
  Transactions []transactions.Transaction `json:"transactions"`
}


func New(name, email, passwordHash string) *Wallet {

  denominations := make(map[int]int, 0)
  contents := make([]uint32, 0)
  coinsByDenomination := make(map[int][]*cloudcoin.CloudCoin, 0)

  denominations[1] = 0
  denominations[5] = 0
  denominations[25] = 0
  denominations[100] = 0
  denominations[250] = 0

  coinsByDenomination[1] = []*cloudcoin.CloudCoin{}
  coinsByDenomination[5] = []*cloudcoin.CloudCoin{}
  coinsByDenomination[25] = []*cloudcoin.CloudCoin{}
  coinsByDenomination[100] = []*cloudcoin.CloudCoin{}
  coinsByDenomination[250] = []*cloudcoin.CloudCoin{}

  return &Wallet{
    Name: name,
    Email: email,
    PasswordHash: passwordHash,
    Balance: 0,
    Denominations: denominations,
    Contents: contents,
    CoinsByDenomination: coinsByDenomination,
  }
}

func (v *Wallet) GetCoinsByDenominations() map[int][]*cloudcoin.CloudCoin {
  return v.CoinsByDenomination
}
