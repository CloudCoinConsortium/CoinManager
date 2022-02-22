package raida

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type ShowPayment struct {
	Servant
}

type ShowPaymentOutput struct {
  Item ShowStatementItem
}
/*
type ShowPaymentItem struct {
  Guid string
  TransactionType int
  Amount int
  Balance int
  TimeStamp time.Time
  MemoRAIDType int
  Memo string
}
*/
func NewShowPayment(progressChannel chan interface{}) *ShowPayment {
	return &ShowPayment{
		*NewServant(progressChannel),
	}
}

func (v *ShowPayment) ShowPayment(ctx context.Context, guid string) (*ShowPaymentOutput, error) {
	logger.L(ctx).Debugf("ShowPayments for GUID %s ", guid)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_SHOW_PAYMENT, idx, cce)
    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    rbs, _ := utils.GenerateHex(8)
    rbsb, _ := hex.DecodeString(rbs)
    encb = append(encb, rbsb...)

    bguid, _ := hex.DecodeString(guid)
    encb = append(encb, bguid...)
    encb = append(encb, rbsb...)
    encb = append(encb, v.GetSignature()...)
    encb, err = v.EncryptIfRequired(ctx, cce, idx, encb)
    if err != nil {
      logger.L(ctx).Debugf("Failed to encrypt body for R%d: %s", idx, err.Error())
      continue
    }
    params[idx] = append(params[idx], encb...)
	}

  v.UpdateHeaderUdpPackets(params)
	results := v.Raida.SendRequest(ctx, params, 0)
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.CommonSuccessFunction, cce)

  logger.L(ctx).Debugf("rd %v", rdata)

  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  counters := make(map[string]int, 0)
  hvvals := make(map[string]ShowStatementItem, 0)
  memoRaw := make([]string, v.Raida.TotalServers())

  so := &ShowPaymentOutput{}
  for idx, item := range(rdata) {
    if item == nil {
      continue
    }

    bitem := item.([]byte)

    transactionType := int(bitem[0])
    amount := binary.BigEndian.Uint32(bitem[1:5])


/*
    ts := make([]byte, 0)
    ts = append(ts, byte(0x0), byte(0x0))
    ts = append(ts, bitem[5:11]...)

    tInt := binary.BigEndian.Uint64(ts)
    tm := time.Unix(int64(tInt), 0)
*/
    tm := utils.GetTimeFromBytes(bitem[5:11])




    account := utils.GetUint24(bitem[11:14])
    sender := utils.GetUint24(bitem[14:17])
   // account := 10
   // sender := 11

    logger.L(ctx).Debugf("acc %d sender %d", account, sender)

    memoRaw[idx] = string(bitem[17:])

    sts := tm.String()
    hashStr := sts + guid + strconv.Itoa(int(account)) + strconv.Itoa(int(transactionType))

    hash := utils.GetHash(hashStr)
    counters[hash]++

    rItem := ShowStatementItem{
      Guid: guid,
      TransactionType: transactionType,
      Amount: int(amount),
      Owner: account,
      TimeStamp: tm,
    }

    hvvals[hash] = rItem

    logger.L(ctx).Debugf("rr%d ttype %d, ts %s(%v), ow %d, sender %d, hs %s, hash %s, amount %d", idx, transactionType, tm.String(), tm, account, sender, hashStr, hash, amount)
  }

  topKey, err := PickTopKey(ctx, counters)
  if err != nil {
    return nil, perror.New(perror.ERROR_RAIDA_QUORUM, "Failed to pick top payment")
  }


  rItem := hvvals[topKey]
  memo, err := v.GetMessageFromStripesAndMirrors(ctx, memoRaw)
  if err != nil {
    logger.L(ctx).Warnf("Failed to assemble memo for statement %s: %s", guid, err.Error())
    return nil, perror.New(perror.ERROR_ASSEMBLE_STRIPE_MIRRORS, "Failed to assemble memo")
  }

  logger.L(ctx).Debugf("memo assembbb %s",memo)
  rItem.Memo = memo

  so.Item = rItem

  return so, nil
}

