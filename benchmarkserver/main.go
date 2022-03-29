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
	//"reflect"
	//reflect.TypeOf(t)
	"benchmarkserver/internal/ab"
	"benchmarkserver/internal/record"
	"benchmarkserver/internal/score"
	"github.com/rs/xid"
)

//group情報を読み込む
var groupInfo = map[string]string{}

func readGroupInfo(){
	csvFile, err := os.Open("data/groupInfo.csv")
	if err != nil {
		log.Println("<Debug> can't open data/group.csv : ", err)
	}
	reader := csv.NewReader(csvFile)

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		}
		groupInfo[line[0]] = line[1]
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
	http.HandleFunc("/record", recordHandler)

	//group情報を読み込む
	readGroupInfo()


	log.Println("Listening...")
	// 3000ポートでサーバーを立ち上げる
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		log.Println("<Debug> http.LinstenAndServe(:3000) : ", err)
	}
}

//main画面
func rootHandler(w http.ResponseWriter, r *http.Request) {

	groups := []string{}

	for groupName, _ := range groupInfo {
		groups = append(groups, "<option value='" + groupName + "'>" + groupName + "</option>")
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
	IsNewRecord bool
	Id string
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

	ret.IsNewRecord = false
	ret.Id = guid.String()

	//abコマンドで負荷をかける．計測時間を返す
	ret.Msg, ret.Time = ab.Ab(logfile, guid.String(), url)

	//これまでの最高値を取り出す
	if ret.Msg == "" {
		ret.IsNewRecord, ret.Msg = score.Score(logfile, guid.String(), ret.Time, groupName)
	}

	// 構造体をJSON文字列化する
	jsonBytes, _ := json.Marshal(ret)
	// index.jsに返す
	fmt.Fprint(w, string(jsonBytes))
}

//score.csvに記録する
func recordHandler(w http.ResponseWriter, r *http.Request) {

	//ログファイルを開く
	logfile := logfileOpen()
	defer logfile.Close()

	//POSTデータのフォームを解析
	err := r.ParseForm()
	if err != nil {
		log.Println("<Debug> r.ParseForm : ", err)
	}

	groupName := r.Form["groupName"][0]
	times := r.Form["time"][0]
	id := r.Form["id"][0]

	record.Record(logfile, id, times, groupName)
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
