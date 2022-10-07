package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"time"
)

type infxseries struct {
	Name    string        `json:"name"`
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
}

type infxresults struct {
	Statement_id int          `json:"statement_id"`
	Series       []infxseries `json:"series"`
}

type infxdata struct {
	Results []infxresults `json:"results"`
}

type dataformat struct {
	max float64
	avg float64
	min float64
}

type tmpNum struct {
	num float64
	c   int
}

func ckdata(data []float64) []float64 {
	tData := dataformat{}
	for i, num := range data {
		tData.avg = (tData.avg*float64(i) + num) / float64(i+1)
		if i == 0 {
			tData.max = num
			tData.min = num
		}
		if tData.max < num {
			tData.max = num
		}
		if tData.min > num {
			tData.min = num
		}
	}
	var out []float64
	var cktmp []tmpNum
	for _, num := range data {
		tmp := math.Log10((num - tData.min + 1) / (tData.avg - tData.min + 1))
		tmp = math.Ceil(math.Abs(tmp) * 2)
		tt := tmpNum{num: num, c: int(tmp)}
		cktmp = append(cktmp, tt)
	}
	sort.Slice(cktmp, func(i, j int) bool { return cktmp[i].c < cktmp[j].c })
	count := 0
	for i := 0; i < len(cktmp); i++ {
		if i < len(cktmp)*3/5 {
			count = cktmp[i].c
		}
		if count == cktmp[i].c {
			out = append(out, cktmp[i].num)
		}

	}

	return out
}

func influxdbday(dayago int, datatype string) dataformat {
	urldata := "http://" + IP + ":" + Port + "/query?db=" + DB
	t := time.Now().Add(-time.Hour * 24 * time.Duration(dayago))
	aTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	body := "SELECT * FROM " + TABLE + " WHERE time <= '" + aTime.UTC().Format("2006-01-02T15:04:05Z") + "' AND time >= '" + aTime.Add(-time.Hour*24).UTC().Format("2006-01-02T15:04:05Z") + "' AND type='" + datatype + "'"
	// body := "SELECT * FROM senser_data WHERE time BETWEEN '" + aTime.UTC().Format("2006-01-02T15:04:05Z") + "' AND '" + aTime.Add(-time.Hour*24).UTC().Format("2006-01-02T15:04:05Z") + "' AND type='" + datatype + "'"
	urldata += "&q=" + url.QueryEscape(body)

	req, err := http.NewRequest("GET",
		urldata,
		nil,
	)
	if err != nil {
		return dataformat{}
	}

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var jsondata infxdata
	json.Unmarshal(byteArray, &jsondata)
	if len(jsondata.Results[0].Series) == 0 {
		return dataformat{}
	}
	tmpdata := jsondata.Results[0].Series[0]
	vid := 0
	for ; vid < len(tmpdata.Columns); vid++ {
		if tmpdata.Columns[vid] == "value" {
			break
		}
	}
	var tdata []float64
	for _, tmp := range tmpdata.Values {
		v := reflect.ValueOf(tmp).Index(vid).Interface()
		tdata = append(tdata, v.(float64))
	}
	ave := 0.0
	max := 0.0
	min := 0.0
	tdata = ckdata(tdata)
	for i, num := range tdata {
		ave = (ave*float64(i) + num) / float64(i+1)
		if i == 0 {
			max = num
			min = num
		}
		if max < num {
			max = num
		}
		if min > num {
			min = num
		}
	}

	return dataformat{max: max, avg: math.Round(ave*10) / 10, min: min}

}

func influxdbBack(backdate time.Duration, datatype string) dataformat {
	// urldata := "http://192.168.0.6:8086/query?db=senser"
	urldata := "http://192.168.0.6:8086/query?db=senser"
	// datatype := "co2"
	// datatype := "tmp"
	body := "SELECT * FROM senser_data WHERE time >= '" + time.Now().Add(-backdate).UTC().Format("2006-01-02T15:04:05Z") + "' AND type='" + datatype + "'"
	urldata += "&q=" + url.QueryEscape(body)
	// values := url.Values{}
	// // values.Set("db", "senser")
	// values.Set("q", body)

	req, err := http.NewRequest("GET",
		urldata,
		// strings.NewReader(values.Encode()),
		nil,
	)
	if err != nil {
		return dataformat{}
	}

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var jsondata infxdata
	json.Unmarshal(byteArray, &jsondata)
	// fmt.Println(string(byteArray))
	if len(jsondata.Results[0].Series) == 0 {
		return dataformat{}
	}
	tmpdata := jsondata.Results[0].Series[0]
	vid := 0
	for ; vid < len(tmpdata.Columns); vid++ {
		if tmpdata.Columns[vid] == "value" {
			break
		}
	}
	var tdata []float64
	for _, tmp := range tmpdata.Values {
		v := reflect.ValueOf(tmp).Index(vid).Interface()
		tdata = append(tdata, v.(float64))
	}
	ave := 0.0
	max := 0.0
	min := 0.0
	tdata = ckdata(tdata)
	for i, num := range tdata {
		ave = (ave*float64(i) + num) / float64(i+1)
		if i == 0 {
			max = num
			min = num
		}
		if max < num {
			max = num
		}
		if min > num {
			min = num
		}
	}

	return dataformat{max: max, avg: math.Round(ave*10) / 10, min: min}

}
