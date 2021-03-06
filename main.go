package main

import (
	"errors"
	"fmt"
	"html"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	dc_i2c "github.com/davecheney/i2c"
	queue "github.com/stevebargelt/MeatGeek-DeviceController/goqueue"
	gobot "github.com/stevebargelt/mygobot"
	"github.com/stevebargelt/mygobot/api"
	"github.com/stevebargelt/mygobot/drivers/spi"
	"github.com/stevebargelt/mygobot/platforms/raspi"
)

func main() {

    master := gobot.NewMaster()
    deviceApi := api.NewAPI(master)
    deviceApi.Port = "3000"

    deviceApi.AddHandler(func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q \n", html.EscapeString(r.URL.Path))
    })
    deviceApi.Debug()
    deviceApi.Start()

    a := raspi.NewAdaptor()
    adc := spi.NewMCP3008Driver(a, a)

    i, err := dc_i2c.New(0x27, 1)
    check(err)
    lcd, err := dc_i2c.NewLcd(i, 2, 1, 0, 4, 5, 6, 7, 3)
    check(err)
    lcd.BacklightOn()
    lcd.Clear()
    lcd.Home()
    SetPosition(*lcd, 1, 0)
    fmt.Fprint(lcd, "MeatGeek Temp")
    SetPosition(*lcd, 2, 0)
    fmt.Fprint(lcd, "Line 2")
    SetPosition(*lcd, 3, 0)
    fmt.Fprint(lcd, "Line 3")
    SetPosition(*lcd, 4, 0)
    fmt.Fprint(lcd, "Line 4")

    RTDs := []*RTD{}
    for i := 0; i<5 ;i++ {
        rtd := new(RTD)
        rtd.title = "P" + strconv.Itoa(i)
        rtd.channel = i 
        rtd.resistanceQueue = *queue.New(100)
        RTDs = append(RTDs, rtd)
    }
    RTDs[0].title = "Grill"
    
    // Ahh yes magic numbers. Every RTD circuit can report resistances differently
    // these are the corrected values that I've observed for MY system and circuits. 
    // TODO: Allow these to be env vars or CLI flags. 
    RTDs[0].tempCorrection = -6.0
    RTDs[1].tempCorrection = -8.0
    RTDs[2].tempCorrection = 2.0
    RTDs[3].tempCorrection = -1.0
    RTDs[4].tempCorrection = -5.0

    work := func() {
        gobot.Every(10*time.Millisecond, func() {
        for _, rtd := range RTDs {
            result, err := adc.Read(rtd.channel)

            if err != nil {
                fmt.Println("ERROR: ", err.Error())
            }
            if result <=0 {
                //fmt.Printf("%s not connected. result = %d \n", rtd.title, result)
                rtd.resistanceQueue.Push(0)
            } else {
                var resistance = GetResistance(result)
                rtd.resistanceQueue.Push(resistance)
                //fmt.Printf("%s resistance %f\n", rtd.title, resistance)
            }
        } 

        })
        gobot.Every(5000*time.Millisecond, func() {
            for _, rtd := range RTDs {
                resAve := rtd.resistanceQueue.Average()
                //fmt.Printf("resAverageQueue %f\n", resAve)
                if resAve >0 {
                    rtd.temp = GetTempFahrenheitFromResistance(resAve) + rtd.tempCorrection
                    //fmt.Printf("%s Temp F %f\n", rtd.title,rtd.temp)
                    // SetPosition(*lcd, i, 0)
                    //fmt.Fprint(lcd, rtd.title,"Temp F ", math.Round(rtd.temp))
                } else {
                    //fmt.Printf("%s unplugged\n", rtd.title)
                    //SetPosition(*lcd, i, 0)
                    //fmt.Fprint(lcd, rtd.title, "Unplg")
                }
            }
            left := formatTemp(RTDs[1].title, RTDs[1].temp)
            right := formatTemp(RTDs[2].title, RTDs[2].temp)
            line := justifyWithSpaces(left, right, 20)
            fmt.Println(line)
            SetPosition(*lcd, 1, 0)
            fmt.Fprint(lcd, line)
            left = formatTemp(RTDs[3].title, RTDs[3].temp)
            right = formatTemp(RTDs[4].title, RTDs[4].temp)
            line = justifyWithSpaces(left, right, 20)
            fmt.Println(line)
            SetPosition(*lcd, 2, 0)
            fmt.Fprint(lcd, justifyWithSpaces(left, right, 20))
            SetPosition(*lcd, 3, 0)
            fmt.Fprint(lcd, formatTemp(RTDs[0].title, RTDs[0].temp))
            fmt.Println(formatTemp(RTDs[0].title, RTDs[0].temp))
            SetPosition(*lcd, 4, 0)
            t := time.Now()
            fmt.Fprint(lcd, t.Format("Mon Jan 2 15:04"))
        })                
    }
    robot := gobot.NewRobot("mcp3008bot",
            []gobot.Connection{a},
            []gobot.Device{adc},
            work,
    )
    robot.AddCommand("get_temps", func(params map[string]interface{}) interface{} {
        temps := []float64{}
        for _, rtd := range RTDs {
            temps = append(temps, rtd.temp)
        }
        return temps
    })

    master.AddRobot(robot)
    master.Start()
}

func GetResistance(adcValue int) (float64) {
    var rtdV float64 = (float64(adcValue) / 1023) * 3.3
    R := ((3.3 * 1000) - (rtdV * 1000)) / rtdV
    return R
}

func GetTempFahrenheitFromResistance(resistance float64) (float64) {
    var A float64 = 3.90830e-3 // Coefficient A
    var B float64 = -5.775e-7 // Coefficient B
    var ReferenceResistor float64 = 1000
    var TempCelsius float64 = (-A + math.Sqrt(A * A - 4 * B * (1 - resistance / ReferenceResistor))) / (2 * B);
    return TempCelsius * 9 / 5 + 32;
}

///RTD is a Resisteance Temperature Detector
type RTD struct {  
    title string
    channel int
    resistanceQueue queue.Queue
    tempCorrection float64
    temp float64
}

func check(err error) {
	if err != nil { log.Fatal(err) }
}

// I had to rewrite the SetPosition becuase the original from the Dave Cheney 
// wasn't working with my hardware. I will investigate at some point. For now,
// this works... 
func SetPosition(lcd dc_i2c.Lcd, top, left int) (err error) {
    const CMD_DDRAM_Set = 0x80
    ErrInvalidPosition := errors.New("invalid position value")
    rowOffsets := []int{ 0, 64, 20, 84 }
    rows := 4

    if top < 1 || top > 4 {
		err = ErrInvalidPosition
		return
	}
    if left < 0 || left > 39 {
		err = ErrInvalidPosition
		return
	}
    var newAddress = left + rowOffsets[top-1];
    if left < 0 || (rows == 1 && newAddress >= 80) || (rows > 1 && newAddress >=104) {
        err = ErrInvalidPosition
        return
    }

	lcd.Command(byte(CMD_DDRAM_Set | newAddress))
    return nil
}

func justifyWithSpaces(string1, string2 string, maxChars int) (string) {
    if len(string1) + len(string2) > maxChars {
        if len(string1) > 10 {
            string1 = string1[0:9]
        }
        if len(string2) > 10 {
            string2 = string2[0:9]
        }
    }
    spacesCount := maxChars - len(string2)
    return (fmt.Sprintf("%-*v%v", spacesCount, string1, string2))
}

func formatTemp(title string, temp float64) (string) {
    if temp > 0 {
        return fmt.Sprintf("%s %.0f F", title, math.Round(temp))
    } else {
        return fmt.Sprintf("%s unplg", title)
    }
}