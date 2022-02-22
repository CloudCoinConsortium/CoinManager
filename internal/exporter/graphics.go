package exporter

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/core"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/perror"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Draws a PNG file and puts the Stack in there
func (v *Exporter) DrawData(ctx context.Context, amountStr, textAmount string, data []byte) ([]byte, error) {
  mainTemplate := config.ROOT_PATH + config.Ps() + config.DIR_TEMPLATES + config.Ps() + "bgImg.png"
  icon1 := core.GetTemplateDirPath() + config.Ps() + "listIcon.svg"
  icon2 := core.GetTemplateDirPath() + config.Ps() + "unionIcon.svg"

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
  c, err := v.ParseHexColor(config.EXPORT_BACKGROUND)
  if err != nil {
    logger.L(ctx).Errorf("Failed to parse config hex color %s: %s", config.EXPORT_BACKGROUND, err.Error())
    return nil, perror.New(perror.ERROR_CONFIG_PARSE, "Failed to parse config hex color")
  }

  draw.Draw(canvas, bounds, image.NewUniform(c), image.ZP, draw.Src)
  draw.Draw(canvas, bounds, img, image.Point{0, 0}, draw.Over)

  // Draw Icons
  in1, err := os.Open(icon1) 
  if err != nil {
    logger.L(ctx).Errorf("Failed to read %s: %s", icon1, err.Error())
    return nil, err
  }

  defer in1.Close()

  in2, err := os.Open(icon2) 
  if err != nil {
    logger.L(ctx).Errorf("Failed to read %s: %s", icon2, err.Error())
    return nil, err
  }

  defer in2.Close()

  iconObj1, err := oksvg.ReadIconStream(in1)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read stream1 %s: %s", icon1, err.Error())
    return nil, err
  }

  iconObj2, err := oksvg.ReadIconStream(in2)
  if err != nil {
    logger.L(ctx).Errorf("Failed to read stream2 %s: %s", icon2, err.Error())
    return nil, err
  }

  w1 := int(iconObj1.ViewBox.W)
  h1 := int(iconObj1.ViewBox.H)

  w2 := int(iconObj2.ViewBox.W)
  h2 := int(iconObj2.ViewBox.H)

  iconObj1.SetTarget(0, 0, float64(w1), float64(h1))
  iconObj2.SetTarget(0, 0, float64(w2), float64(w2))

  logger.L(ctx).Debugf("read %dx%d %dx%d", w1, h1, w2, h2)


  rgba1 := image.NewNRGBA(image.Rect(0, 0, w1, h1))
  iconObj1.Draw(rasterx.NewDasher(w1, h1, rasterx.NewScannerGV(w1, h1, rgba1, rgba1.Bounds())), 1)

  rgba2 := image.NewNRGBA(image.Rect(0, 0, w2, h2))
  iconObj2.Draw(rasterx.NewDasher(w2, h2, rasterx.NewScannerGV(w2, h2, rgba2, rgba2.Bounds())), 1)

  draw.Draw(canvas, bounds.Add(image.Pt(24, 325)), rgba1, image.ZP, draw.Over)
  draw.Draw(canvas, bounds.Add(image.Pt(24, 360)), rgba2, image.ZP, draw.Over)

	// Draw texts
  c, err = v.ParseHexColor(config.BRAND_COLOR)
  if err != nil {
    logger.L(ctx).Errorf("Failed to parse config hex color %s: %s", config.BRAND_COLOR, err.Error())
    return nil, perror.New(perror.ERROR_CONFIG_PARSE, "Failed to parse config hex brand color")
  }

  wcolor := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} 
  v.DrawText(ctx, canvas, 5, 20, c, 32, "CC", true, true)
  v.DrawText(ctx, canvas, 5, 85, wcolor, 32, amountStr, true, true)
  v.DrawText(ctx, canvas, 15, 485, wcolor, 32, amountStr, false, true)
  v.DrawText(ctx, canvas, 15, 575, wcolor, 16, "Upload this file to your Skyvault", false, false)
  v.DrawText(ctx, canvas, 15, 598, wcolor, 16, "or POWN it and keep it wherever you want", false, false)
  v.DrawText(ctx, canvas, 15, 635, c, 14, "More info on Cloudcoin.global", false, false)
  v.DrawText(ctx, canvas, 205, 20, c, 32, "CloudCoin", false, true)

  v.DrawText(ctx, canvas, 55, 322, wcolor, 16, strings.ToUpper(textAmount), false, false)

  currentTime := time.Now()
  dateTime := currentTime.Format("01/02/2006")
  v.DrawText(ctx, canvas, 55, 358, wcolor, 16, dateTime, false, false)


  buff := bytes.NewBuffer([]byte{})
  switch fileType {
  case "png":
    png.Encode(buff, canvas)
  default:
    return nil, perror.New(perror.ERROR_UNSUPPORTED_FILE_TYPE, "Template file type " + fileType + " is not supported")
  }

  // Embeds data
  bytes, err := v.InsertPNGData(ctx, buff.Bytes(), data)
  if err != nil {
    return nil, err
  }

  return bytes, nil
}

// Draw a text possibly rotating it
func (v *Exporter) DrawText(ctx context.Context, img *image.NRGBA, x, y int, color color.RGBA, fsize int, text string, isVertical bool, isBold bool) {

  var font *truetype.Font
  if (isBold) {
    font = v.font
  } else {
    font = v.rfont
  }

  c := freetype.NewContext()
  size := float64(fsize)

  ptf := c.PointToFixed(size) >> 6

  fbounds := font.Bounds(ptf)

  gw := float32(fbounds.Max.X - fbounds.Min.X)
  gh := float32(fbounds.Max.Y - fbounds.Min.Y)

  glyphCount := len(text)

  imageWidth := uint32(gw * float32(glyphCount))
  imageHeight := uint32(gh)

  logger.L(ctx).Debugf("image size  %dx%d px", imageWidth, imageHeight)

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

  if isVertical {
    rgba = imaging.Rotate270(rgba)
  }

  b := img.Bounds()
  draw.Draw(img, b.Add(image.Pt(x, y)), rgba, image.ZP, draw.Over)
}

func (v *Exporter) ParseHexColor(s string) (color.RGBA, error) {
  var c color.RGBA

  c.A = 0xFF
  _, err := fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
  
  return c, err
}
