package parent

import "github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"

type Parent struct {
  StartSN int
  MaxCoins int
}


func New() *Parent {
  return &Parent{
    500000,
    20,
  }
}


func (v *Parent) GetCoin(sn uint32) *cloudcoin.CloudCoin {

  cc := cloudcoin.NewFromData(sn)

  return cc
}

