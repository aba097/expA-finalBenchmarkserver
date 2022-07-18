package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
	"encoding/csv"
	"io"
	"container/list"
	"strconv"
	"regexp"
	//"reflect"
	//reflect.TypeOf(t)
	"benchmarkserver/internal/ab"
	"benchmarkserver/internal/abtest"
	"benchmarkserver/internal/record"
	"benchmarkserver/internal/score"
	"benchmarkserver/internal/uploadaws"
	"golang.org/x/net/websocket"
	"bufio"
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
	//ここfor回してるけどもっといい方法がありそう
	for _, groupinfo := range groupInfo {
		if groupinfo.groupName == groupName {
			groupinfo.Num--
		}
		recordData += groupinfo.groupName + "," + groupinfo.Pass + "," + strconv.Itoa(groupinfo.Num) + "\n"
	}

	//group情報を書き込み
	file, err := os.Create("data/groupInfo.csv")
	if err != nil{
		log.Println("<Debug> can't open or create data/groupInfo.csv : ", err)
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(recordData)
	if err != nil {
		log.Println("<Debug> cant' write data/groupInfo.csv : ", err)
	}
	writer.Flush()

}

func main() {
	// webフォルダにアクセスできるようにする
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./web/css/"))))
	http.Handle("/script/", http.StripPrefix("/script/", http.FileServer(http.Dir("./web/script/"))))
	http.Handle("/image/", http.StripPrefix("/image", http.FileServer(http.Dir("./web/image/"))))

	//ルーティング設定 "/"というアクセスがきたら rootHandlerを呼び出す
	http.HandleFunc("/", rootHandler)

	http.Handle("/ws", websocket.Handler(measureHandler))
	http.HandleFunc("/record", recordHandler)

	//group情報を読み込む
	//group情報は変化するため，rootHandlerで読み込む
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

	//index.htmlに載せるデータを用意する
	type indexDataElem struct {
		Groups []string
		Que  string
	}

	var indexData indexDataElem
	indexData.Que = strconv.Itoa(que.Len())

	//group情報をhtmlに埋め込む
	for _, groupinfo := range groupInfo {
		//if groupinfo.Num != 0 {
			//indexData.Groups = append(indexData.Groups, "<option value='" + groupinfo.groupName + "'>" + groupinfo.groupName + " 残り" + strconv.Itoa(groupinfo.Num) + "回" + "</option>")
			indexData.Groups = append(indexData.Groups, "<option value='" + groupinfo.groupName + "'>" + groupinfo.groupName  + "</option>")
		//}
	}

	//index.htmlを表示させる
	tmpl := template.Must(template.ParseFiles("./web/html/finalindex.html"))
	err := tmpl.Execute(w, indexData)
	if err != nil {
		log.Println("<Debug> can't open ./web/html/finalindex.html : ", err)
	}
}

// ajax戻り値のJSON用構造体
type measureParam struct {
	Time string
	Msg  string
	IsNewRecord string
}

type QueData struct {
	uuid string
	groupName string
	url string
 	w *websocket.Conn
}
//待ち行列を管理するキュー
var que = list.New()

func measureHandler(ws *websocket.Conn) {
	//websocketが接続されると呼び出される，計測する

	defer ws.Close()

	//ログファイルを開く
	logfile := logfileOpen()
	defer logfile.Close()


	// jsからhtml-inputの入力を受信する
	msg := ""
	err := websocket.Message.Receive(ws, &msg)
	if err != nil {
		log.Println("<Debug> web socket-readinfo", err)
		return
	}

	//msgはuuid,groupName,url,passの形式で送られてくるので","で分割する
	reg := "[,]"
  	tmp := regexp.MustCompile(reg).Split(msg, -1)

	uuid := tmp[0]
	groupName := tmp[1]
	url := tmp[2]
	pass := tmp[3]

	log.Println("<Info> request URL: " + url + ", GroupName: " + groupName + ", id: " + uuid)
	fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05")+"<Info> request URL: "+url+", GroupName: "+groupName+", id: "+uuid)

	//password認証
	for _, groupinfo := range groupInfo {
		if groupinfo.groupName == groupName {
			if groupinfo.Pass != pass {
				//passwordが異なる
				log.Println("<Info> id: " + uuid + ", Password is incorrect")
				fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + uuid + ", Password is incorrect")
				websocket.Message.Send(ws, "missmatch")
				return
			}
			break
		}
	}

	//url確認
	//dos対策 ipアドレスが学内のnetworkアドレスかで判断する
	isntDosAttack := false
	reg = "[/.]"
  	splitUrl := regexp.MustCompile(reg).Split(url, -1)
	//========ip split=====
	if len(splitUrl) > 3 && splitUrl[2] == "10" && splitUrl[3] == "70" {
		isntDosAttack = true

	}
	isntDosAttack = true
	if !isntDosAttack {
		//urlが異なる
		log.Println("<Info> id: " + uuid + ", url is incorrect")
		fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + uuid + ", url is incorrect")
		websocket.Message.Send(ws, "urlError")
		return
	}

	//現在の待ち数を知らせる
	websocket.Message.Send(ws, "queNum," + strconv.Itoa(que.Len()))

	//キューに追加
	que.PushBack(QueData{uuid, groupName, url, ws})

	//キューが空のとき
	//空でないときは下で呼び出している
	if que.Front().Value.(QueData).uuid == uuid {
		err := websocket.Message.Send(ws, "yourturn")
		if err != nil {
			log.Println("<Debug> web socket-can't send yourturn", err)
			return
		}
	}

	//websocket通信はReceiveで処理がとまる，Receiveを受け取ると下に処理される
	err = websocket.Message.Receive(ws, &msg)
	if err != nil {
		log.Println("<Debug> web socket-cant't receive start message", err, uuid)
		return
	}

	if msg != "start" {
		return
	}

	//キューから取り出す
	quuid := que.Front().Value.(QueData).uuid
	qgroupName := que.Front().Value.(QueData).groupName
	qurl := que.Front().Value.(QueData).url
	qw := que.Front().Value.(QueData).w

	//他のクライアントに待ち行列を通知する
	queNum := 0
	for e := que.Front(); e != nil; e = e.Next(){
		websocket.Message.Send(e.Value.(QueData).w, "queNum," + strconv.Itoa(queNum))
		queNum++
	}

	//index.jsに返すデータ変数
	var ret measureParam
	ret.IsNewRecord = "0"

	//まだ計測回数があるかcheck
	readGroupInfo()
	var canMeasure = 0
	for _, groupinfo := range groupInfo {
		if groupinfo.groupName == groupName {
			canMeasure = groupinfo.Num
			break
		}
	}

	//0でもいいけど念の為
	if canMeasure <= 0 {
		//計測回数がない
		log.Println("<Info> id: " + uuid + ", measure num is 0")
		fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + uuid + ", measure num is 0")
		websocket.Message.Send(qw, "measureNumError")
		queTidy(que)
		return
	}else{
		//残りの計測回数を送信
		websocket.Message.Send(qw, "groupNeasureNum," + strconv.Itoa(canMeasure - 1))
	}

	//abコマンド1度のみ計測を行い正常に計測できているか確認，できてる場合は計測回数を減らす
	abTestMsg := abtest.AbTest(logfile, quuid, qurl)
	//正常に計測できなかった
	if(abTestMsg != ""){

		err = websocket.Message.Send(qw, "measureResult," + "0.00" + "," + abTestMsg + "," + ret.IsNewRecord)
	
		log.Println("<Info> id: " + uuid + ", AbTest Error", err)
		fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + uuid + ", AbTestError")

		queTidy(que)
		return
	}

	//abコマンドで負荷をかける．計測時間を返す
	ret.Msg, ret.Time = ab.Ab(ws, logfile, quuid, qurl, canMeasure)

	//abコマンド中にソケットエラー発生
	if ret.Msg == "socketErr" {
		queTidy(que)
		return
	}

	//計測回数を減らす
	readGroupInfo()
	writeGroupInfo(qgroupName)

	//これまでの最高値を取り出す
	if ret.Msg == "" {
		ret.IsNewRecord, ret.Msg = score.Score(logfile, quuid, ret.Time, qgroupName)
	}

	err = websocket.Message.Send(qw, "measureResult," + ret.Time + "," + ret.Msg + "," + ret.IsNewRecord)
	if err != nil {
		log.Println("<Debug> web socket-cant't send measureResult", err)
		queTidy(que)
		return
	}

	//キューの整理
	queTidy(que)

	/*
	//正常に計測終了したら記録する
	if ret.Msg == "" {
		record.Record(logfile, quuid, ret.Time, qgroupName)
		ret.Msg = "計測完了"

		//計測回数を減らす
		writeGroupInfo(groupName)
	}
	*/

	//time.Sleep(20 * time.Second)

	
}

func queTidy(q *list.List){
	//キューから削除
	q.Remove(q.Front())

	//ページ更新や閉じるで消したリクエストを削除
	for e := q.Front(); e != nil; {
		//log.Println(e.Value.(QueData).uuid)
		err := websocket.Message.Send(e.Value.(QueData).w, "existCheck")
		if err != nil {
			tmp := e
			e = e.Next()
			q.Remove(tmp)
		}else{
			e = e.Next()
		}
	}

	//queの先頭に対して，計測開始指示する
	if q.Len() >= 1 {
		err := websocket.Message.Send(q.Front().Value.(QueData).w, "yourturn")
		if err != nil {
			log.Println("<Debug> web socket-cant't send yourturn", err)
			return
		}
	}
	
}

//records.jsonに記録する
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

	//records.jsonやdata.jsonに書き込む
	record.Record(logfile, id, times, groupName)

	//awsにアップロード
	uploadaws.UploadToAWS(logfile, id, "data/data.json", "data.json")
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
