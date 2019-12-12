import React from 'react';
import { Timer } from './timer';

export class Timers extends React.Component {
    render() {
        const rows = [];
        appData.data.timers.forEach(function (timer, i) {
            rows.push(<Timer key={i} timer={timer} />);
        });

        return (
            <div id="main-content">
                {rows}
            </div>
        );
    }
}