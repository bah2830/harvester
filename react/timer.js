import React from 'react';

export class Timer extends React.Component {
    constructor(props) {
        super(props);

        this.stopTimer = this.stopTimer.bind(this);
        this.startTimer = this.startTimer.bind(this);
        this.openLink = this.openLink.bind(this);
    }

    stopTimer() {
        astilectron.sendMessage("stop|" + this.props.timer.key);
    }

    startTimer() {
        astilectron.sendMessage("start|" + this.props.timer.key);
    }

    openLink() {
        astilectron.sendMessage("open|" + this.props.timer.key);
    }

    render() {
        const timer = this.props.timer;

        let iconSrc = '/img/icons/jira.png';
        if (timer.harvest != undefined) {
            iconSrc = '/img/icons/harvest.png';
        }
        const icon = <img src={iconSrc} height="20px" />;


        let description = "";
        if (timer.jira != undefined) {
            description = timer.jira.fields.summary;
        } else if (timer.harvest != undefined) {
            description = timer.harvest.project.name;
        }

        const playImg = '/img/icons/play.png';
        const stopImg = '/img/icons/stop.png';
        const button = (
            <button
                type="button"
                onClick={timer.running ? this.stopTimer : this.startTimer}
                className="btn btn-dark btn-sm timer-btn"
            >
                {timer.running ? <img src={stopImg} height="20px" /> : <img src={playImg} height="20px" />}
                {timer.running && timer.runtime}
            </button >
        )

        return (
            <div className="d-flex flex-row align-middle task-timer align-items-center">
                <div className="p-1">{icon}</div>
                <div className="col text-truncate">
                    <a href="#" onClick={this.openLink} className="jira-link">{timer.key}: {description}</a>
                </div>
                <div className="p-1">
                    {button}
                </div >
            </div>
        );
    }
}