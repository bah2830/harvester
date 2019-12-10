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

    harvest() {
        astilectron.sendMessage("harvest");
    }

    render() {
        return (
            <nav className="navbar navbar-expand-lg fixed-top navbar-dark d-flex flex-row">
                <div className="p-2"><img onClick={this.harvest} src="/img/icons/harvest.png" height="20px" /></div>
                <div className="p-2"><img onClick={this.refresh} src="/img/icons/refresh.png" height="20px" /></div>
                <div className="col">&nbsp;</div>
                <div className="p-2">
                    <img
                        onClick={this.timesheet}
                        src={appData.data.view === 'timesheet' ? "/img/icons/close.png" : "/img/icons/info.png"}
                        height="20px"
                    />
                </div>
                <div className="p-2">
                    <img
                        onClick={this.settings}
                        src={appData.data.view === 'settings' ? "/img/icons/close.png" : "/img/icons/settings.png"}
                        height="20px"
                    />
                </div>
            </nav>
        );
    }
}
