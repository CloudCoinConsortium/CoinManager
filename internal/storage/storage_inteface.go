package storage

import (
	"context"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/transactions"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/wallets"
)

type StorageInterface interface {
  InitWallets(context.Context) error
  GetWallets(context.Context) ([]wallets.Wallet, error)
  GetWallet(context.Context, string) (*wallets.Wallet, error)
  GetWalletWithBalance(context.Context, string) (*wallets.Wallet, error)
  GetWalletWithContents(context.Context, string) (*wallets.Wallet, error)
  DeleteWallet(context.Context, string) error
  RenameWallet(context.Context, *wallets.Wallet, string) error
  CreateWallet(context.Context, string, string, string) error

  GetSkyWallets(context.Context) ([]wallets.SkyWallet, error)
  GetSkyWallet(context.Context, string) (*wallets.SkyWallet, error)
  GetFirstSkyWallet(context.Context) (*wallets.SkyWallet, error)
  DeleteSkyWallet(context.Context, string) error
  CreateSkyWallet(context.Context, *cloudcoin.CloudCoin) error
  CreateSkyWalletWithData(context.Context, *cloudcoin.CloudCoin, []byte)  error

  EmptyLocation(context.Context, *wallets.Wallet, int)  error

 // PutInImport(*wallets.Wallet, string) error
  PutInImported(context.Context, *wallets.Wallet, string) error
  PutInSuspect(context.Context, *wallets.Wallet, []cloudcoin.CloudCoin) error
  CoinExistsInTheWallet(context.Context, *wallets.Wallet, *cloudcoin.CloudCoin) bool
  CoinExistsInTheWalletAndAuthentic(context.Context, *wallets.Wallet, *cloudcoin.CloudCoin) bool

  // Updates coin in the same directory (for fracked fixer for instance)
  UpdateCoins(context.Context, *wallets.Wallet, []cloudcoin.CloudCoin) error

  // Updates status by coin's Grade
  UpdateStatus(context.Context, *wallets.Wallet, []cloudcoin.CloudCoin) error
  UpdateStatusForNewCoin(context.Context, *wallets.Wallet, []cloudcoin.CloudCoin) error

  // Forcibly sets status
  SetLocation(context.Context, *wallets.Wallet, []cloudcoin.CloudCoin, int) error


  GetCoins(context.Context, *wallets.Wallet, int) ([]cloudcoin.CloudCoin, error)
  ReadCoin(context.Context, *wallets.Wallet, *cloudcoin.CloudCoin) error

  UpdateWalletBalance(context.Context, *wallets.Wallet) error

  AppendTransaction(context.Context, *wallets.Wallet, *transactions.Transaction) error
  GetTransactions(context.Context, *wallets.Wallet) ([]transactions.Transaction, error)

  AppendSkyTransactionDetails(context.Context, *wallets.SkyWallet, *wallets.StatementTransaction) error
  GetSkyTransactionDetails(context.Context, *wallets.SkyWallet, string) (*wallets.StatementTransaction, error)

  MoveCoins(context.Context, *wallets.Wallet, *wallets.Wallet, []cloudcoin.CloudCoin) (int, error)

  SaveCoin(context.Context, *cloudcoin.CloudCoin, string, string) error

  GetReceipt(context.Context, *wallets.Wallet, string) ([]transactions.TransactionDetail, error)
  DeleteTransactionsAndReceipts(context.Context, *wallets.Wallet) error


  AppendSenderHistory(context.Context, string) error
  GetSenderHistory(context.Context, string) ([]string, error)
  UpdateWalletContentsForLocation(context.Context, *wallets.Wallet, int) error 

  ReadCoinsInLocation(context.Context, *wallets.Wallet, int) ([]cloudcoin.CloudCoin, int, error)
}

