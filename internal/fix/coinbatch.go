package fix

import "github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"

type FixBatch struct {
  CoinsPerRaida map[int][]cloudcoin.CloudCoin
  Tickets []string 
}



