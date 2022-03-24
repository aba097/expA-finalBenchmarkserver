import React from 'react'
import './components.css';

export default class Score extends React.Component {
    render() {
        return (
            <div class='score'>
                <div class='rank'>
                    {this.props.rank + 1}位
                </div>
                <div class='id'>
                    {this.props.score[0]}
                </div>
                <div class='value'>
                    {this.props.score[1]}
                </div>
            </div>
        )
    }
}
