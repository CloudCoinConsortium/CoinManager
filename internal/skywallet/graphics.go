package skywallet

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/cloudcoin"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
)



func (v *SkyWalleter) DrawIDCard(ctx context.Context, cc *cloudcoin.CloudCoin, cardNumber, cvv, expDate, ipAddress string) ([]byte, error) {
  mainTemplate := config.ROOT_PATH + config.Ps() + config.DIR_TEMPLATES + config.Ps() + "card.png"

  rFontPath := config.ROOT_PATH + config.Ps() + config.DIR_TEMPLATES + config.Ps() + "Barlow-Regular.ttf"
  //fontPath := core.GetTemplateDirPath() + config.Ps() + "Barlow-Bold.ttf"

  fontBytes, err := ioutil.ReadFile(rFontPath)
  if err != nil {
    return nil, err
  }

  v.rfont, err = freetype.ParseFont(fontBytes)
  if err != nil {
    return nil, err
  }

  logger.L(ctx).Debugf("reading %s", mainTemplate)
  pngBytes, err := ioutil.ReadFile(mainTemplate)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read %s: %s", mainTemplate, err.Error())
    return nil, err
  }
  
  img, fileType, err := image.Decode(bytes.NewBuffer(pngBytes))
  if err != nil {
    logger.L(ctx).Errorf("Failed to decode %s:", mainTemplate, err.Error())
    return nil, err
  }

  bounds := img.Bounds()
  logger.L(ctx).Debugf("PNG %v", &png.Encoder{})
  logger.L(ctx).Debugf("Decoded PNG %s: %dx%d", fileType, bounds.Max.X, bounds.Max.Y)

  canvas := image.NewNRGBA(bounds)
  draw.Draw(canvas, bounds, img, image.Point{0, 0}, draw.Src)


  wcolor := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} 
  gcolor := color.RGBA{0xDD, 0xDD, 0xDD, 0xFF} 
  bcolor := color.RGBA{0x0, 0x0, 0x0, 0xFF} 
  v.DrawText(ctx, canvas, 64, 245, wcolor, 48, cardNumber)
  v.DrawText(ctx, canvas, 450, 322, wcolor, 35, expDate)
  v.DrawText(ctx, canvas, 64, 375, wcolor, 50, cc.GetSkyName())
  v.DrawText(ctx, canvas, 64, 300, gcolor, 16, "Keep these numbers secret. Do not give to merchants")
  v.DrawText(ctx, canvas, 64, 638, bcolor, 35, "CVV (Keep Secret): " + cvv)
  v.DrawText(ctx, canvas, 174, 716, wcolor, 18, "IP " + ipAddress)

  buff := bytes.NewBuffer([]byte{})
  switch fileType {
  case "png":
    png.Encode(buff, canvas)
  default:
    return nil, perror.New(perror.ERROR_UNSUPPORTED_FILE_TYPE, "Template file type " + fileType + " is not supported")
  }

  // Embeds data
  data, _ := cc.GetData()
  bytes, err := v.InsertPNGData(ctx, buff.Bytes(), data)
  if err != nil {
    return nil, err
  }

  return bytes, nil
}

func (v *SkyWalleter) DrawText(ctx context.Context, img *image.NRGBA, x, y int, color color.RGBA, fsize int, text string) {

  var font *truetype.Font
//  if (isBold) {
    font = v.rfont
//  } else {
 //   font = v.rfont
//  }

  c := freetype.NewContext()
  size := float64(fsize)

  ptf := c.PointToFixed(size) >> 6

  fbounds := font.Bounds(ptf)

  gw := float32(fbounds.Max.X - fbounds.Min.X)
  gh := float32(fbounds.Max.Y - fbounds.Min.Y)

  glyphCount := len(text)

  imageWidth := uint32(gw * float32(glyphCount))
  imageHeight := uint32(gh)

  logger.L(ctx).Debugf("p %d %d", imageWidth, imageHeight)

  imageBounds := image.Rect(0, 0, int(imageWidth), int(imageHeight))

  rgba := image.NewNRGBA(imageBounds)

  c.SetFont(font)
  c.SetClip(rgba.Bounds())
  c.SetFontSize(size)
  c.SetDst(rgba)
  c.SetSrc(image.NewUniform(color))

  pt := freetype.Pt(0, 0 + int(ptf))
  for _, s := range(text) {
    _, err := c.DrawString(string(s), pt)
    if err != nil {
      logger.L(ctx).Errorf("Failed to write a symbol: %s", err.Error())
      return
    }

    idx := font.Index(rune(s))
    spacing := float64(font.HMetric(ptf, idx).AdvanceWidth) 
  	pt.X += c.PointToFixed(spacing + 0.05)
  }

  b := img.Bounds()
  draw.Draw(img, b.Add(image.Pt(x, y)), rgba, image.ZP, draw.Over)
}
