import React from 'react';

export class TimeSheet extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            firstRender: true,
            activeView: 'day',
            currentViewContent: <></>
        };

        this.activateTab = this.activateTab.bind(this);
        this.day = this.day.bind(this);
        this.week = this.week.bind(this);
    }

    activateTab(tab) {
        astilectron.sendMessage(tab, function (response) {
            if (response === undefined) {
                return;
            }

            this.setState({
                firstRender: false,
                activeView: tab,
                currentViewContent: this[tab](response),
            });
        }.bind(this));
    }

    day(timesheet) {
        return (
            <table className="table time-table ">
                <tbody>
                    <tr>
                        <td>Jira</td>
                        <td align="right">Hours</td>
                    </tr>
                    {timesheet.tasks.map((jira, i) => {
                        return (
                            <tr key={i}>
                                <td>{jira.key}</td>
                                <td align="right">{jira.totalTime}</td>
                            </tr>
                        );
                    })}
                    <tr><td colSpan="2">&nbsp;</td></tr>
                    <tr>
                        <td>Total</td>
                        <td align="right">{timesheet.total}</td>
                    </tr>
                </tbody>
            </table>
        );
    }

    week(timesheet) {
        return (
            <table className="table time-table ">
                <tbody>
                    <tr>
                        <td>Jira</td>
                        <td align="right">Mon</td>
                        <td align="right">Tue</td>
                        <td align="right">Wed</td>
                        <td align="right">Thu</td>
                        <td align="right">Fri</td>
                        <td align="right">Sat</td>
                        <td align="right">Sun</td>
                        <td align="right">Total</td>
                    </tr>
                    {timesheet.tasks.map((jira, i) => {
                        return (
                            <tr key={i}>
                                <td>{jira.key}</td>
                                {jira.durations.map((duration, j) => {
                                    return <td key={j} align="right">{duration}</td>
                                })}
                                <td align="right">{jira.totalTime}</td>
                            </tr>
                        );
                    })}
                    <tr><td colSpan="9">&nbsp;</td></tr>
                    <tr>
                        <td>Total</td>
                        {timesheet.daysTotal.map((t, i) => {
                            return <td key={i} align="right">{t}</td>;
                        })}
                        <td align="right">{timesheet.total}</td>
                    </tr>
                </tbody>
            </table>
        );
    }

    render() {
        const tabs = [
            'day',
            'week',
        ];

        return (
            <div className="container-fluid">
                <div className="row">
                    <div className="p-2">
                        <div className="btn-group btn-group-sm" role="group">
                            {tabs.map((tab, i) => {
                                let colorClass = this.state.activeView === tab ? 'btn-secondary' : 'btn-dark';
                                return (
                                    <button
                                        type="button"
                                        key={i}
                                        onClick={() => this.activateTab(tab)}
                                        className={"btn btn-sm " + colorClass}
                                    >
                                        {tab}
                                    </button>
                                );
                            })}
                        </div>
                    </div>
                    <div className="col">&nbsp;</div>
                    <div className="p-2">
                        <img id="refresh" onClick={() => this.activateTab(this.state.activeView)} src="/img/icons/refresh.png" height="20px" />
                    </div>
                </div>

                {this.state.firstRender && this.activateTab('day')}

                {this.state.currentViewContent}
            </div>
        );
    }
}
