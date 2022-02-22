package raida

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/utils"
)


type ShowStatement struct {
	Servant
}

type ShowStatementOutput struct {
  Items []ShowStatementItem
}

type ShowStatementItem struct {
  Guid string
  TransactionType int
  Amount int
  Balance int
  TimeStamp time.Time
  MemoRAIDType int
  Memo string
  Owner uint32
}

func NewShowStatement(progressChannel chan interface{}) *ShowStatement {
	return &ShowStatement{
		*NewServant(progressChannel),
	}
}

func (v *ShowStatement) ShowStatement(ctx context.Context, cc *cloudcoin.CloudCoin) (*ShowStatementOutput, error) {
	logger.L(ctx).Debugf("ShowStatements for coin %d ", cc.Sn)

	params := make([][]byte, v.Raida.TotalServers())
  cce, err := v.GetEncryptionCoin(ctx)
  if err != nil {
    logger.L(ctx).Warnf("Failed to get ID coin to encrypt body. Do you have at least one ID coin?")
  }

  // Iterate over raida servers
	for idx, _ := range params {
  	params[idx] = v.GetHeaderSky(COMMAND_SHOW_STATEMENT, idx, cce)

    encb := make([]byte, 0)
    encb = append(encb, v.GetChallenge()...)

    cbytes := utils.ExplodeSn(cc.Sn)
    an, _ := hex.DecodeString(cc.Ans[idx])

    encb = append(encb, cbytes...)
    encb = append(encb, an...)

    // Rows * 100 0x1 - means 100 rows. 0 - max rows
    encb = append(encb, byte(0x0))

    // Year 00 = 2000
    encb = append(encb, byte(0x0))

    // Month
    encb = append(encb, byte(0x0))

    // Day
    encb = append(encb, byte(0x0))

    // 00 - stripe 01 - mirror, 11 - 2nd mirror, FF stripe, mirror and 2nd mirror
    encb = append(encb, byte(0x0))

    // Encryption
    encb = append(encb, v.GetChallenge()...)

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
  pownArray, rdata := v.ProcessGenericResponsesCommon(ctx, nil, results, v.StatementSuccessFunction, cce)

  if !v.IsQuorumCollected(pownArray) {
    details := v.GetErrorsFromResults(results)
    logger.L(ctx).Errorf("Not enough valid responses. Quorum is not reached")
    return nil, perror.NewWithDetails(perror.ERROR_RAIDA_QUORUM, "Not enough valid responses from the RAIDA servers", details)
  }

  itemSize := 31 + config.MAX_MEMO_LENGTH
  so := &ShowStatementOutput{}
  so.Items = make([]ShowStatementItem, 0)

  itemsCnt := make(map[string]int)
  itemsVals := make(map[string]ShowStatementItem)

  memoRaw := make(map[string][]string, 0)

  for idx, item := range(rdata) {
    if item == nil {
      continue
    }

    bitem := item.([]byte)
    balance := binary.BigEndian.Uint32(bitem[0:4])

    btrs := bitem[4:]
    logger.L(ctx).Debugf("raida %d, balance %d, datalen %d", idx, balance, len(btrs))

    ti := 0
    for i := 0; i < len(btrs);  {
      logger.L(ctx).Debugf("DOing idx %d i=%d len=%d", idx, i, len(btrs))
      slice := btrs[i:i + itemSize]
      if (len(slice) == 0) {
        logger.L(ctx).Debugf("Done raida %d", idx)
        break
      }

      if len(slice) != itemSize {
        logger.L(ctx).Warnf("Raida %d, Invalid transaction item#%d size %d. It must be at least %d bytes", idx, ti, len(slice), itemSize)
        break
      }

      guid := hex.EncodeToString(slice[0:16])
      transactionType := int(slice[16])
      amount := binary.BigEndian.Uint32(slice[17:21])
      balance := binary.BigEndian.Uint32(slice[21:25])





  /*    
      ts := make([]byte, 0)
      ts = append(ts, byte(0x0), byte(0x0))
      ts = append(ts, slice[25:31]...)

      tInt := binary.BigEndian.Uint64(ts)
      tm := time.Unix(int64(tInt), 0)
*/
      ts := utils.GetTimeFromBytes(slice[25:31])









      rType := int(slice[31])
      // Looking for two zero bytes
      // We assume there are 50 bytes
      //memoSlice := slice[32:]

      bmemo := bytes.Trim(slice[32:], "\x00")
      memo := string(bmemo)
      i += itemSize

      logger.L(ctx).Debugf("Ti %s %d type=%d amount=%d balance=%d ts=%s rtype=%d memo %s", guid, ti, transactionType, amount, balance, ts.String(), rType, memo)

      itemsCnt[guid]++
      itemsVals[guid] = ShowStatementItem{
        Guid: guid,
        TransactionType: transactionType,
        Amount: int(amount),
        Balance: int(balance),
        TimeStamp: ts,
        MemoRAIDType: rType,
      }

      _, ok := memoRaw[guid]
      if !ok {
        memoRaw[guid] = make([]string, v.Raida.TotalServers())
      }

      memoRaw[guid][idx] = memo

      ti++
    }
  }

  for guid, cnt := range(itemsCnt) {
    if cnt < config.MIN_QUORUM_COUNT {
      logger.L(ctx).Warnf("Statement %s doesn't have enough votes. Skipping it", guid)
      continue
    }

    logger.L(ctx).Debugf("Have in mind guid %s memo %v", guid, memoRaw[guid])

    memo, err := v.GetMessageFromStripesAndMirrors(ctx, memoRaw[guid])
    if err != nil {
      logger.L(ctx).Warnf("Failed to assemble statement %s: %s", guid, err.Error())
      continue
    }

    si := itemsVals[guid]
    si.Memo = memo

    so.Items = append(so.Items, si)
  }

  return so, nil
}

func (v *Servant) StatementSuccessFunction(ctx context.Context, idx int, status int, coins []cloudcoin.CloudCoin, rdata []byte) (int, interface{}) {
  cmdName := v.Raida.DetectionAgents[idx].CurrentContext
  if (status == RESPONSE_STATUS_SUCCESS) {
    logger.L(ctx).Debugf("RAIDA%d (command %s) Success Status %d, rest data len %d", idx, cmdName, status, len(rdata))
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  if (status == RESPONSE_STATUS_NO_STATEMENTS) {
    logger.L(ctx).Debugf("RAIDA%d (command %s) Success Status %d, (No Statements Yet)", idx, cmdName, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_PASS)
    return config.RAIDA_STATUS_PASS, rdata[config.RAIDA_RESPONSE_HEADER_SIZE:]
  }

  if (status == RESPONSE_STATUS_FAIL || status == RESPONSE_STATUS_FAILED_AUTH) {
    logger.L(ctx).Debugf("Raida%d Fail Status %d", idx, status)
    v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_FAIL)
    return config.RAIDA_STATUS_FAIL, nil
  } 
  
  logger.L(ctx).Debugf("Raida%d Unknown Status %d", idx, status)
  v.SetCoinsStatus(coins, idx, config.RAIDA_STATUS_ERROR)

  return config.RAIDA_STATUS_ERROR, nil
}


