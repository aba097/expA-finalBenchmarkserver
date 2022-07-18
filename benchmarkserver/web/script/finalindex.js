var groupName = "";
var time = "";
var uuid = "";
var socket = null;
var isConnection = false
var didRecord = false

//HTMLが読み込まれたとき
$(document).ready(function () {

	//uuidの取得
	uuid = UUID.randomUUID()
	console.log(uuid)

	//接続
	//socket = new WebSocket("ws://10.70.174.61:3000/ws");
	socket = new WebSocket("ws://localhost:3000/ws");
	//ソケットが確立した時に呼び出される
	socket.onopen = function () {
		console.log("system", "connect to server")
	};

	// 発生しうるエラーを待ち受けする
	socket.addEventListener('error', function (event) {
		console.log('WebSocket error: ', event)
		alert("ベンチマークサーバとの接続エラーが発生しました")
		location.reload();
  	});

	//serverからのイベント受け取り
	socket.onmessage = function (event) {
		var revMsg = event.data.split(',');
		console.log(revMsg)

		// passwordが異なる
		if(revMsg[0] == "missmatch"){
			console.log("missmatch")
			alert("パスードが異なります")
			location.reload();
		}

		// urlが異なる
		if (revMsg[0] == "urlError") {
			console.log("urlError")
			alert("学外のURLが指定されています。10.70からはじまるIPアドレスを指定してください")
			location.reload();
		}

		//計測回数が0回
		if (revMsg[0] == "measureNumError") {
			console.log("measureNumError")
			alert("計測回数の上限を超えています")
			location.reload();
		}

		// サーバーからメッセージを受け取る
		if(revMsg[0] == "yourturn"){
			console.log("measureStart")
			socket.send("start")
		}

		//キューの待ち数
		if(revMsg[0] == "queNum"){
			console.log("queNum")

			if(revMsg[1] == "0"){
				$('#queNum').text("計測中")
			}else{
				$('#queNum').html("<span class=\"queue_num\">" + revMsg[1] + "</span> 組 計測中")
			}
			
		}

		//何タグ目を計測しているか
		if (revMsg[0] == "measureNum"){
			$('#queNum').text("計測中" + revMsg[1] + "/50")
		}

		//何回目の計測か
		if (revMsg[0] == "groupNeasureNum"){
			$('#groupNeasureNum').text( Number(10 - revMsg[1]) + "回目の計測（残りの計測可能回数：" + revMsg[1] + "回）")
		}

		//計測結果
		if(revMsg[0] == "measureResult"){
			console.log("measureResult")
			
			var revTime = revMsg[1];
			var revMsgg = revMsg[2];
			var revIsNewRecord = revMsg[3];

			//記録するボタンを表示する
			if (revIsNewRecord == "1") {
				//recordBtnActionで使用
				groupName = $('[name="groupName"] option:selected').val();
				time = revTime;

				//buttonタグ作成
				$('#measureResult').append('<button class="startBtn" id="recordBtn">計測結果を記録する</button>')
			}

			$('#MeasureTime').text('Requests per second：' + revTime)
			$('#Msg').html(revMsgg)

			//画面表示
			$('#startedMeasure').toggle();
			$('#measureResult').toggle();

		}
	};

	//計測開始ボタンクリックアクション
	$('#startMeasureBtn').on('click', function(){
		console.log("<startMeasureAction> uuid: " + uuid + "url: " + $('input[name="url"]').val() + ", group: " + $('[name="groupName"] option:selected').val() + ", pass: " + $('input[name="password"]').val())

		//入力フォームを非表示にし，計測中を表示
		$('#topPage').toggle();
		$('#startedMeasure').toggle();

		//ベンチマークサーバに入力を送る
		socket.send(uuid + "," + $('[name="groupName"] option:selected').val() + "," + $('input[name="url"]').val() + "," + $('input[name="password"]').val());

	});

	//記録ボタンクリックアクション
	$(document).on('click', '#recordBtn', function () {

		console.log("recordBtnAction")
		
		//一度記録した後に、ページを戻ることでもう一度記録できるのを防ぐ
		if(didRecord){
			alert("すでに記録されています")
			location.reload();
		}else{
			//ajax urlとgroupNameとuuidを/recordに送る
			$.ajax({
				type: "POST",
				//送信先URL
				url: "record",
				data: {
					//送信データ
					"groupName": groupName,
					"time": time,
					"id": uuid,
				},
				//受け取り成功時
				success: function () {
					console.log("recordResult")
					if (confirm('記録しました\nランキングページを確認する場合はOK\nトップページに戻る場合はキャンセル\nをクリックしてください')) {
						didRecord = true
						window.location.href = 'http://expa-ranking.s3-website-ap-northeast-1.amazonaws.com/?group_id=' + groupName;
					} else {
						location.reload();
					}
				}
			});
		}
	});

	//結果画面にあるトップへボタンを押したとき
	$('#restartBtn').on('click', function(){
		location.reload();
	});

	//トップページにある更新画像を押した時
	$('#reloadImg').on('click', function(){
		location.reload();
	});

});

//キューから退出するボタンを押した時
$(document).on('click', '#breakMeasureBtn', function(){
	console.log("breakMeasureBtnAction")
	//ページ更新することで，beforeunloadが呼ばれ，pageout()が実行される
	location.reload();
});

//gen uuid
class UUID {

	static #uuidIte = ( function* () {
  
		const HEXOCTETS = Object.freeze( [ ...Array( 0x100 ) ].map( ( e, i ) => i.toString( 0x10 ).padStart( 2, "0" ).toUpperCase() ) );
	  	const VARSION = 0x40;
	  	const VARIANT = 0x80;
	  	const bytes = new Uint8Array( 16 );
	  	const rand = new Uint32Array( bytes.buffer );
  
	  	for (;;) {
  
			for ( let i = 0; i < rand.length; i++ ) {
		  		rand[ i ] = Math.random() * 0x100000000 >>> 0;
			}
  
			yield "" +
			HEXOCTETS[ bytes[ 0 ] ] +
			HEXOCTETS[ bytes[ 1 ] ] +
			HEXOCTETS[ bytes[ 2 ] ] +
			HEXOCTETS[ bytes[ 3 ] ] + "-" +
			HEXOCTETS[ bytes[ 4 ] ] +
			HEXOCTETS[ bytes[ 5 ] ] + "-" +
			HEXOCTETS[ bytes[ 6 ] & 0x0f | VARSION ] +
			HEXOCTETS[ bytes[ 7 ] ] + "-" +
			HEXOCTETS[ bytes[ 8 ] & 0x3f | VARIANT ] +
			HEXOCTETS[ bytes[ 9 ] ] + "-" +
			HEXOCTETS[ bytes[ 10 ] ] +
			HEXOCTETS[ bytes[ 11 ] ] +
			HEXOCTETS[ bytes[ 12 ] ] +
			HEXOCTETS[ bytes[ 13 ] ] +
			HEXOCTETS[ bytes[ 14 ] ] +
			HEXOCTETS[ bytes[ 15 ] ];
		}
	} )();
  
	static randomUUID() {
	  	return this.#uuidIte.next().value;
	}
}