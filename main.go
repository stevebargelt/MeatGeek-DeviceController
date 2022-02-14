package main

import (
	"fmt"
	"math"
	"strconv"
	"time"

	queue "github.com/stevebargelt/MeatGeek-DeviceController/goqueue"
	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/spi"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
        a := raspi.NewAdaptor()
		adc := spi.NewMCP3008Driver(a)
        RTDs := []*RTD{}
        for i := 0; i<5 ;i++ {
            rtd := new(RTD)
            rtd.title = "A" + strconv.Itoa(i)
            rtd.channel = i 
            // rtd.tempsList = *list.New()
            rtd.resistanceQueue = *queue.New(100)
            // rtd.Lock = &sync.Mutex{}
            RTDs = append(RTDs, rtd)
        }
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
                        //fmt.Printf("%s not connected\n", rtd.title)
                        // rtd.tempsList.PushFront(0)
                        rtd.resistanceQueue.Push(0)
                    } else {
                        var resistance = GetResistance(result)
                        // rtd.tempsList.PushFront(resistance)
                        rtd.resistanceQueue.Push(resistance)
                        //fmt.Printf("%s resistance %f. List LEN = %d\n", rtd.title, resistance, rtd.tempsList.Len())
                    }
                    // for rtd.tempsList.Len() > 100 {
                    //     rtd.Lock.Lock()
                    //     rtd.tempsList.Remove(rtd.tempsList.Back())
                    //     rtd.Lock.Unlock()
                    // }
                } 

                })
                gobot.Every(1000*time.Millisecond, func() {
                    for _, rtd := range RTDs {
                        // rtd.Lock.Lock()
                        
                        // var sum float64 = 0.0
                        //fmt.Printf("%s - len %d \n", rtd.title, rtd.tempsList.Len())
                        // var counter = 0
                        // for res := rtd.tempsList.Front(); res.Value != nil; res = res.Next() {
                        //     counter++
                        //     //fmt.Printf("%s resistance %f | Counter %d List Len %d\n", rtd.title, res.Value, counter, rtd.tempsList.Len())
                        //     if res.Value != nil {
                        //         sum += res.Value.(float64)
                        //     }
                        // }
                        // resAverage := sum / float64(rtd.tempsList.Len())
                        // rtd.Lock.Unlock()
                        // sum := 0.0
                        // queuelen := rtd.resistanceQueue.Len()
                        // for {
                        //     res := rtd.resistanceQueue.Pop()
                        //     if res == nil {
                        //         break
                        //     }
                        //     sum += res.(float64)
                        // }
                        // resAverageQueue := sum / float64(queuelen)
                        resAve := rtd.resistanceQueue.Average()
                        fmt.Printf("resAverageQueue %f\n", resAve)
                        fmt.Printf("%s Temp F %f\n", rtd.title, GetTempFahrenheitFromResistance(resAve) + rtd.tempCorrection)
                    }
                })                
        }
        robot := gobot.NewRobot("mcp3008bot",
                []gobot.Connection{a},
                []gobot.Device{adc},
                work,
        )

        robot.Start()
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

type RTD struct {  
    title string
    channel int
    // tempsList list.List
    resistanceQueue queue.Queue
    tempCorrection float64
    // Lock *sync.Mutex
}
