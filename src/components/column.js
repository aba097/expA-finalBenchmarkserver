import React from 'react'
import './components.css';

export default class Column extends React.Component {
    render() {
        return (
            <div class='column'>
                <div class='rank'>
                    順位
                </div>
                <div class='id'>
                    グループ名
                </div>
                <div class='value'>
                    スコア
                </div>
            </div>
        )
    }
}