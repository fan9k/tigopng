package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"tigopng/id"
	"tigopng/memo"
	"time"

	"github.com/golang/glog"
	"github.com/nfnt/resize"
)

var (
	root       string // 输入文件路径
	outputPath string // 输出文件路径
	outputIs   int    // 是否维持原文件名
	width      int    // 输出图像宽度
	quality    int    // 输出图像质量
)

type X struct {
	img  image.Image
	name string
}

// get file path.
func retrieveData(root string) (value chan string, err chan error) {
	err = make(chan error, 1)
	value = make(chan string)
	go func() {
		defer close(value)
		err <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}
			value <- path
			return nil
		})
	}()
	return
}

// ReceiveData get file and send the file to a channel.
func ReceiveData(file chan string, value chan io.Reader, wg *sync.WaitGroup) {
	for v := range file {
		dif, err := memo.MemoDiff()
		if err != nil {
			fmt.Println(err)
		}
		if dif > 0.2 {
			time.Sleep(time.Second >> 1)
			fmt.Println("waiting for memory less.")
		}
		fi, err := os.Open(v)
		if err != nil {
			fmt.Println(err)
		} else {
			value <- fi
		}
	}
	wg.Done()
}

func DataProcessing(root string, outputFile string, wid int, q int) {
	reader := make(chan io.Reader)
	b := make(chan *X)
	c := make(chan *X)
	value, err := retrieveData(root)
	// wg0 获取文件路径
	wg0 := new(sync.WaitGroup)
	wg0.Add(2)
	for i := 0; i < 2; i++ {
		mark(i, "Getting the path of files: ")
		go ReceiveData(value, reader, wg0)
	}
	//
	go func() {
		wg0.Wait()
		close(reader)
	}()
	// wg1 解析文件名及后缀
	wg1 := new(sync.WaitGroup)
	wg1.Add(32)
	for i := 0; i < 32; i++ {
		go func(i int) {
			defer wg1.Done()
			mark(i, "Analysing...")
			for r := range reader {
				v, ok := r.(*os.File)
				if !ok {
					glog.Errorln("Not a photo")
				}
				_, fname := filepath.Split(v.Name())
				// fname 是文件名，name 是文件后缀格式
				name := findName(fname)
				if name == "" && fname != ".DS_Store" {
					glog.Errorln("Not a file, the filename is ", name)
				}
				img, err := isJpg(name, r)
				if err != nil {
					glog.Errorln(err)
				} else {
					b <- &X{
						img:  img,
						name: fname,
					}
				}
			}
		}(i)
	}
	go func() {
		wg1.Wait()
		close(b)
	}()
	// wg2 压缩图片
	wg2 := new(sync.WaitGroup)
	wg2.Add(32)
	for i := 0; i < 32; i++ {
		go func(i int) {
			mark(i, "Compressing...")
			defer wg2.Done()
			for i := range b {
				i.img = resize.Resize(uint(wid), 0, i.img, resize.NearestNeighbor)
				c <- i
			}
		}(i)
	}
	go func() {
		wg2.Wait()
		close(c)
	}()
	//
	wg3 := new(sync.WaitGroup)
	wg3.Add(32)
	for i := 0; i < 32; i++ {
		go func(i int) {
			mark(i, "Creating files...")
			defer wg3.Done()
			for i := range c {
				defaultName := ""
				if outputIs == 0 {
					defaultName = i.name
				} else {
					defaultName = onlyID1() + ".jpeg"
				}
				file, err := os.Create(outputFile + "/" + defaultName)
				defer file.Close()
				stat, _ := file.Stat()
				fmt.Println("Output success:\n", stat.Name())
				if err != nil {
					fmt.Println(err)
				}
				if q < 20 {
					q = 20
				}
				if err := jpeg.Encode(file, i.img, &jpeg.Options{q}); err != nil {
					glog.Errorln("Photo creating process error:", err)
				}
			}
		}(i)
	}
	//
	if er := <-err; er != nil {
		fmt.Println("can not find file or no order to find", er)
	}
	//
	wg3.Wait()
}

func onlyID() string {
	snow, err := id.NewSnowFlake(1)
	if err != nil {
		glog.Error(err)
	}
	glog.V(1).Info("use snowFlake")
	return strconv.FormatInt(snow.GetID(), 10)
}

func onlyID1() string {
	u, err := id.NewUUID(id.Version1, nil)
	if err != nil {
		glog.Error(err)
	}
	glog.V(1).Info("use UUID")
	return u.String()
}

func findName(name string) string {
	name = strings.ToLower(name)
	v := name[len(name)-4:]
	v1 := name[len(name)-3:]
	if v == "jpeg" {
		return v
	}
	if v1 == "jpg" || v1 == "png" || v1 == "gif" {
		return v1
	}
	return ""
}

func isJpg(name string, r io.Reader) (image.Image, error) {
	switch name {
	case "jpeg":
		return jpeg.Decode(r)
	case "jpg":
		return jpeg.Decode(r)
	case "png":
		return png.Decode(r)
	case "gif":
		return gif.Decode(r)
	default:
		return nil, fmt.Errorf("Note: only support .jpg .jpeg .png .gif format, output format is jpeg.")
	}
}

func mark(i int, name string) {
	if i == 0 {
		fmt.Printf("%s\n", name)
	}
}

func init() {
	flag.StringVar(&root, "r", "./image", "This is the path where you put you images, support subdir, all images will be found.")
	flag.StringVar(&outputPath, "o", ".", "This is the path of output images.")
	flag.IntVar(&width, "w", 0, "The width of output images, original size default, unit is px.")
	flag.IntVar(&quality, "q", 80, "The quality of output images, range from 1 to 100, 80 default.")
	flag.IntVar(&outputIs, "i", 0, "Original and output images keep the same name? 0:Yes, 1:No. If choose no, will use unique id as name.")
	flag.Parse()
}

func main() {
	fmt.Println("Tigopng⛵️")
	fmt.Println("---------------------------")
	DataProcessing(root, outputPath, width, quality)
	fmt.Println("---------------------------")
	fmt.Println("Compressed success")
	fmt.Println("Check them here\n", outputPath)
}
