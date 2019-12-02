class Timers extends React.Component {
    render() {
        var self = this;
        const rows = [];
        timers.data.tasks.forEach(function (timer, i) {
            rows.push(<Timer key={i} timer={timer} />);
        });

        return (
            <div id="main-content">
                {rows}
            </div>
        );
    }
}