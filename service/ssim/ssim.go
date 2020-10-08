package ssim 
import (
  "fmt"
  "os"
 // "log"
  "image"
  "image/color"
  "image/jpeg"
  "math"
  "errors"
)

// Default SSIM constants
var (
  L = 255.0
  K1 = 0.01
  K2 = 0.03
  C1 = math.Pow((K1*L), 2.0)
  C2 = math.Pow((K2*L), 2.0)
)

func HandleError(err error) {
  if err != nil {
//    log.Fatal(err)
  }
}

// Given a path to an image file, read and return as
// an image.Image
func ReadImage(fname string) image.Image {
  file, err := os.Open(fname)
  HandleError(err)
  defer file.Close()

  img, _, err := image.Decode(file)
  HandleError(err)
  return img
}

// Convert an Image to grayscale which
// equalize RGB values
func ConvertToGray(originalImg image.Image) image.Image {
  bounds := originalImg.Bounds()
  w, h := dim(originalImg)

  grayImg := image.NewGray(bounds)

  for x := 0; x < w; x++ {
    for y := 0; y < h; y++ {
      originalColor := originalImg.At(x, y)
      grayColor := color.GrayModel.Convert(originalColor)
      grayImg.Set(x, y, grayColor)
    }
  }

  return grayImg
}

// Write an image.Image to a jpg file of quality 100
func writeImage(img image.Image, path string) {
  w, err := os.Create(path+".jpg")
  HandleError(err)
  defer w.Close()

  quality := jpeg.Options{100}
  jpeg.Encode(w, img, &quality)
}

// Convert uint32 R value to a float. The returnng
// float will have a range of 0-255
func getPixVal(c color.Color) float64 {
  r, _, _, _ := c.RGBA()
  return float64(r >> 8)
}

// Helper function that return the dimension of an image
func dim(img image.Image) (w, h int) {
  w, h = img.Bounds().Max.X, img.Bounds().Max.Y
  return
}

// Check if two images have the same dimension
func equalDim(img1, img2 image.Image) (bool,string) {
  w1, h1 := dim(img1)
  w2, h2 := dim(img2)
  return (w1 == w2) && (h1 == h2),fmt.Sprintf("图一:%dx%d,图二:%dx%d",w1,h1,w2,h2)
}

// Given an Image, calculate the mean of its 
// pixel values
func Mean(img image.Image) float64 {
  w, h := dim(img)
  n := float64((w*h)-1)
  sum := 0.0

  for x := 0; x < w; x++ {
    for y := 0; y < h; y++ {
      sum += getPixVal(img.At(x, y))
    }
  }
  m := sum/n
  fmt.Printf("mean:%f\n",m)
  return m
}

// Compute the standard deviation with pixel values of Image
func Stdev(img image.Image) float64  {
  w, h := dim(img)

  n := float64((w*h)-1)
  sum := 0.0
  avg := Mean(img)

  for x := 0; x < w; x++ {
    for y := 0; y < h; y++ {
      pix := getPixVal(img.At(x, y))
      sum += math.Pow((pix - avg), 2.0)
    }
  }
  r := math.Sqrt(sum/n)
  fmt.Printf("stdev:%f\n",r)
  return r 
}

// Calculate the covariance of 2 images
func Covar(img1, img2 image.Image) (c float64,  err error) {
    if b,str := equalDim(img1, img2);!b {
    err = errors.New("两图必须是相同分辨率"+str)
    return 0,err
  }
  avg1 := Mean(img1)
  avg2 := Mean(img2)
  fmt.Printf("avg1:%v,avg2:%v\n",avg1,avg2)
  w, h := dim(img1)
  sum := 0.0
  n := float64((w*h)-1)

  for x := 0; x < w; x++ {
    for y := 0; y < h; y++ {
      pix1 := getPixVal(img1.At(x, y))
      pix2 := getPixVal(img2.At(x, y))
      sum += (pix1 - avg1)*(pix2 - avg2)
    }
  }
  c = sum/n
  return c,nil
}

func Ssim(x, y image.Image) (float64,error) {
  avg_x := Mean(x)
  avg_y := Mean(y)

  stdev_x := Stdev(x)
  stdev_y := Stdev(y)

  cov, err := Covar(x, y)
  if err != nil{
    return 0,err

  } 
  HandleError(err)

  numerator := ((2.0 * avg_x * avg_y) + C1) * ((2.0 * cov) + C2)
  denominator := ( math.Pow(avg_x, 2.0) + math.Pow(avg_y, 2.0) + C1) *
                 ( math.Pow(stdev_x, 2.0) + math.Pow(stdev_y, 2.0) + C2 )

  r := numerator/denominator
  fmt.Printf("ssim:%f\n",r)
  return r,nil
}
/*
func main() {
  img := ConvertToGray(readImage("1242x2208x1.jpg"))
  img2 := ConvertToGray(readImage("1242x2208x3.jpg"))

  c, err := Covar(img, img2)
  HandleError(err)
  
  index := Ssim(img, img2)

   fmt.Printf("AVG   %f\n", mean(img))
  fmt.Printf("STDEV %f\n", stdev(img))
   fmt.Printf("COV   %f\n", c)
  _ = c
  fmt.Printf("\nSSIM = %f\n", index)
}
*/
