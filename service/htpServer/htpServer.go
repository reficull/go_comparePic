package htpServer

import (
	"flag"
    "fmt"
//    "log"
//    "bytes"
    "strings"
    "os"
    "io"
    "net/http"
    //    "os"
    "encoding/json"
    //"../ssim"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

var (
	useAll            = flag.Bool("all", false, "Compute all hashes")
	usePHash          = flag.Bool("phash", false, "Compute PHash")
	useAverage        = flag.Bool("average", false, "Compute AverageHash")
	useBlockMean0     = flag.Bool("blockmean0", false, "Compute BlockMeanHash mode 0")
	useBlockMean1     = flag.Bool("blockmean1", false, "Compute BlockMeanHash mode 1")
	useColorMoment    = flag.Bool("colormoment", false, "Compute ColorMomentHash")
	useMarrHildreth   = flag.Bool("marrhildreth", false, "Compute MarrHildrethHash")
	useRadialVariance = flag.Bool("radialvariance", false, "Compute RadialVarianceHash")
)

type CommandType int

const (
    GetCommand = iota
    SetCommand
    IncCommand
    UFCommand
)

type Command struct {
    ty        CommandType
    name      string
    str1    string
    str2    string
    val       float32
    eng     string 
    replyChan chan string
}

type Res struct{
    Err string
    Hashes map[string]float64
}
type Server struct {
    Cmds chan<- Command
}

func setupHashes() []contrib.ImgHashBase {
	var hashes []contrib.ImgHashBase

	if *usePHash || *useAll {
		hashes = append(hashes, contrib.PHash{})
	}
	if *useAverage || *useAll {
		hashes = append(hashes, contrib.AverageHash{})
	}
	if *useBlockMean0 || *useAll {
		hashes = append(hashes, contrib.BlockMeanHash{})
	}
	if *useBlockMean1 || *useAll {
		hashes = append(hashes, contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1})
	}
	if *useColorMoment || *useAll {
		hashes = append(hashes, contrib.ColorMomentHash{})
	}
	if *useMarrHildreth || *useAll {
		// MarrHildreth has default parameters for alpha/scale
		hashes = append(hashes, contrib.NewMarrHildrethHash())
	}
	if *useRadialVariance || *useAll {
		// RadialVariance has default parameters too
		hashes = append(hashes, contrib.NewRadialVarianceHash())
	}

	// If no hashes were selected, behave as if all hashes were selected
	if len(hashes) == 0 {
		*useAll = true
		return setupHashes()
	}

	return hashes
}

func StartProcessManager(initvals map[string]float64) chan<- Command {
    counters := make(map[string]float64)
    for k, v := range initvals {
        counters[k] = v
    }
    cmds := make(chan Command)
    go func() {
        for cmd := range cmds {
            switch cmd.ty {
            case UFCommand:
                var ret string
                fmt.Println("form:%v",cmd)
                cmd.replyChan <- ret
            default:
                fmt.Println("unknown command type",cmd.ty)
                //log.Fatal("unknown command type", cmd.ty)
            }
        }
    }()
    return cmds
}

func makeRes(str string,hashes map[string]float64) string{
    response := &Res{}
    response.Err= str  
    response.Hashes =hashes 
    ret ,err := json.Marshal(response)
    if err != nil{
        fmt.Printf("json make fail:%s",err)
    }
    return string(ret)

}

func (s *Server) UF(w http.ResponseWriter, r *http.Request) {

    err := r.ParseMultipartForm(32 << 20) // limit your max input length 32MB!
    if err != nil{
        fmt.Fprintln(w, err)
        return
    } 

    formData := r.MultipartForm

    files := formData.File["files"]
    if files == nil{
        fmt.Println("no files parsed")
    }
    if len(files) < 2{
        ret := makeRes("need 2 files to compare",nil) 
        fmt.Fprintf(w,ret)
        return
    }
    var file1 string
    var file2 string
    //fmt.Printf("get file:%v",files)
    for i, _ := range files { // loop through the files one by one
        file, err := files[i].Open()
        defer file.Close()
        if err != nil {
            fmt.Fprintln(w, err)
            return
        }
        if i == 0{
            file1 = "/tmp/" + files[i].Filename 
        }else if i==1{
            file2 = "/tmp/" + files[i].Filename 
        }
        out, err := os.Create("/tmp/" + files[i].Filename)
        
        defer out.Close()
        if err != nil {
            fmt.Fprintf(w, "Unable to create the file for writing. Check your write access privilege")
            return
        }
        fmt.Printf("get file:i:%d,name:%s\n",i,files[i].Filename)

        _, err = io.Copy(out, file) // file not files[i] !

        if err != nil {
            fmt.Fprintln(w, err)
            return
        }
    }

    inputs := []string{file1,file2}

	images := make([]gocv.Mat, len(inputs))

	for i := 0; i < 2; i++ {
		img := gocv.IMRead(inputs[i], gocv.IMReadColor)
		if img.Empty() {
			fmt.Printf("cannot read image %s\n", inputs[i])
			return
		}
		defer img.Close()

		images[i] = img
	}

	// construct all of the hash types in a list. normally, you'd only use one of these.
	hashes := setupHashes()
    resHashes := make(map[string]float64,0)

	// compute and compare the images for each hash type
	for _, hash := range hashes {
		results := make([]gocv.Mat, len(images))

		for i, img := range images {
			results[i] = gocv.NewMat()
			defer results[i].Close()
			hash.Compute(img, &results[i])
			if results[i].Empty() {
				fmt.Printf("error computing hash for %s\n", inputs[i])
//				return
			}
		}

		// compare for similarity; this returns a float64, but the meaning of values is
		// unique to each algorithm.
		similar := hash.Compare(results[0], results[1])

		// make a pretty name for the hash
		name := strings.TrimPrefix(fmt.Sprintf("%T", hash), "contrib.")
		fmt.Printf("%s: similarity %g\n", name, similar)
        resHashes[name] = similar
        // print hash result for each image
        /*
        for i, path := range inputs {
            fmt.Printf("\t%s = %x\n", path, results[i].ToBytes())
        }
        */
    }
    ret := makeRes("ok",resHashes) 
    //fmt.Println(contents)
    // I reset the buffer in case I want to use it again
    // reduces memory allocations in more intense projects
    fmt.Fprintf(w,ret)
    return 

}
/*
    img := ssim.ConvertToGray(ssim.ReadImage(file1))
    img2 := ssim.ConvertToGray(ssim.ReadImage(file2))
    index,err := ssim.Ssim(img, img2)
    if err != nil{
        ret := makeRes(err.Error(),0) 
        fmt.Fprintf(w,ret)
        return 
    }
    //indexFix := (index + 1) * 0.5
    if index < 0{
        index = 0
    }
    ret := makeRes("ok",index) 
    fmt.Printf("res:%s",ret)
    fmt.Fprintf(w,ret)
    */
