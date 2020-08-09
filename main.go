package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/swgloomy/gutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type dataStruct struct {
	Id                 int    `json:"id"`
	Courseid           string `json:"courseid"`
	Bookcode           int    `json:"bookcode"`
	Bookname           string `json:"bookname"`
	Chaptercode        string `json:"chaptercode"`
	Chaptername        string `json:"chaptername"`
	Sectioncode        string `json:"sectioncode"`
	Sectionname        string `json:"sectionname"`
	Sectionaudiourl    string `json:"sectionaudiourl"`
	Articlecode        string `json:"articlecode"`
	Articletextcontent string `json:"articletextcontent"`
	ArticletextHtml    string `json:"articletext_html"`
	Articlestartframe  int    `json:"articlestartframe"`
}

var (
	mysqlDb = gutil.MySqlDBStruct{
		DbUser: "root",
		DbHost: "cdb-0ihp1dux.gz.tencentcdb.com",
		DbPort: 10028,
		DbPass: "tagger_mysql_2020",
		DbName: "taggeropen",
	}
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	tableName := "courseware"
	createTableSqlStr := "create table if not exists %s(id int auto_increment,%s,PRIMARY KEY (`id`));"
	tableColumn := []string{
		"courseid nvarchar(200) default '' not null",
		"bookcode int not null",
		"bookname nvarchar(200) default '' not null",
		"chaptercode nvarchar(200) default '' not null",
		"chaptername nvarchar(200) default '' not null",
		"sectioncode nvarchar(200) default '' not null",
		"sectionname nvarchar(200) default '' not null",
		"sectionaudiourl nvarchar(255) default '' not null",
		"articlecode nvarchar(200) default '' not null",
		"articletextcontent nvarchar(5000) default '' not null",
		"articletext_html nvarchar(5000) default '' not null",
		"articlestartframe int not null",
	}

	var dbs *sql.DB
	_, err := gutil.MySqlSqlExec(dbs, mysqlDb, fmt.Sprintf(createTableSqlStr, tableName, strings.Join(tableColumn, ",")))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer gutil.MySqlClose(dbs)

	excelArrayIn, err := gutil.ReadExcel("./03204A.xlsx")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	var modalArray []dataStruct
	for _, listArray := range *excelArrayIn {
		for index, itemArray := range listArray {
			if index == 0 {
				continue
			}
			bookcode, err := strconv.Atoi(itemArray[2])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			articlestartframe, err := strconv.Atoi(itemArray[17])
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			subscript := existCode(itemArray[6], &modalArray)
			if subscript == -1 {
				modalArray = append(modalArray, dataStruct{
					Id:                 index,
					Courseid:           itemArray[1],
					Bookcode:           bookcode,
					Bookname:           itemArray[3],
					Chaptercode:        itemArray[4],
					Chaptername:        itemArray[5],
					Sectioncode:        itemArray[6],
					Sectionname:        itemArray[7],
					Articlecode:        itemArray[10],
					Articletextcontent: strings.Replace(strings.Replace(itemArray[14], "?", "", -1), "　", "", -1),
					ArticletextHtml:    strings.Replace(strings.Replace(strings.Replace(itemArray[15], "@", "", -1), "　", "", -1), "?", "", -1),
					Articlestartframe:  articlestartframe,
				})
				subscript = len(modalArray) - 1
			}
			path := findContent(fmt.Sprintf("%s.files", itemArray[10]), false)
			if path == "" {
				fmt.Println(itemArray[10])
				continue
			}
			htmlByte, err := ioutil.ReadFile(fmt.Sprintf("%s/slide0001.htm", path))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			docQuery, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlByte))
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			modalArray[subscript].Articletextcontent = docQuery.Text()
			htmlStr, err := docQuery.Html()
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			modalArray[subscript].ArticletextHtml = htmlStr
		}
	}
	sqlStr := fmt.Sprintf("insert into %s(courseid,bookcode,bookname,chaptercode,chaptername,sectioncode,sectionname,sectionaudiourl,articlecode,articletextcontent,articletext_html,articlestartframe) values(?,?,?,?,?,?,?,?,?,?,?,?)", tableName)
	for _, item := range modalArray {
		sqlResult, err := gutil.MySqlSqlExec(dbs, mysqlDb, sqlStr, item.Courseid, item.Bookcode, item.Bookname, item.Chaptercode, item.Chaptername, item.Sectioncode, item.Sectionname, item.Sectionaudiourl, item.Articlecode, item.Articletextcontent, item.ArticletextHtml, item.Articlestartframe)
		if err != nil {
			fmt.Println(sqlStr)
			fmt.Println(item.Articletextcontent)
			fmt.Println(err.Error())
			continue
		}
		rowsAffect, err := sqlResult.RowsAffected()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Println(rowsAffect)
		time.Sleep(100 * time.Millisecond)
	}
}

func existCode(code string, listArrayIn *[]dataStruct) int {
	for index, item := range *listArrayIn {
		if item.Sectioncode == code {
			return index
		}
	}
	return -1
}

func findContent(name string, isRegexp bool) string {
	path := ""
	err := filepath.Walk("./data",
		func(pathStr string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if isRegexp {
				flysnowRegexp := regexp.MustCompile(`[a-zA-Z]{0,}`)
				_ = flysnowRegexp.FindAllStringSubmatch("dyty-1-1.mp3", -1)
			}
			if strings.Split(pathStr, "\\")[len(strings.Split(pathStr, "\\"))-1] == name {
				path = pathStr
			}
			return nil
		})
	if err != nil {
		fmt.Println(err.Error())
		return path
	}
	return path
}
