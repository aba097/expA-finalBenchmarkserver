package ab

import (
  "io/ioutil"
  "time"
  "strings"
  "log"
  "os/exec"
  "regexp"
  "strconv"
  "fmt"
  "os"
  "sort"
  "golang.org/x/net/websocket"
)

//検索時間がどんなものかをチェックする関数
func Ab(ws *websocket.Conn, logfile *os.File, id string, url string, canMeasure int) (string, string) {
  
  var measureTimes []float64
  //複数タグで検索し，計測(test)

  file, _ := ioutil.ReadFile("./data/finalTag" + strconv.Itoa(canMeasure) + ".csv")
  tags := strings.Split(string(file), "\n")

   //最終の空白行対応
   if tags[len(tags) - 1] == "" {
    tags = tags[0:len(tags) - 1]
  }

  for i := 0; i < 50; i++ {
    socketErr := websocket.Message.Send(ws, "measureNum," + strconv.Itoa(i + 1))
		if socketErr != nil {
      return "socketErr", "0.00"
    }

    tag := tags[i]
    fmt.Println(i, tag)
    //log.Println("<Info> id: " + id + ", selected tag: " + s)
    //-c -nを変更する
    //out, err := exec.Command("ab", "-c", "1", "-n", "1", "-t", "2", url + "?tag=" + tag).Output()
    out, err := exec.Command("../hey-master/hey", "-c", "5", "-n", "10", "-t", "10", url + "?tag=" + tag).Output()
   
    if err != nil {
      log.Println(fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 5 -n 10 -t 10 " + url + "?tag=" + tag + ")" , err))
      fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 5 -n 10 -t 10" + url + "?tag=" + tag + ")" , err))
      return "エラー", "0.00"
    }

    execRes := string(out)
    //abコマンドの結果を:と改行で分割する
    reg := "[:\n]"
    splitExecRes := regexp.MustCompile(reg).Split(execRes, -1)
    //分割したものからRequests per secondを探す
    //次にあるのが計測値なので，j+1して指定，空白で分割し，数値のみ取り出す
    //例：Requests/sec:	2.3470
    for j, ss := range splitExecRes {
      if strings.Contains(ss, "Requests/sec") {
        sss := strings.Split(splitExecRes[j + 1], "\t")
        //float64に変換して加算
        measureTime, _ := strconv.ParseFloat(sss[len(sss) - 1], 64)
        //tag, timeを表示
        //fmt.Printf("%s,%.2f\n",tag, measureTime)
        measureTimes = append(measureTimes, measureTime)
      }
    }

    //heyに表示されるerrorチェック
    reg = "\n"
    splitExecRes = regexp.MustCompile(reg).Split(execRes, -1)
    for j, ss := range splitExecRes {
      if ss == "Error distribution:" {
        errMsg := ""
        for k := j + 1; k < len(splitExecRes); k += 1 {
          errMsg += splitExecRes[k] + "<br>"
        }
        log.Println("<Error> id: " + id + ", " + errMsg)
        return strconv.Itoa(i) + "/50タグ" + "エラー" + errMsg, "0.00"
      }
    }
    
  
    //curlでhtmlを取得し，imgタグ内の.staticflickr.comの数が100個あるか数える
    //htmlが正常か簡易的にチェック
    if !Checkhtml(logfile, id, url, tag) {
      return tag + "タグのHTMLファイルの取得失敗", "0.00"
    }

  }
//*/
  //ネットワークの関係で遅くなったタグを下位10件を削除
  sort.Slice(measureTimes, func(i, j int) bool {
    return measureTimes[i] > measureTimes[j]
  })
  var measureTime float64 = 0
  for i := 0; i < 40; i++ {
    measureTime += measureTimes[i]
  }
  //文字列にして返す measureTime / タグ数に変更する
  log.Println(fmt.Sprintf("<Info> id: " + id + " complete hey, measureTime = ", measureTime))
  fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Info> id: " + id + " complete hey, measureTime = ", measureTime))
  return "", strconv.FormatFloat(measureTime, 'f', 2, 64)
}

//htmlファイルが簡易的に正常かどうか確認する
func Checkhtml(logfile *os.File, id string, url string, tag string) bool {
  //.staticflickr.comという文字列が何個あるか確認する
  //.staticflickr.comは，Flickrサーバ上の画像URL	http://farm5.staticflickr.com/40～略～m.jpgで使われている

  count := 0

  //curlでhtmlを取得する
  out, err := exec.Command("curl", url + "?tag=" + tag).Output()

  if err != nil {
    log.Println(fmt.Sprintf("<Error> id: " + id + " execCmd(curl " + url + "?tag=" + tag + ")" , err))
    fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Error> id: " + id + " execCmd(curl " + url + "?tag=" + tag + ")" , err))
    return false
  }

  html := string(out)

  //"<"でファイルを分割する
  reg := "[<]"
  splitHtml := regexp.MustCompile(reg).Split(html, -1)
  //分割したものから .static.flickr.comが含まれているか確認する
  for _, s := range splitHtml {
    if strings.Contains(s, ".static.flickr.com") {
      count++
    }
  }

  //.static.flickr.comが100個あった場合，正常そう
  if(count == 100){
    //log.Println(fmt.Sprintf("<Info> id: " + id + ", htmlchek Success: .static.flickr.com num: ", count))
    //fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Info> id: " + id + ", htmlchek Success: .static.flickr.com num: ", count))
    return true
  }else{
    log.Println(fmt.Sprintf("<Info> id: " + id + ", htmlchek Failure: .static.flickr.com num: ", count))
    fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Info> id: " + id + ", " + tag + "tag htmlchek Failure: .static.flickr.com num: ", count))
    return false
  }
}
