package getprometheus

import "context"

func contextWriteReadUrlData(ctx context.Context, timeAgo interface{}) context.Context {
	return context.WithValue(ctx, "timeAgo", timeAgo)
}

func contextReadReadUrlData(ctx context.Context) (interface{}, bool) {
	data, ok := ctx.Value("timeAgo").(interface{})
	return data, ok
}

func contextWriteUrl(ctx context.Context, url string) context.Context {
	return context.WithValue(ctx, "url", url)
}

func contextReadUrl(ctx context.Context) (string, bool) {
	data, ok := ctx.Value("url").(string)
	return data, ok
}

func contextWriteDataName(ctx context.Context, dataname string) context.Context {
	return context.WithValue(ctx, "dataname", dataname)
}

func contextReadDataName(ctx context.Context) (string, bool) {
	data, ok := ctx.Value("dataname").(string)
	return data, ok
}
