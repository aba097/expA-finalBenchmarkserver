package score

import(
  "os"
  "encoding/csv"
  "io"
  "strconv"
  "log"
  "fmt"
  "time"
)

func Score(logfile *os.File, id string, times string, groupName string) (bool, string) {
  msg := "" //返すメッセージ
  isnewrecord := false

  //score.csvを読み込む
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
    //グループ名を探し，計測時間を比較
    if line[0] == groupName {
      nowData, _ := strconv.ParseFloat(times, 64)
      highData, _ := strconv.ParseFloat(line[1], 64)
      if nowData > highData {
        msg = "記録更新（これまでの最高値：" + line[1] + "）"
        isnewrecord = true
      }else{
        msg = "記録更新ならず，現在の最高値：" + line[1]
      }
      break
    }
  }
  csvFile.Close()

  log.Println("<Info> id: " + id + ", record msg: " + msg)
  fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + id + ", record msg: " + msg)

  return isnewrecord, msg

}