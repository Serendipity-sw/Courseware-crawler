package main

import (
	"encoding/json"
	"fmt"
	"github.com/swgloomy/gutil"
	"strings"
	"sync"
)

type zhangStruct struct {
	ChapterDiscount float64            `json:"chapterDiscount"`
	ChapterOriginal float64            `json:"chapterOriginal"`
	IsFree          int                `json:"isFree"`
	ChapterDesc     string             `json:"chapterDesc"`
	ChapterName     string             `json:"chapterName"`
	CourseId        int                `json:"courseId"`
	StatusId        int                `json:"statusId"`
	Id              int                `json:"id"`
	PeriodList      []periodListStruct `json:"periodList"`
}

type periodListStruct struct {
	Id             int     `json:"id"`
	StatusId       int     `json:"statusId"`
	CourseId       int     `json:"courseId"`
	ChapterId      int     `json:"chapterId"`
	PeriodName     string  `json:"periodName"`
	PeriodDesc     string  `json:"periodDesc"`
	IsFree         int     `json:"isFree"`
	PeriodOriginal float64 `json:"periodOriginal"`
	PeriodDiscount float64 `json:"periodDiscount"`
	CountBuy       int     `json:"countBuy"`
	CountStudy     int     `json:"countStudy"`
	IsDoc          int     `json:"isDoc"`
	DocName        string  `json:"docName"`
	DocUrl         string  `json:"docUrl"`
	IsVideo        int     `json:"isVideo"`
	VideoNo        int     `json:"videoNo"`
	VideoName      string  `json:"videoName"`
	VideoLength    string  `json:"videoLength"`
	VideoVid       string  `json:"videoVid"`
	VideoOasId     string  `json:"videoOasId"`
}

func main() {
	dirPath := "./excelData"
	fileNameArrayIn, err := gutil.GetMyAllFileByDir(dirPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var threadLock sync.WaitGroup
	for index, fileName := range *fileNameArrayIn {
		threadLock.Add(1)
		go excelProcess(fmt.Sprintf("%s/%s", dirPath, fileName), index, fmt.Sprintf("data%d", index), &threadLock)
	}
	threadLock.Wait()
	fmt.Println("run success!")
}

func excelProcess(excelPath string, id int, fileNam string, threadLock *sync.WaitGroup) {
	defer threadLock.Done()
	excelArrayIn, err := gutil.ReadExcel(excelPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var dataList []zhangStruct
	for _, listArray := range *excelArrayIn {
		for index, itemArray := range listArray {
			if index == 0 {
				continue
			}
			if len(itemArray) > 2 {
				index = dataProcess(dataList, itemArray[3])
				if index == -1 {
					dataList = append(dataList, zhangStruct{
						ChapterDiscount: 0.0,
						ChapterOriginal: 0.0,
						IsFree:          1,
						ChapterDesc:     "",
						ChapterName:     itemArray[3],
						CourseId:        1,
						StatusId:        1,
						Id:              id,
					})
					index = len(dataList) - 1
				}
				two := jieProcess(dataList, itemArray[3], itemArray[5])
				if two == -1 {
					dataList[index].PeriodList = append(dataList[index].PeriodList, periodListStruct{
						Id:             1,
						StatusId:       1,
						CourseId:       id,
						ChapterId:      2,
						PeriodName:     itemArray[5],
						PeriodDesc:     strings.Replace(strings.Replace(itemArray[14], "?", "", -1), "　", "", -1),
						IsFree:         1,
						PeriodOriginal: 0.0,
						PeriodDiscount: 0.0,
						CountBuy:       20,
						CountStudy:     40,
						IsDoc:          0,
						DocName:        "文档1",
						DocUrl:         "文档地址",
						IsVideo:        0,
						VideoNo:        1231231,
						VideoName:      "视频名称",
						VideoLength:    "12:30",
						VideoVid:       "12121",
						VideoOasId:     "1212",
					})
				} else {
					dataList[index].PeriodList[two].PeriodDesc += strings.Replace(strings.Replace(itemArray[14], "?", "", -1), "　", "", -1)
				}
			}
		}
	}
	jsonByte, err := json.Marshal(dataList)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	gutil.FileCreateAndWrite(&jsonByte, fmt.Sprintf("./excelData/%s", fileNam), false)
}

func dataProcess(dataList []zhangStruct, name string) int {
	for index, item := range dataList {
		if item.ChapterName == name {
			return index
		}
	}
	return -1
}

func jieProcess(dataList []zhangStruct, name string, jie string) int {
	for _, item := range dataList {
		if item.ChapterName == name {
			for index, modal := range item.PeriodList {
				if modal.PeriodName == jie {
					return index
				}
			}
		}
	}
	return -1
}
