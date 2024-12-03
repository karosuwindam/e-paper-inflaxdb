package getprometheus_test

import (
	"context"
	"epaperifdb/controller/getprometheus"
	"fmt"
	"testing"
)

func TestApi(t *testing.T) {
	ctx := context.Background()
	data1day, err := getprometheus.GetPrometheusDays(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(data1day.ConvertCO2(ctx))
	fmt.Println(data1day.ConvertHum(ctx))
	fmt.Println(data1day.ConvertTmp(ctx))
}
