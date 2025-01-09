package getinflux

import "context"

func contextWriteReadUrlData(ctx context.Context, timeAgo interface{}, dataType string) context.Context {
	return context.WithValue(ctx, strReadUrlData{}, strReadUrlData{timeAgo, dataType})
}

func contextReadReadUrlData(ctx context.Context) (strReadUrlData, bool) {
	data, ok := ctx.Value(strReadUrlData{}).(strReadUrlData)
	return data, ok
}

func contextWriteUrl(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, "url", url)
}

func contextReadUrl(ctx context.Context) (string, bool) {
	data, ok := ctx.Value("url").(string)
	return data, ok
}
