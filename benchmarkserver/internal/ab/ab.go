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
)

//検索時間がどんなものかをチェックする関数
func Ab(logfile *os.File, id string, url string) (string, string) {
  var measureTimes float64 //計測時間の合計
  measureTimes = 0
  //*

  //複数タグで検索し，計測(test)
  file, _ := ioutil.ReadFile("./data/searchtag.txt")

  tags := strings.Split(string(file), "\n")
  for i, tag := range tags{
    if i >= len(tags) - 1{
      break
    }

    //log.Println("<Info> id: " + id + ", selected tag: " + s)
    //-c -nを変更する
    //out, err := exec.Command("ab", "-c", "1", "-n", "1", "-t", "2", url + "?tag=" + tag).Output()
    out, err := exec.Command("../hey-master/hey", "-c", "1", "-n", "1", "-t", "100", url + "?tag=" + tag).Output()
    if err != nil {
      log.Println(fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 1 -n 10 -t 100 " + url + "?tag=" + tag + ")" , err))
      fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 1 -n 10 -t 100" + url + "?tag=" + tag + ")" , err))
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
        //log.Println("<Info> id: " + id + ", Requests per second: " + sss[len(sss) - 1])
        //float64に変換して加算
        measureTime, _ := strconv.ParseFloat(sss[len(sss) - 1], 64)
        fmt.Printf("%s,%.2f\n",tag, measureTime)
        measureTimes += measureTime
      }
      if ss == "Error distribution"{
        log.Println("<Error> id: " + id + ", " + splitExecRes[j + 2] + ": " + splitExecRes[j + 3])

        return "エラー " + splitExecRes[j + 2] + ": " + splitExecRes[j + 3], "0.00"
        
      }
    }

    //curlでhtmlを取得し，imgタグ内の.staticflickr.comの数が100個あるか数える
    //htmlが正常か簡易的にチェック
    if !Checkhtml(logfile, id, url, tag) {
      return "HTMLファイルが改ざんされている可能性があります", "0.00"
    }
    

  }


  //文字列にして返す measureTime / タグ数に変更する
  return "", strconv.FormatFloat(measureTimes, 'f', 2, 64)
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
    log.Println(fmt.Sprintf("<Info> id: " + id + ", htmlchek Success: .static.flickr.com num: ", count))
    fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Info> id: " + id + ", htmlchek Success: .static.flickr.com num: ", count))
    return true
  }else{
    log.Println(fmt.Sprintf("<Info> id: " + id + ", htmlchek Failure: .static.flickr.com num: ", count))
    fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Info> id: " + id + ", htmlchek Failure: .static.flickr.com num: ", count))
    return false
  }
}
