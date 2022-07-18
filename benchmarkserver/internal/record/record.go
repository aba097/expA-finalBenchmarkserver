package record

import(
  "os"
  "encoding/json"
  "strconv"
  "io/ioutil"
  "log"
  "fmt"
  "time"
)

type Recorddata struct {
	Date    string  `json:"date"`
	Time	string	`json:"time"`
	GroupID string  `json:"group_id"`
	Score   float64 `json:"score"`
}

type Log struct {
	Records []Recorddata `json:"records"`
}

type Colors_st struct {
	Colors []string `json:"colors"`
}

type Groups_st struct {
	Groups []string	`json:"groups"`
}

func Record(logfile *os.File, id string, times string, groupName string) {

  //records.jsonに記録
  writeLogsjson(logfile, id, times, groupName, "data/records.json")

  //records.json colors.json groups.jsonからdata.jsonを作成する
  transformRechartsJson(logfile, id, "data/colors.json", "data/groups.json", "data/records.json", "data/data.json")
    
}


//records.jsonに書き込む
func writeLogsjson(logfile *os.File, id string, times string, groupName string, filePath string){
  date := time.Now().Format("1/2")
	hour:= time.Now().Format("15:04:15")

  nowData, _ := strconv.ParseFloat(times, 64)

  	// Json読取部
	jsonFromFile, err := os.ReadFile(filePath)
	if err != nil { 
    log.Println("<Debug>(record) can't open " + filePath + " : ", err)
  }
	var jsonData Log
	err = json.Unmarshal(jsonFromFile, &jsonData)
	if err != nil { 
    log.Println("<Debug>(record) can't decode " + filePath + " : ", err)
  }

  //追加
  jsonData.Records = append(jsonData.Records, Recorddata{date, hour, groupName, nowData})

	// Json出力部
	jsonStr, err := json.Marshal(jsonData)
	if err != nil { 
    log.Println("<Debug>(record) can't encode " + filePath + " : ", err)
   }
	fp, err := os.Create(filePath)
	if err != nil { 
    log.Println("<Debug>(record) can't open " + filePath + " : ", err)
   }
	defer fp.Close()
	err = ioutil.WriteFile(filePath, jsonStr, 0666)
	if err != nil {
    log.Println("<Debug>(record) can't write " + filePath + " : ", err)
   }

   log.Println("<Info> id: " + id + ", data record records.json")
   fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + id + ", data record records.json")
 
}


//records.json colors.json groups.jsonからdata.jsonを作成する
func transformRechartsJson(logfile *os.File, id string, colorsPath string, groupsPath string, logsPath string, rechartsPath string) {
	colorsFromFile, err := os.ReadFile(colorsPath)
	if err != nil { 
    log.Println("<Debug>can't open " + colorsPath + " : ", err)
  }
  var colorsData Colors_st
	err = json.Unmarshal(colorsFromFile, &colorsData)
	if err != nil {
    log.Println("<Debug>(record) can't encode " + colorsPath + " : ", err)

  }

	groupsFromFile, err := os.ReadFile(groupsPath)
	if err != nil { 
    log.Println("<Debug>can't open " + groupsPath + " : ", err)
  }
	var groupsData Groups_st
	err = json.Unmarshal(groupsFromFile, &groupsData)
	if err != nil { 
    log.Println("<Debug>(record) can't encode " + groupsPath + " : ", err)
  }

	logsFromFile, err := os.ReadFile(logsPath)
	if err != nil { 
    log.Println("<Debug>can't open " + logsPath + " : ", err)
  }
	var logsData Log
	err = json.Unmarshal(logsFromFile, &logsData)
	if err != nil {
    log.Println("<Debug>(record) can't encode " + logsPath + " : ", err)
  }

	max_scores := make(map[string]interface{})
	for _, v := range groupsData.Groups {
		max_scores[v] = 0.0
	}

	var recharts_list []interface{};
	for i, _ := range logsData.Records {
		// はじめではなく，前回のログと日付が異なる
		if i != 0 && logsData.Records[i].Date != logsData.Records[i - 1].Date {
			day_record := map[string]interface{}{"name": logsData.Records[i - 1].Date}
			record := merge(max_scores, day_record) 
			recharts_list = append(recharts_list, record)
		}
		// 最後で，記録更新している
		if i == len(logsData.Records) - 1 {
			max_scores[logsData.Records[i].GroupID] = logsData.Records[i].Score;
			day_record := map[string]interface{}{"name": logsData.Records[i].Date}
			record := merge(max_scores, day_record) 
			recharts_list = append(recharts_list, record)
		}
		if max_scores[logsData.Records[i].GroupID].(float64) < logsData.Records[i].Score {
			max_scores[logsData.Records[i].GroupID] = logsData.Records[i].Score
		}
	}

	trans_colors := map[string]interface{}{"colors": colorsData.Colors}
	trans_groups := map[string]interface{}{"groups": groupsData.Groups}
	trans_recharts := map[string]interface{}{"recharts": recharts_list}

	trans := merge(merge(trans_colors, trans_groups), trans_recharts)

	jsonStr, err := json.Marshal(trans, )
	if err != nil { 
    log.Println("<Debug>(record) can't decode trans : ", err)
  }
	fp, err := os.Create(rechartsPath)
	if err != nil {
    log.Println("<Debug>can't open " + rechartsPath + " : ", err)
  }
	defer fp.Close()
	err = ioutil.WriteFile(rechartsPath, jsonStr, 0666)
	if err != nil {
    log.Println("<Debug>(record) can't write " + rechartsPath + " : ", err)
  }

  log.Println("<Info> id: " + id + ", data record data.json")
  fmt.Fprintln(logfile, time.Now().Format("2006/01/02 15:04:05") + "<Info> id: " + id + ", data record data.json")

}

func merge(m1 map[string]interface{}, m2 map[string]interface{}) map[string]interface{} {
    ans := make(map[string]interface{})
    for k, v := range m1 {
        ans[k] = v
    }
    for k, v := range m2 {
        ans[k] = v
    }
    return (ans)
}


