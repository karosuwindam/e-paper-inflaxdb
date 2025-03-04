package epaperv2

const (
	REG_DRIVER_OUTPUT_CTL                    = 0x01
	REG_GATE_DRIVER_VOLTAGE_CTL              = 0x03
	REG_SOURCE_DRIVER_VOLTAGE_CTL            = 0x04
	REG_INI_CODE_SETTING_OTP_PROGRAM         = 0x08
	REG_WRITE_REGISTER_FOR_INI_CODE_SETTING  = 0x09
	REG_READ_REGISTER_FOR_INI_CODE_SETTING   = 0x0A
	REG_BOOSTER_SOFT_START_CTL               = 0x0C
	REG_DEEP_SLEEP_MODE                      = 0x10 //スリープモード 1:0 00:Nomal 01:Sleep1 11:sleep2
	REG_DATA_ENTRY_MODE_SETTING              = 0x11 //データの追加方法設定 2:2 on: Y off:X  0:1 Y_X 1:増加 0:減少
	REG_SW_RESET                             = 0x12 //ソフトウェアリセット
	REG_HV_READY_DETECTION                   = 0x14
	REG_VCI_DETECTION                        = 0x15
	REG_TEMP_SENSOR_CTL                      = 0x18
	REG_TEMP_SENSOR_CTL_W                    = 0x1A
	REG_TEMP_SENSER_CTL_R                    = 0x1B
	REG_TEMP_SENSER_CTL_WC                   = 0x1C
	REG_IC_REVISION_READ                     = 0x1F
	REG_MASTER_ACTIVATION                    = 0x20
	REG_DISPLAY_UPDATE_CTL                   = 0x21
	REG_DISPLAY_UPDATE_CTL_2                 = 0x22
	REG_WRITE_RAM_BW                         = 0x24 //書き込み先の色で黒/白
	REG_WRITE_RAM_R                          = 0x26 //書き込み先の色で赤
	REG_READ_RAM                             = 0x27
	REG_VCOM_SENSE                           = 0x28
	REG_VCOM_SENSE_DURATION                  = 0x29
	REG_PROGRAM_VCCOM_OTP                    = 0x2A
	REG_WRITE_VCOM_REGISTER                  = 0x2C
	REG_OTP_REGISTER_READ_FOR_DISPLAY_OPTION = 0x2D
	REG_USER_ID_READ                         = 0x2E
	REG_STATUS_BIT_READ                      = 0x2F
	REG_PROGRAM_WS_OTP                       = 0x30
	REG_LOAD_WS_OTP                          = 0x31
	REG_WITETE_LUT_REGISTER                  = 0x32
	REG_CRC_CALCULATION                      = 0x34
	REG_CRC_STATUS_READ                      = 0x35
	REG_PROGRAM_OPT_SELECTION                = 0x36
	REG_WRITE_REGISTER_FOR_DISPLAY_OPTION    = 0x37
	REG_WRITE_REGSISTER_FOR_USER_ID          = 0x38
	REG_OTP_PROGRAM_MODE                     = 0x39
	REG_BORDER_WAVEFORM_CTL                  = 0x3C
	REG_END_OPTION                           = 0x3F
	REG_READ_RAM_OPTION                      = 0x41
	REG_SET_RAM_X_SE                         = 0x44
	REG_SET_RAM_Y_SE                         = 0x45
	REG_SET_AUTE_WRITE_RED_RAM               = 0x46
	REG_SET_AUTE_WRITE_BW_RAM                = 0x47
	REG_SET_RAM_X_ADDRESS_COUNTER            = 0x4E
	REG_SET_RAM_Y_ADDRESS_COUNTER            = 0x4F
	REG_NOP                                  = 0x7F
)

const (
	CHECKPASS     string = "/sys/class/gpiomem0"
	GPIO_RST      string = "GPIO17"
	GPIO_DC       string = "GPIO25"
	GPIO_CS       string = "GPIO8"
	GPIO_BUSY     string = "GPIO24"
	GPIO_RST_PI5  string = "GPIO588"
	GPIO_DC_PI5   string = "GPIO596"
	GPIO_CS_PI5   string = "GPIO579"
	GPIO_BUSY_PI5 string = "GPIO595"
)
