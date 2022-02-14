import busio
import digitalio
import board
import adafruit_mcp3xxx.mcp3008 as MCP
from adafruit_mcp3xxx.analog_in import AnalogIn
import time
import math

spi = busio.SPI(clock=board.SCK, MISO=board.MISO, MOSI=board.MOSI)
cs = digitalio.DigitalInOut(board.D5)
mcp = MCP.MCP3008(spi, cs)
chan0 = AnalogIn(mcp, MCP.P0)
chan1 = AnalogIn(mcp, MCP.P1)
chan2 = AnalogIn(mcp, MCP.P2)
chan3 = AnalogIn(mcp, MCP.P3)

def getresistance(adcValue):
  rtdV = (adcValue / 1023) * 3.3
#   print("rtdV = ", rtdV)
  R = ((3.3 * 1000) - (rtdV * 1000)) / rtdV
#   print("R = ", R)
  return R

def gettemp(resistance):
  A = 70.27453460e-3 
  B = -127.0393538e-4
  C = 641.9441691e-7
  tempK = 1 / (A + B *  math.log(resistance) + C * (math.pow(math.log(resistance), 3)))
#   print("Temp K = ", tempK)
  tempC = tempK - 273.15
#   print("Temp C = ", tempC)
  return tempC * 9 / 5 + 32

while True:
  print('Raw ADC Value chan0: ', chan0.value)
#   print('ADC Voltage: ' + str(chan0.voltage) + 'V')  
  res = getresistance(chan0.value)
  print('Resistance chan0: ', res)
#   print('Temp?? chan0: ', gettemp(abs(res)))
#   print('Raw ADC Value chan1: ', chan1.value)
#   print('Raw ADC Value chan2: ', chan2.value)
#   print('Raw ADC Value chan3: ', chan3.value)
#   print('ADC Voltage: ' + str(chan.voltage) + 'V')
  
  time.sleep(1)