import React from 'react'
import Column from '../components/column';
import Score from '../components/score'
import './ranking.css';

function csvToArray(filename) {
    let srt = new XMLHttpRequest();
    srt.open("GET", filename, false);
    try {
        srt.send(null);
    } catch (err) {
        console.log(err)
    }
    let csletr = [];
    let lines = srt.responseText.split("\n");
    for (let i = 0; i < lines.length; ++i) {
        let cells = lines[i].split(",");
        if (cells.length !== 1) {
            csletr.push(cells);
        }
    }
    return csletr;
}

export default class TopPage extends React.Component{
    render() {
        const arr = csvToArray("./score.csv");
        arr.sort(function(a,b){return(b[1] - a[1]);});
        return (
            <div class="ranking">
                <h1>実験A  Webシステム高速化ランキング</h1>
                <Column />
                { arr.map((value, index) => 
                    <Score score={value} rank={index} /> 
                )}
            </div>
        )
    }
}
