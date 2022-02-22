package wallets

import (
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
)


type SkyWallet struct {
  Name string `json:"name"`
  Balance int `json:"balance"`
  Denominations map[int]int `json:"denominations"`
  Contents []uint32 `json:"contents"`
  CoinsByDenomination map[int][]*cloudcoin.CloudCoin `json:"-"`
  Statements []Statement `json:"statements"`
  IDCoin *cloudcoin.CloudCoin `json:"idcoin"`
  PNG string `json:"png,omitempty"`
}

type Statement struct {
  Guid string `json:"guid"`
  Type string `json:"type"`
  Amount int `json:"amount"`
  Balance int `json:"balance"`
  Time time.Time `json:"time"`
  Memo string `json:"memo"`
  Owner uint32 `json:"owner"`
}

type StatementDetail struct {
  Sn uint32 `json:"sn"`
  PownString string `json:"pownstring"`
  Result string `json:"result"`
}

type StatementTransaction struct {
  ID string `json:"id"`
  To string `json:"to"`
  From string `json:"from"`
  Type string `json:"type"`
  Details []StatementDetail `json:"details"`
}

func NewSkyWallet(name string) *SkyWallet {

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

  return &SkyWallet{
    Name: name,
    Balance: 0,
    Denominations: denominations,
    Contents: contents,
    CoinsByDenomination: coinsByDenomination,
  }
}

func (v *SkyWallet) SetIDCoin(cc *cloudcoin.CloudCoin) {
  v.IDCoin = cc
}

func (v *SkyWallet) GetCoinsByDenominations() map[int][]*cloudcoin.CloudCoin {
  return v.CoinsByDenomination
}
