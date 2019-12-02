class TimeSheet extends React.Component {
    render() {
        return (
            <div>
                hi
            </div>
        );
    }
}

if (typeof timesheet !== 'undefined') {
    const render = () =>  ReactDOM.render(<TimeSheet />, document.getElementById('app'));
    timesheet.render = render;
}