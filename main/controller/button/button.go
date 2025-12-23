package button

import (
	"os"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

type KEY struct {
	key1       gpio.PinIO
	key2       gpio.PinIO
	key3       gpio.PinIO
	key4       gpio.PinIO
	key_status [4]bool
}

func InitKey() KEY {
	if checkGpioMemo0() {
		return InitKeyPins_PI5()
	}
	return InitKeyPins()
}

func (k *KEY) KeyInit() {
	for i := 0; i < len(k.key_status); i++ {
		k.key_status[i] = false
	}
}

func (k *KEY) Read(number int) bool {
	return k.key_status[number]
}

func (k *KEY) KeyInitNum(number int) {
	k.key_status[number] = false
}

func (k *KEY) KeyRead() {
	if !k.Read1() {
		k.key_status[0] = true
	}
	if !k.Read2() {
		k.key_status[1] = true
	}
	if !k.Read3() {
		k.key_status[2] = true
	}
	if !k.Read4() {
		k.key_status[3] = true
	}
}

func InitKeyPins() KEY {
	key := KEY{}
	key.key1 = gpioreg.ByName(GPIO_KEY1)
	key.key2 = gpioreg.ByName(GPIO_KEY2)
	key.key3 = gpioreg.ByName(GPIO_KEY3)
	key.key4 = gpioreg.ByName(GPIO_KEY4)

	key.key1.Out(true)
	key.key2.Out(true)
	key.key3.Out(true)
	key.key4.Out(true)
	return key
}
func InitKeyPins_PI5() KEY {
	key := KEY{}
	key.key1 = gpioreg.ByName(GPIO_KEY1_PI5)
	key.key2 = gpioreg.ByName(GPIO_KEY2_PI5)
	key.key3 = gpioreg.ByName(GPIO_KEY3_PI5)
	key.key4 = gpioreg.ByName(GPIO_KEY4_PI5)

	key.key1.Out(true)
	key.key2.Out(true)
	key.key3.Out(true)
	key.key4.Out(true)

	return key
}

func (k *KEY) Read1() bool {
	return k.key1.Read() == gpio.High
}

func (k *KEY) Read2() bool {
	return k.key2.Read() == gpio.High
}
func (k *KEY) Read3() bool {
	return k.key3.Read() == gpio.High
}
func (k *KEY) Read4() bool {
	return k.key4.Read() == gpio.High
}

// フォルダの存在確認
func checkGpioMemo0() bool {
	if _, err := os.Stat(CHECKPASS); os.IsNotExist(err) {
		return false
	}
	return true
}
