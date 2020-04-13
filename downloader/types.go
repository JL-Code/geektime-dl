package downloader

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

//URL for url infomation
type URL struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
	Ext  string `json:"ext"`
}

//Stream data
type Stream struct {
	URLs    []URL  `json:"urls"`
	Size    int    `json:"size"`
	Quality string `json:"quality"`
	name    string
}

//Datum download infomation
type Datum struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	IsCanDL bool   `json:"is_can_dl"`

	Streams map[string]Stream `json:"streams"`

	URL string `json:"url"`
}

//Data 课程信息
type Data struct {
	Title string  `json:"title"`
	Type  string  `json:"type"`
	Data  []Datum `json:"articles"`
}

//VideoMediaMap 视频大小信息
type VideoMediaMap struct {
	Size int `json:"size"`
}

//EmptyData empty data list
var EmptyData = make([]Datum, 0)

//PrintInfo print info
func (data *Data) PrintInfo() {

	table := tablewriter.NewWriter(os.Stdout)

	header := []string{"#", "ID", "类型", "名称"}
	for key := range data.Data[0].Streams {
		header = append(header, key)
	}
	header = append(header, "下载")

	table.SetHeader(header)
	table.SetAutoWrapText(false)
	i := 0
	for _, p := range data.Data {
		reg, _ := regexp.Compile(" \\| ")
		title := reg.ReplaceAllString(p.Title, " ")

		isCanDL := ""
		if p.IsCanDL {
			isCanDL = " ✔"
		}

		value := []string{strconv.Itoa(i), strconv.Itoa(p.ID), p.Type, title}

		if len(p.Streams) > 0 {
			for _, stream := range p.Streams {
				value = append(value, fmt.Sprintf("%.2fM", float64(stream.Size)/1024/1024))
			}
		} else {
			for range data.Data[0].Streams {
				value = append(value, " -")
			}
		}

		value = append(value, isCanDL)

		table.Append(value)
		i++
	}
	table.Render()
}

func (stream *Stream) calculateTotalSize() {

	if stream.Size > 0 {
		return
	}

	size := 0
	for _, url := range stream.URLs {
		size += url.Size
	}
	stream.Size = size
}
