package main

import (
	"bytes"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	StockItem struct {
		Idx   int
		Desc  string
		Value string
	}

	arg struct {
		code  string
		count float64
	}
)

var (
	totalProfit  float64
	totalColors  int
	profitColors int
)

func main() {
	var headers []string
	for _, v := range template {
		headers = append(headers, v.Desc)
	}

	args := []*arg{
		{
			code:  "sz002466",
			count: 2,
		},
		{
			code:  "sh603659",
			count: 1,
		},
		{
			code:  "sz002460",
			count: 1,
		},
		{
			code:  "sz000422",
			count: 3,
		},
		{
			code:  "sz000831",
			count: 4,
		},
		{
			code:  "sh600111",
			count: 0,
		},
		{
			code:  "sh603799",
			count: 1,
		},
		{
			code:  "sh600418",
			count: 7,
		},
		{
			code:  "sh601015",
			count: 6,
		},
		{
			code:  "sh600460",
			count: 1,
		},
	}

	for {
		totalProfit = 0
		fmt.Fprintf(os.Stdout, "\u001b[2K\u001b[0;0H")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader(headers)
		table.SetBorder(true)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.FgHiBlueColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgWhiteColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgCyanColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgMagentaColor, tablewriter.Bold},
			tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold},

			tablewriter.Colors{tablewriter.FgRedColor, tablewriter.Bold},
		)
		table.SetColumnColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgWhiteColor},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgMagentaColor},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgHiRedColor},

			tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		)

		for _, a := range args {
			row, err := stockPrice(a.code, a.count)

			if err != nil {
				log.Println(err)
				return
			}
			row6, err := strconv.ParseFloat(row[6], 64)
			if err != nil {
				return
			}
			if row6 >= 0 {
				profitColors = tablewriter.FgHiRedColor
			} else {
				profitColors = tablewriter.FgHiGreenColor
			}
			table.Rich(row, []tablewriter.Colors{
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{},
				tablewriter.Colors{tablewriter.Normal, profitColors},
				tablewriter.Colors{tablewriter.Normal, profitColors},
				tablewriter.Colors{tablewriter.Normal, profitColors},
				tablewriter.Colors{tablewriter.Normal, profitColors},
			})
			//table.Render()
		}

		if totalProfit > 0 {
			totalColors = tablewriter.FgHiRedColor
		} else {
			totalColors = tablewriter.FgHiGreenColor
		}
		table.SetCaption(true, time.Now().Format("UpdatedAt:"+time.RFC3339))
		table.SetFooter([]string{"", "", "", "", "", "total", fmt.Sprintf("%.2f", totalProfit)})
		table.SetFooterColor(
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{tablewriter.Bold},
			tablewriter.Colors{tablewriter.Bold, totalColors})
		table.Render()

		time.Sleep(time.Duration(5) * time.Second)
	}
}

func stockPrice(stockCode string, count float64) (list []string, err error) {
	url := fmt.Sprintf("http://hq.sinajs.cn/list=%s", stockCode)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	utf8, err := GbkToUtf8(bs)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(utf8), `="`)
	body := parts[1]

	ps := strings.Split(body, ",")

	if len(ps) != 0 {
		ps = ps[:7]
	}

	for _, v := range template {
		list = append(list, ps[v.Idx])
	}
	stringYesterday := ps[2]
	stringNow := ps[3]

	yestarday, _ := strconv.ParseFloat(stringYesterday, 64)
	now, _ := strconv.ParseFloat(stringNow, 64)

	diff := now - yestarday
	percent := diff / yestarday

	list[4] = fmt.Sprintf("%.2f", diff)
	list[5] = fmt.Sprintf("%.2f%%", percent*100)

	profit := yestarday * percent * 100 * count

	totalProfit = totalProfit + profit

	list[6] = fmt.Sprintf("%.2f", profit)

	return

}

var template = []StockItem{
	{
		Idx:   0,
		Desc:  "Name",
		Value: "",
	},
	{
		Idx:   1,
		Desc:  "今开",
		Value: "",
	},
	{
		Idx:   2,
		Desc:  "昨收",
		Value: "",
	},
	{
		Idx:   3,
		Desc:  "最新",
		Value: "",
	},
	{
		Idx:   4,
		Desc:  "涨跌",
		Value: "",
	},
	{
		Idx:   5,
		Desc:  "涨幅",
		Value: "",
	},
	{
		Idx:   6,
		Desc:  "收益",
		Value: "",
	},
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
