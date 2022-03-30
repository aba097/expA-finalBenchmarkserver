package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"encoding/csv"
	"io"
	"strconv"
	//"reflect"
	//reflect.TypeOf(t)
	"benchmarkserver/internal/ab"
	"benchmarkserver/internal/record"
	"github.com/rs/xid"
)

type GroupInfo struct {
	groupName string
	Pass  string
	Num int
}

//group情報を読み込む
var groupInfo = []GroupInfo{}

func readGroupInfo(){
	groupInfo = []GroupInfo{}
	csvFile, err := os.Open("data/groupInfo.csv")
	if err != nil {
		log.Println("<Debug> can't open data/groupInfo.csv : ", err)
	}
	reader := csv.NewReader(csvFile)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		num, err := strconv.Atoi(line[2])
		if err != nil {
			log.Println("<Debug> Not the numbers data/groupInfo.csv : ", err)
		}
		groupInfo = append(groupInfo, GroupInfo{line[0], line[1], num})
	}
}

func writeGroupInfo(groupName string){
	recordData := ""
	for _, groupinfo := range groupInfo {
		if groupinfo.groupName == groupName {
			groupinfo.Num--
		}
		recordData += groupinfo.groupName + "," + groupinfo.Pass + "," + strconv.Itoa(groupinfo.Num) + "\n"
	}

	log.Println(recordData)

	//ファイル書き込み
	file, err := os.Create("data/groupInfo.csv")
	if err != nil{
		log.Println("<Debug> can't open or create data/groupInfo.csv : ", err)
	}
	defer file.Close()
	_, err = file.WriteString(recordData)
	if err != nil {
		log.Println("<Debug> cant' write data/groupInfo.csv : ", err)
	}

}

func main() {
	// webフォルダにアクセスできるようにする
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css/"))))
	http.Handle("/script/", http.StripPrefix("/script/", http.FileServer(http.Dir("./web/script/"))))
	http.Handle("/gif/", http.StripPrefix("/gif/", http.FileServer(http.Dir("./web/gif/"))))

	//ルーティング設定 "/"というアクセスがきたら rootHandlerを呼び出す
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/measure", measureHandler)

	log.Println("Listening...")
	// 3000ポートでサーバーを立ち上げる
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Println("<Debug> http.LinstenAndServe(:3000) : ", err)
	}
}

//main画面
func rootHandler(w http.ResponseWriter, r *http.Request) {

	//group情報を読み込む
	readGroupInfo()

	groups := []string{}

	for _, groupinfo := range groupInfo {
		groups = append(groups, "<option value='" + groupinfo.groupName + "'>" + groupinfo.groupName + ", 残り" + strconv.Itoa(groupinfo.Num) + "回" + "</option>")
	}

	//index.htmlを表示させる
	tmpl := template.Must(template.ParseFiles("./web/html/index.html"))
	err := tmpl.Execute(w, groups)
	if err != nil {
		log.Println("<Debug> can't open ./web/html/index.htm : ", err)
	}
}

// ajax戻り値のJSON用構造体
type measureParam struct {
	Time string
	Msg  string
}

//フォームからの入力を処理 index.jsから受け取る
func measureHandler(w http.ResponseWriter, r *http.Request) {

	//ログファイルを開く
	logfile := logfileOpen()
	defer logfile.Close()

	//index.jsに返すJSONデータ変数
	var ret measureParam
	//POSTデータのフォームを解析
	err := r.ParseForm()
	if err != nil {
		log.Println("<Debug> r.ParseForm : ", err)
	}

	url := r.Form["url"][0]
	groupName := r.Form["groupName"][0]

	//idを設定(logを対応づけるため)
	guid := xid.New()
	log.Println("<Info> request URL: " + url + ", GroupName: " + groupName + ", id: " + guid.String())
	fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05")+"<Info> request URL: "+url+", GroupName: "+groupName+", id: "+guid.String())


	//まだ計測回数があるか
	var canMeasure = true
	for _, groupinfo := range groupInfo {
		if groupinfo.groupName == groupName {
			if groupinfo.Num == 0 {
				canMeasure = false
			}
			break
		}
	}

	if canMeasure {
		//abコマンドで負荷をかける．計測時間を返す
		ret.Msg, ret.Time = ab.Ab(logfile, guid.String(), url)

		//正常に計測終了したら記録する
		if ret.Msg == "" {
			record.Record(logfile, guid.String(), ret.Time, groupName)
			ret.Msg = "計測完了"

			//計測回数を減らす
			writeGroupInfo(groupName)
		}
	}else{
		ret.Time = "0.00"
		ret.Msg = "計測回数の上限を超えています"
	}

	// 構造体をJSON文字列化する
	jsonBytes, _ := json.Marshal(ret)
	// index.jsに返す
	fmt.Fprint(w, string(jsonBytes))
}

//ログファイルを開く，ログファイルをgithubにpushする
func logfileOpen() *os.File {

	//ログファイルを開く(logを記録するファイル)
	logfile, err := os.OpenFile("data/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("<Debug> can't open data/log.txt : ", err)
	}
	return logfile
}
