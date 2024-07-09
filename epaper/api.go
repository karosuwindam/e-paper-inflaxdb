package epaper

type api struct{}

func Init() (*api, error) {
	if err := initEpaper(); err != nil {
		return nil, err
	}
	return &api{}, nil
}

func (a *api) TextPut(x, y int, text []string, size float64) {
	testPut(x, y, text, size)
}

func (a *api) ClearScreen() {
	device.ClearScreen()
}
