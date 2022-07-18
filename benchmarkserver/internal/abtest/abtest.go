package abtest

import (
  "io/ioutil"
  "time"
  "strings"
  "log"
  "os/exec"
  "regexp"
  "fmt"
  "math/rand"
  "os"
)



//検索時間がどんなものかをチェックする関数
func AbTest(logfile *os.File, id string, url string) string {
  //func Ab(logfile *os.File, id string, url string) (string, string) {

  	//tagsをシャッフルする
  	file, _ := ioutil.ReadFile("./data/searchtag.txt")
  	tags := strings.Split(string(file), "\n")
   	//最終の空白行対応
   	if tags[len(tags) - 1] == "" {
    	tags = tags[0:len(tags) - 1]
  	}
  	rand.Seed(time.Now().UnixNano())
  	rand.Shuffle(len(tags), func(i, j int) { tags[i], tags[j] = tags[j], tags[i] })


	tag := tags[0]
	//log.Println("<Info> id: " + id + ", selected tag: " + s)
	//-c -nを変更する
	//out, err := exec.Command("ab", "-c", "1", "-n", "1", "-t", "2", url + "?tag=" + tag).Output()
	out, err := exec.Command("../hey-master/hey", "-c", "5", "-n", "10", "-t", "5", url + "?tag=" + tag).Output()
	if err != nil {
		log.Println(fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 5 -n 10 -t 5 " + url + "?tag=" + tag + ")" , err))
		fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + fmt.Sprintf("<Error> id: " + id + " execCmd(./hey -c 5 -n 10 -t 5" + url + "?tag=" + tag + ")" , err))
		return "エラー"
	}

	//heyの実行結果をバイナリから文字列に変換
	execRes := string(out)
	//heyに表示されるerrorチェック
	reg := "\n"
	splitExecRes := regexp.MustCompile(reg).Split(execRes, -1)
	for j, ss := range splitExecRes {
		if ss == "Error distribution:" {
		errMsg := ""
		for k := j + 1; k < len(splitExecRes); k += 1 {
			errMsg += splitExecRes[k] + "<br>"
		}
		log.Println("<Error> id: " + id + ", " + errMsg)
		return "エラー" + errMsg
		}
	}

	//curlでhtmlを取得し，imgタグ内の.staticflickr.comの数が100個あるか数える
	//htmlが正常か簡易的にチェック
	if !Checkhtml(logfile, id, url, tag) {
		return tag + "タグのHTMLファイルの取得失敗"
	}

	//正常であれば””を返す
  	return ""
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
