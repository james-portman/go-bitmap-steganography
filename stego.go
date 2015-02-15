package main

import (
        "fmt"
        "os"
        "image"
        "image/color"
        "image/draw"
        "math"
        "math/rand"
        "golang.org/x/image/bmp"
)


func getLsb(img *image.RGBA, x,y int, channel int) int {
    color := img.At(x,y)
    r,g,b,_ := color.RGBA()
    var lsb uint32
    switch channel {
        case 0://r
            lsb = (r&1)
        case 1://g
            lsb = (g&1)
        case 2://b
            lsb = (b&1)
    }
    return int(lsb)
}

func getMessage(img *image.RGBA, seedText string) string {

        seedRandom(seedText)

        b := img.Bounds()

        totalpx := b.Dx() * b.Dy()
        totalbytes := totalpx * 3 / 8

        used := make3dImageArray(b.Dx(),b.Dy())

        var output string

        // for each letter up to total possible bytes
        for h := 0; h < totalbytes; h++ {

            var letter int

                for i := 7; i >= 0; i-- {

                        x,y,c := getNextFreeSpace(img,used)
                        used[x][y][c] = 1

                        bit := getLsb(img,x,y,c)

                        pow := rune(math.Pow(2,float64(i)))
                        letter += (int(pow) * bit)
                }

                // check for 0 terminator
                if (letter == 0) { break; }
                output += string(letter)

        }
        return output
}

func getNextFreeSpace(img *image.RGBA, used [][][]int) (x,y,c int) {
    alreadyUsed := true
    for alreadyUsed {
        x, y, c = randomXYC(img)
        if (used[x][y][c] == 0) {
            alreadyUsed = false
            return x,y,c
        }
    }
    return -1,-1,-1
}

func make3dImageArray(dx,dy int) ([][][]int) {
        used := make([][][]int,dx)
        for i := 0; i < dx; i++ {
                usedy := make([][]int,dy)
                for j := 0; j < dy; j++ {
                        usedcolor := make([]int,3)
                        usedy[j] = usedcolor
                }
                used[i] = usedy
        }
        return used
}

func randomXYC(img *image.RGBA) (int,int,int) {
    b := img.Bounds()
    x := rand.Intn(b.Dx()-1)
    y := rand.Intn(b.Dy()-1)
    c := (rand.Intn(99)&3)
    if (c > 2) { c = 2 }
    return x,y,c
}

func saveMessage(img *image.RGBA,message, seedText string) bool {

        b := img.Bounds()

        totalpx := b.Dx() * b.Dy()
        totalbytes := totalpx * 3 / 8

        if (len(message) > totalbytes) {
            fmt.Println("That message is too big to embed sorry")
            os.Exit(1)
        }

        used := make3dImageArray(b.Dx(),b.Dy())

        seedRandom(seedText)

        for _,value := range message {

                for i := 7; i >= 0; i-- {
                        mul := rune(math.Pow(2,float64(i)))
                        bit := uint32((value&mul)>>uint(i))
                        //fmt.Println(mul,": ",bit)

                        x,y,c := getNextFreeSpace(img,used)
                        used[x][y][c] = 1

                        setLsb(img,x,y,c,bit)
                }
        }

        // terminating 0
        for i := 7; i >= 0; i-- {
                x,y,c := getNextFreeSpace(img,used)
                used[x][y][c] = 1
                setLsb(img,x,y,c,uint32(0))
        }

        return true
}

func seedRandom(seedText string) {
        seed := int64(0)
        for _,value := range seedText {
                seed += int64(value)
        }
        seed = seed * int64(seedText[0])
        //fmt.Println("rand seed: ",seed)
        rand.Seed(seed)
}

func setLsb(img *image.RGBA, x,y int, channel int, set uint32) {
        old := img.At(x,y)
        oldr,oldg,oldb,olda := old.RGBA()

        switch channel {
                case 0://r
                        oldr = (oldr>>1<<1)|set
                case 1://g
                        oldg = (oldg>>1<<1)|set
                case 2://b
                        oldb = (oldb>>1<<1)|set
        }
        newcolor := color.RGBA{uint8(oldr),uint8(oldg),uint8(oldb),uint8(olda)}
        img.Set(x,y,newcolor)
}

func main() {

        // get options/usage
        fmt.Println("[r]ead or [w]rite message? [r]")
        var action string
        _, err := fmt.Scanf("%s", &action)

        fmt.Println("input .bmp filename?")
        var imageFileName string
        _, err = fmt.Scanf("%s", &imageFileName)

        fmt.Println("password?")
        var password string
        _, err = fmt.Scanf("%s", &password)



        bitmapfile, err := os.Open(imageFileName)
        if err != nil {
            fmt.Println(err)
        }
        defer bitmapfile.Close()
        bitmap, err := bmp.Decode(bitmapfile)
        b := bitmap.Bounds()

        totalpx := b.Dx() * b.Dy()
        totalbytes := totalpx * 3 / 8
        fmt.Println("total possible bytes ",totalbytes)

        // copy to new RGBA image
        newimage := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
        draw.Draw(newimage, newimage.Bounds(), bitmap, b.Min, draw.Src)



        if (action == "r" || action == "") {

            output := getMessage(newimage, password)
            fmt.Println("Message read out from image:",output)


        } else if (action == "w") {

            fmt.Println("message?")
            var message string
            _, err = fmt.Scanf("%s", &message)

            if (message == "") {
                fmt.Println("No message given, see -help")
                os.Exit(1)
            }

            if (!saveMessage(newimage,message,password)) {
                fmt.Println("Failed to save the message")
            } else {
                fmt.Println("Message saved into image")
            }


            // save altered image to file
            newFileName := imageFileName+".altered.bmp"
            newfile, _ := os.Create(newFileName)
            defer newfile.Close()
            bmp.Encode(newfile, newimage)

            fmt.Println(newFileName,"saved")


        }

}
