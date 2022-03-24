var groupName = "";
var time = "";
var id = "";

//HTMLが読み込まれたとき
$(document).ready(function(){
  //計測開始ボタンクリックアクション
	$('#startMeasureBtn').on('click', function(){

		console.log("startMeasureAction")
    //inputタグの入力
    console.log($('input[name="url"]').val())
    //selectタグの入力
    console.log($('[name="groupName"] option:selected').val())

    //入力フォームを非表示にし，計測中を表示
    $('#topPage').toggle();
    $('#startedMeasure').toggle();

    //ajax urlとgroupNameを/measureに送る
		$.ajax({
			type: "POST",
      //送信先URL
			url: "measure",
			data: {
        //送信データ
				"url": $('input[name="url"]').val(),
				"groupName": $('[name="groupName"] option:selected').val(),
			},
      //サーバから受け取るデータ(data)の形式
			dataType: "json",
      //受け取り成功時
			success: function(measureResultData){
				//計測中を非表示にし，計測結果を表示する
				console.log("measureResult")
				console.log(measureResultData.Time)
				console.log(measureResultData.Msg)
				console.log(measureResultData.IsNewRecord)
				console.log(measureResultData.Id)

				if(measureResultData.IsNewRecord){
					//recordBtnActionで使用
					groupName = $('[name="groupName"] option:selected').val();
					time = measureResultData.Time;
					id = measureResultData.Id;

					//buttonタグ作成
					$('#measureResult').append('<button class="startBtn" id="recordBtn">計測結果を記録する</button>')
				}


				//計測結果を表示する
				$('#MeasureTime').text('Requests per second：' + measureResultData.Time)
				$('#Msg').text(measureResultData.Msg)


				setTimeout(function(){
					//画面表示
					$('#startedMeasure').toggle();
					$('#measureResult').toggle();
				}, 3000);
			}
		});
	});

	//記録ボタンクリックアクション

	$(document).on('click', '#recordBtn', function(){

		console.log("recordBtnAction")

		//ajax urlとgroupNameを/measureに送る
		$.ajax({
			type: "POST",
      //送信先URL
			url: "record",
			data: {
        //送信データ
				"groupName": groupName,
				"time": time,
				"id": id,
			},
      //受け取り成功時
			success: function(){
				console.log("recordResult")
				alert('記録しました')
				location.reload();
			}
		});
	});

	//結果画面にあるトップへボタンを押したとき
	//非表示・表示を切り替える
	$('#restartBtn').on('click', function(){
		$('#measureResult').toggle();
		$('#topPage').toggle();
	});
});
