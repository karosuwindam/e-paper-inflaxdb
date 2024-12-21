package epaperv2

import "context"

func contextWriteWriteData(ctx context.Context, text []string, size float64) context.Context {
	// return context.WithValue(ctx, strWriteData{}, strWriteData{text, size})
	return context.WithValue(ctx, "strWriteData", strWriteData{text, size})
}

func contextReadWriteData(ctx context.Context) (strWriteData, bool) {
	// data, ok := ctx.Value(strWriteData{}).(strWriteData)
	data, ok := ctx.Value("strWriteData").(strWriteData)
	return data, ok
}
