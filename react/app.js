import React from 'react';
import ReactDOM from 'react-dom';
import { Error } from './error';
import { Toolbar } from './toolbar';
import { Timers } from './timers';
import { TimeSheet } from './timesheet';
import { Settings } from './settings';

class App extends React.Component {
    render() {
        return (
            <div>
                <Toolbar />}
                {appData.data.error && <Error />}
                {appData.data.timers && <Timers />}
                {appData.data.view === 'timesheet' && <TimeSheet />}
                {appData.data.view === 'settings' && <Settings />}
            </div>
        );
    }
}

const render = () => ReactDOM.render(<App />, document.getElementById('app'));
appData.render = render;
render();
