import React from 'react';
import Moment from 'react-moment';

export class TimeSheet extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            firstRender: true,
            activeView: 'day',
            currentTimesheet: "",
        };

        this.activateTab = this.activateTab.bind(this);
        this.sendToBackend = this.sendToBackend.bind(this);
        this.dateBack = this.dateBack.bind(this);
        this.dateForward = this.dateForward.bind(this);
        this.day = this.day.bind(this);
        this.week = this.week.bind(this);
        this.copy = this.copy.bind(this);
    }

    sendToBackend(tab, message) {
        if (message) {
            message = tab + message;
        } else {
            message = tab;
        }

        astilectron.sendMessage(message, function (response) {
            if (response === undefined) {
                return;
            }

            this.setState({
                firstRender: false,
                activeView: tab,
                currentTimesheet: response
            });
        }.bind(this));
    }

    activateTab(tab) {
        this.sendToBackend(tab, '|' + this.state.currentTimesheet.timeStart + '|=');
    }

    dateBack(tab) {
        this.sendToBackend(tab, '|' + this.state.currentTimesheet.timeStart + '|-');
    }

    dateForward(tab) {
        this.sendToBackend(tab, '|' + this.state.currentTimesheet.timeStart + '|+');
    }

    copy(tab) {
        this.sendToBackend(tab, '|' + this.state.currentTimesheet.timeStart + '|copy');
    }

    datePicker(tab) {
        if (!this.state.currentTimesheet.timeStart) {
            return <div>{tab}</div>;
        }

        const timeStart = this.state.currentTimesheet.timeStart;
        const timeEnd = this.state.currentTimesheet.timeEnd;

        let content = <Moment format="MMM Do" date={timeStart} />;
        if (tab === 'week') {
            content = <div><Moment format="MMM Do" date ={timeStart} /> - <Moment format="MMM Do" date={timeEnd} /></div>;
        }

        return (
            <div className="btn-group btn-group-sm btn-datepicker" role="group">
                <button type="button" className="btn btn-sm btn-dark" onClick={() => this.dateBack(tab)}>
                    <img src="/img/icons/left.png" height="20px" />
                </button>
                <button type="button" className="btn btn-sm btn-secondary" onClick={() => this.activateTab(tab)}>{content}</button>
                <button type="button" className="btn btn-sm btn-dark" onClick={() => this.dateForward(tab)}>
                    <img src="/img/icons/right.png" height="20px" />
                </button>
            </div>
        );
    }

    day() {
        const timesheet = this.state.currentTimesheet
        if (!timesheet) {
            return <></>;
        }

        return (
            <table className="table time-table">
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

    week() {
        const timesheet = this.state.currentTimesheet
        if (!timesheet) {
            return <></>;
        }

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
                    <div className="p-2">{this.datePicker(this.state.activeView)}</div>
                    <div className="col">&nbsp;</div>
                    <div className="p-2">
                        <img
                            onClick={() => this.copy(this.state.activeView)}
                            src="/img/icons/copy.png"
                            height="20px"
                            alt="Copy all projects without a harvest project for the month to the clipboard"
                        />
                    </div>
                </div>

                {this.state.firstRender && this.activateTab('day')}
                {this[this.state.activeView](this.state.currentTimesheet)}
            </div>
        );
    }
}
