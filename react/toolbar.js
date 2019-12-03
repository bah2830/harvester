import React from 'react';

export class Toolbar extends React.Component {
    refresh() {
        astilectron.sendMessage("refresh");
    }

    settings() {
        astilectron.sendMessage("settings");
    }

    timesheet() {
        astilectron.sendMessage("timesheet");
    }

    render() {
        return (
            <nav className="navbar navbar-expand-lg fixed-top navbar-dark d-flex flex-row">
                <div className="p-2"><img id="copy" src="/img/icons/copy.png" height="20px" /></div>
                <div className="p-2"><img id="refresh" onClick={this.refresh} src="/img/icons/refresh.png" height="20px" /></div>
                <div className="col">&nbsp;</div>
                <div className="p-2"><img id="timesheet" onClick={this.timesheet} src="/img/icons/info.png" height="20px" /></div>
                <div className="p-2"><img id="settings" onClick={this.settings} src="/img/icons/settings.png" height="20px" /></div>
            </nav>
        );
    }
}
