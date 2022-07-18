package score

import(
  "os"
  "encoding/json"
  "strconv"
  "log"
  "fmt"
  "time"
)


type Record struct {
	Date    string  `json:"date"`
	Time	string	`json:"time"`
	GroupID string  `json:"group_id"`
	Score   float64 `json:"score"`
}

type Log struct {
	Records []Record `json:"records"`
}

//json 最高記録を取得する
func get_max(groupName string, filePath string) float64 {
	logsFromFile, err := os.ReadFile(filePath)
	if err != nil { 
    log.Println("<Debug>(score) can't open " + filePath + " : ", err)
  }
	var logsData Log
	err = json.Unmarshal(logsFromFile, &logsData)
	if err != nil { 
    log.Println("<Debug>(score) can't decode " + filePath + " : ", err)
  }

	max_score := 0.0
	for _, v := range logsData.Records {
		if v.GroupID == groupName && max_score < v.Score {
			max_score = v.Score
		}
	}

	return max_score
}

//現在の最高値と比較する
func Score(logfile *os.File, id string, times string, groupName string) (string, string) {
  msg := "" //返すメッセージ
  isnewrecord := "0"

  highData := get_max(groupName, "data/records.json")

  nowData, _ := strconv.ParseFloat(times, 64)

  if nowData > highData {
    msg = "記録更新（これまでの最高値：" + strconv.FormatFloat(highData, 'f', 2, 64) + "）"
    isnewrecord = "1"
  }else{
    msg = "記録更新ならず，現在の最高値：" + strconv.FormatFloat(highData, 'f', 2, 64)
  }

  log.Println("<Info> id: " + id + ", record msg: " + msg)
  fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + id + ", record msg: " + msg)

  return isnewrecord, msg

}