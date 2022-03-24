package record

import(
  "os"
  "encoding/csv"
  "io"
  "strconv"
  "log"
  "os/exec"
  "fmt"
  "time"
)

func Record(logfile *os.File, id string, times string, groupName string) {

  recordData := "" //書き込みデータ
  doUpdate := false //記録が更新したかどうか

  //data.csvに記録する
  //data.csvを読み込む
  csvFile, err := os.Open("../public/score.csv")
  if err != nil {
    log.Println("<Debug> can't open ../public/score.csv : ", err)
  }
  reader := csv.NewReader(csvFile)

  //groupNameの一致を探し，数値を比較する
  for {
    line, err := reader.Read()
    if err == io.EOF {
        break
    }
    //同時に書き込みデータを作成する
    recordData += line[0] + ","
    //グループ名を探し，計測時間を比較
    if line[0] == groupName {
      nowData, _ := strconv.ParseFloat(times, 64)
      highData, _ := strconv.ParseFloat(line[1], 64)
      if nowData > highData {
        recordData += times + "\n"
        doUpdate = true
      }else{
        recordData += line[1] + "\n"
      }
    }else{
      recordData += line[1] + "\n"
    }
  }
  csvFile.Close()

  //ファイル書き込み
  file, err := os.Create("../public/score.csv")
  if err != nil{
    log.Println("<Debug> can't open or create ../public/score.csv : ", err)
  }
  defer file.Close()
  _, err = file.WriteString(recordData)
  if err != nil {
    log.Println("<Debug> cant' write ../public/score.csv : ", err)
  }

  //csvファイルをgithubにpush
  if doUpdate {
    csvPush(logfile, id, groupName)
  }
}

func csvPush(logfile *os.File, id string, groupName string){
  //git add ../exp1_ranking/public/score.csv
  err := exec.Command("git", "add", "../public/score.csv").Run()
  if err != nil {
    log.Println("<Debug> can't execute git add ../public/score.csv : ", err)
  }
  err = exec.Command("git", "commit", "-m", groupName + "の記録更新").Run()
  if err != nil {
    log.Println("<Debug> can't execute git commit -m \"grouphogeの記録更新\" : ", err)
  }
  err = exec.Command("git", "push").Run()
  if err != nil {
    log.Println("<Debug> can't execute git push : ", err)
  }

  log.Println("<Info> id: " + id + ",git push new record")
  fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + id + ",git push new record")


}
